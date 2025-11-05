package cmds

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	currencycmds "github.com/ProtoconNet/mitum-currency/v3/cmds"
	currencydigest "github.com/ProtoconNet/mitum-currency/v3/digest"
	"github.com/ProtoconNet/mitum-storage/digest"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/isaac"
	isaacstates "github.com/ProtoconNet/mitum2/isaac/states"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/network/quicmemberlist"
	"github.com/ProtoconNet/mitum2/network/quicstream"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/logging"
	"github.com/ProtoconNet/mitum2/util/ps"
	"github.com/arl/statsviz"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type RunCommand struct { //nolint:govet //...
	//revive:disable:line-length-limit
	launch.DesignFlag
	launch.DevFlags `embed:"" prefix:"dev."`
	launch.PrivatekeyFlags
	Discovery []launch.ConnInfoFlag `help:"member discovery" placeholder:"ConnInfo"`
	Hold      launch.HeightFlag     `help:"hold consensus states"`
	HTTPState string                `name:"http-state" help:"runtime statistics thru https" placeholder:"bind address"`
	launch.ACLFlags
	exitf  func(error)
	log    *zerolog.Logger
	holded bool
	//revive:enable:line-length-limit
}

func (cmd *RunCommand) Run(pctx context.Context) error {
	var log *logging.Logging
	if err := util.LoadFromContextOK(pctx, launch.LoggingContextKey, &log); err != nil {
		return err
	}

	log.Log().Debug().
		Interface("design", cmd.DesignFlag).
		Interface("privatekey", cmd.PrivatekeyFlags).
		Interface("discovery", cmd.Discovery).
		Interface("hold", cmd.Hold).
		Interface("http_state", cmd.HTTPState).
		Interface("dev", cmd.DevFlags).
		Interface("acl", cmd.ACLFlags).
		Msg("flags")

	cmd.log = log.Log()

	if len(cmd.HTTPState) > 0 {
		if err := cmd.runHTTPState(cmd.HTTPState); err != nil {
			return errors.Wrap(err, "failed to run http state")
		}
	}

	nctx := util.ContextWithValues(pctx, map[util.ContextKey]interface{}{
		launch.DesignFlagContextKey:    cmd.DesignFlag,
		launch.DevFlagsContextKey:      cmd.DevFlags,
		launch.DiscoveryFlagContextKey: cmd.Discovery,
		launch.PrivatekeyContextKey:    string(cmd.PrivatekeyFlags.Flag.Body()),
		launch.ACLFlagsContextKey:      cmd.ACLFlags,
	})

	pps := currencycmds.DefaultRunPS()

	_ = pps.AddOK(currencycmds.PNameDigester, ProcessDigester, nil, currencycmds.PNameDigesterDataBase).
		AddOK(currencycmds.PNameStartDigester, ProcessStartDigester, nil, currencycmds.PNameStartAPI)
	_ = pps.POK(launch.PNameStorage).PostAddOK(ps.Name("check-hold"), cmd.pCheckHold)
	_ = pps.POK(launch.PNameStates).
		PreAddOK(PNameOperationProcessorsMap, POperationProcessorsMap).
		PreAddOK(ps.Name("when-new-block-saved-in-consensus-state-func"), cmd.pWhenNewBlockSavedInConsensusStateFunc).
		PreAddOK(ps.Name("when-new-block-saved-in-syncing-state-func"), cmd.pWhenNewBlockSavedInSyncingStateFunc).
		PreAddOK(ps.Name("when-new-block-confirmed-func"), cmd.pWhenNewBlockConfirmed)
	_ = pps.POK(launch.PNameEncoder).
		PostAddOK(launch.PNameAddHinters, PAddHinters)
	_ = pps.POK(currencycmds.PNameAPI).
		PostAddOK(currencycmds.PNameDigestAPIHandlers, cmd.pDigestAPIHandlers)
	_ = pps.POK(currencycmds.PNameDigester).
		PostAddOK(currencycmds.PNameDigesterFollowUp, PDigesterFollowUp)

	_ = pps.SetLogging(log)

	log.Log().Debug().Interface("process", pps.Verbose()).Msg("process ready")

	nctx, err := pps.Run(nctx) //revive:disable-line:modifies-parameter
	defer func() {
		log.Log().Debug().Interface("process", pps.Verbose()).Msg("process will be closed")

		if _, err = pps.Close(pctx); err != nil {
			log.Log().Error().Err(err).Msg("failed to close")
		}
	}()

	if err != nil {
		return err
	}

	log.Log().Debug().
		Interface("discovery", cmd.Discovery).
		Interface("hold", cmd.Hold.Height()).
		Msg("node started")

	return cmd.run(nctx)
}

var errHoldStop = util.NewIDError("hold stop")

func (cmd *RunCommand) run(pctx context.Context) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	exitch := make(chan error)

	cmd.exitf = func(err error) {
		exitch <- err
	}

	stopstates := func() {}

	if !cmd.holded {
		deferred, err := cmd.runStates(ctx, pctx)
		if err != nil {
			return err
		}

		stopstates = deferred
	}

	select {
	case <-ctx.Done(): // NOTE graceful stop
		return errors.WithStack(ctx.Err())
	case err := <-exitch:
		if errors.Is(err, errHoldStop) {
			stopstates()

			<-ctx.Done()

			return errors.WithStack(ctx.Err())
		}

		return err
	}
}

func (cmd *RunCommand) runStates(ctx, pctx context.Context) (func(), error) {
	var discoveries *util.Locked[[]quicstream.ConnInfo]
	var states *isaacstates.States

	if err := util.LoadFromContextOK(pctx,
		launch.DiscoveryContextKey, &discoveries,
		launch.StatesContextKey, &states,
	); err != nil {
		return nil, err
	}

	if dis := launch.GetDiscoveriesFromLocked(discoveries); len(dis) < 1 {
		cmd.log.Warn().Msg("empty discoveries; will wait to be joined by remote nodes")
	}

	go func() {
		cmd.exitf(<-states.Wait(ctx))
	}()

	return func() {
		if err := states.Hold(); err != nil && !errors.Is(err, util.ErrDaemonAlreadyStopped) {
			cmd.log.Error().Err(err).Msg("failed to stop states")

			return
		}

		cmd.log.Debug().Msg("states stopped")
	}, nil
}

func (cmd *RunCommand) pWhenNewBlockSavedInSyncingStateFunc(pctx context.Context) (context.Context, error) {
	var log *logging.Logging
	var db isaac.Database
	var design currencydigest.YamlDigestDesign

	if err := util.LoadFromContextOK(pctx,
		launch.LoggingContextKey, &log,
		launch.CenterDatabaseContextKey, &db,
		currencydigest.ContextValueDigestDesign, &design,
	); err != nil {
		return pctx, err
	}

	var f func(height base.Height)
	if !design.Equal(currencydigest.YamlDigestDesign{}) && design.Digest {
		var di *digest.Digester
		if err := util.LoadFromContextOK(pctx,
			currencydigest.ContextValueDigester, &di,
		); err != nil {
			return pctx, err
		}

		g := cmd.whenBlockSaved(db, di)

		f = func(height base.Height) {
			g(pctx)
			l := log.Log().With().Interface("height", height).Logger()

			if cmd.Hold.IsSet() && height == cmd.Hold.Height() {
				l.Debug().Msg("will be stopped by hold")
				cmd.exitf(errHoldStop.WithStack())

				return
			}
		}
	} else {
		f = func(height base.Height) {
			l := log.Log().With().Interface("height", height).Logger()

			if cmd.Hold.IsSet() && height == cmd.Hold.Height() {
				l.Debug().Msg("will be stopped by hold")
				cmd.exitf(errHoldStop.WithStack())

				return
			}
		}
	}

	return context.WithValue(pctx,
		launch.WhenNewBlockSavedInSyncingStateFuncContextKey, f,
	), nil
}

func (cmd *RunCommand) pWhenNewBlockSavedInConsensusStateFunc(pctx context.Context) (context.Context, error) {
	var log *logging.Logging

	if err := util.LoadFromContextOK(pctx,
		launch.LoggingContextKey, &log,
	); err != nil {
		return pctx, err
	}

	f := func(bm base.BlockMap) {
		l := log.Log().With().
			Interface("blockmap", bm).
			Interface("height", bm.Manifest().Height()).
			Logger()

		if cmd.Hold.IsSet() && bm.Manifest().Height() == cmd.Hold.Height() {
			l.Debug().Msg("will be stopped by hold")

			cmd.exitf(errHoldStop.WithStack())

			return
		}
	}

	return context.WithValue(pctx, launch.WhenNewBlockSavedInConsensusStateFuncContextKey, f), nil
}

func (cmd *RunCommand) pWhenNewBlockConfirmed(pctx context.Context) (context.Context, error) {
	var log *logging.Logging
	var db isaac.Database
	var design currencydigest.YamlDigestDesign

	if err := util.LoadFromContextOK(pctx,
		launch.LoggingContextKey, &log,
		launch.CenterDatabaseContextKey, &db,
		currencydigest.ContextValueDigestDesign, &design,
	); err != nil {
		return pctx, err
	}

	var f func(height base.Height)
	if !design.Equal(currencydigest.YamlDigestDesign{}) && design.Digest {
		f = func(height base.Height) {
			l := log.Log().With().Interface("height", height).Logger()

			err := digestFollowup(pctx, height)
			if err != nil {
				cmd.exitf(err)

				return
			}

			if cmd.Hold.IsSet() && height == cmd.Hold.Height() {
				l.Debug().Msg("will be stopped by hold")
				cmd.exitf(errHoldStop.WithStack())

				return
			}
		}
	} else {
		f = func(height base.Height) {
			l := log.Log().With().Interface("height", height).Logger()

			if cmd.Hold.IsSet() && height == cmd.Hold.Height() {
				l.Debug().Msg("will be stopped by hold")
				cmd.exitf(errHoldStop.WithStack())

				return
			}
		}
	}

	return context.WithValue(pctx,
		launch.WhenNewBlockConfirmedFuncContextKey, f,
	), nil
}

func (cmd *RunCommand) whenBlockSaved(
	db isaac.Database,
	di *digest.Digester,
) ps.Func {
	return func(ctx context.Context) (context.Context, error) {
		switch m, found, err := db.LastBlockMap(); {
		case err != nil:
			return ctx, err
		case !found:
			return ctx, errors.Errorf("last BlockMap not found")
		default:
			if di != nil {
				go func() {
					di.Digest([]base.BlockMap{m})
				}()
			}
		}
		return ctx, nil
	}
}

func (cmd *RunCommand) pCheckHold(pctx context.Context) (context.Context, error) {
	var db isaac.Database
	if err := util.LoadFromContextOK(pctx, launch.CenterDatabaseContextKey, &db); err != nil {
		return pctx, err
	}

	switch {
	case !cmd.Hold.IsSet():
	case cmd.Hold.Height() < base.GenesisHeight:
		cmd.holded = true
	default:
		switch m, found, err := db.LastBlockMap(); {
		case err != nil:
			return pctx, err
		case !found:
		case cmd.Hold.Height() <= m.Manifest().Height():
			cmd.holded = true
		}
	}

	return pctx, nil
}

func (cmd *RunCommand) runHTTPState(bind string) error {
	addr, err := net.ResolveTCPAddr("tcp", bind)
	if err != nil {
		return errors.Wrap(err, "failed to parse --http-state")
	}

	m := http.NewServeMux()
	if err := statsviz.Register(m); err != nil {
		return errors.Wrap(err, "failed to register statsviz for http-state")
	}

	cmd.log.Debug().Stringer("bind", addr).Msg("statsviz started")

	go func() {
		_ = http.ListenAndServe(addr.String(), m)
	}()

	return nil
}

func (cmd *RunCommand) pDigestAPIHandlers(ctx context.Context) (context.Context, error) {
	var params *launch.LocalParams
	var local base.LocalNode
	var design currencydigest.YamlDigestDesign

	if err := util.LoadFromContextOK(ctx,
		launch.LocalContextKey, &local,
		launch.LocalParamsContextKey, &params,
		currencydigest.ContextValueDigestDesign, &design,
	); err != nil {
		return nil, err
	}

	if design.Equal(currencydigest.YamlDigestDesign{}) {
		return ctx, nil
	}

	cache, err := cmd.loadCache(ctx, design)
	if err != nil {
		return ctx, err
	}

	var dnt *currencydigest.HTTP2Server
	if err := util.LoadFromContext(ctx, currencydigest.ContextValueDigestNetwork, &dnt); err != nil {
		return ctx, err
	}

	router := dnt.Router()

	defaultHandlers, err := cmd.setDigestAPIDefaultHandlers(ctx, params, cache, router, dnt.Queue())
	if err != nil {
		return ctx, err
	}

	if err := defaultHandlers.Initialize(design.Digest); err != nil {
		return ctx, err
	}

	handlers, err := cmd.setDigestHandlers(ctx, params, cache, router, defaultHandlers.Routes())
	if err != nil {
		return ctx, err
	}

	if err := handlers.Initialize(); err != nil {
		return ctx, err
	}

	dnt.SetEncoder(encs)

	return ctx, nil
}

func (cmd *RunCommand) loadCache(_ context.Context, design currencydigest.YamlDigestDesign) (currencydigest.Cache, error) {
	c, err := currencydigest.NewCacheFromURI(design.Cache().String())
	if err != nil {
		cmd.log.Err(err).Str("cache", design.Cache().String()).Msg("failed to connect cache server")
		cmd.log.Warn().Msg("instead of remote cache server, internal mem cache can be available, `memory://`")

		return nil, err
	}
	return c, nil
}

func (cmd *RunCommand) setDigestAPIDefaultHandlers(
	ctx context.Context,
	params *launch.LocalParams,
	cache currencydigest.Cache,
	router *mux.Router,
	queue chan currencydigest.RequestWrapper,
) (*currencydigest.Handlers, error) {
	var nodeDesign launch.NodeDesign
	var design currencydigest.YamlDigestDesign
	var st *currencydigest.Database
	if err := util.LoadFromContext(ctx,
		launch.DesignContextKey, &nodeDesign,
		currencydigest.ContextValueDigestDesign, &design,
	); err != nil {
		return nil, err
	}
	if design.Digest {
		if err := util.LoadFromContext(ctx, currencydigest.ContextValueDigestDatabase, &st); err != nil {
			return nil, err
		}
	}

	node, err := quicstream.NewConnInfoFromStringAddr(nodeDesign.Network.PublishString, nodeDesign.Network.TLSInsecure)
	if err != nil {
		return nil, err
	}

	handlers := currencydigest.NewHandlers(ctx, params.ISAAC.NetworkID(), encs, enc, st, cache, router, queue, node)

	h, err := cmd.setDigestAPINetworkClient(ctx, params, handlers)
	if err != nil {
		return nil, err
	}
	handlers = h

	return handlers, nil
}

func (cmd *RunCommand) setDigestHandlers(
	ctx context.Context,
	params *launch.LocalParams,
	cache currencydigest.Cache,
	router *mux.Router,
	routes map[string]*mux.Route,
) (*digest.Handlers, error) {
	var st *currencydigest.Database
	if err := util.LoadFromContext(ctx, currencydigest.ContextValueDigestDatabase, &st); err != nil {
		return nil, err
	}

	handlers := digest.NewHandlers(ctx, params.ISAAC.NetworkID(), encs, enc, st, cache, router, routes)

	return handlers, nil
}

func (cmd *RunCommand) setDigestAPINetworkClient(
	ctx context.Context,
	params *launch.LocalParams,
	handlers *currencydigest.Handlers,
) (*currencydigest.Handlers, error) {
	var memberList *quicmemberlist.Memberlist
	if err := util.LoadFromContextOK(ctx, launch.MemberlistContextKey, &memberList); err != nil {
		return nil, err
	}

	connectionPool, err := launch.NewConnectionPool(
		1<<9,
		params.ISAAC.NetworkID(),
		nil,
	)

	if err != nil {
		return nil, err
	}

	handlers = handlers.SetNetworkClientFunc(
		func() (*quicstream.ConnectionPool, *quicmemberlist.Memberlist, []quicstream.ConnInfo, error) { // nolint:contextcheck
			return connectionPool, memberList, []quicstream.ConnInfo{}, nil
		},
	)

	var design currencydigest.YamlDigestDesign
	if err := util.LoadFromContext(ctx, currencydigest.ContextValueDigestDesign, &design); err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return handlers, nil
		}

		return nil, err
	}

	if design.Equal(currencydigest.YamlDigestDesign{}) {
		return handlers, nil
	}

	handlers = handlers.SetNetworkClientFunc(
		func() (*quicstream.ConnectionPool, *quicmemberlist.Memberlist, []quicstream.ConnInfo, error) { // nolint:contextcheck
			return connectionPool, memberList, design.ConnInfo, nil
		},
	)

	cmd.log.Debug().Msg("send handler attached")

	return handlers, nil
}
