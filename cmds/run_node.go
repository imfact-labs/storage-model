package cmds

import (
	"context"

	apic "github.com/imfact-labs/currency-model/api"
	ccmds "github.com/imfact-labs/currency-model/app/cmds"
	cdigest "github.com/imfact-labs/currency-model/digest"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/logging"
	"github.com/imfact-labs/mitum2/util/ps"
	"github.com/imfact-labs/storage-model/api"
	"github.com/imfact-labs/storage-model/digest"
	"github.com/pkg/errors"
)

type RunCommand struct {
	ccmds.RunCommand
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

	cmd.RunCommand.SetLog(log.Log())

	if len(cmd.HTTPState) > 0 {
		if err := cmd.RunCommand.RunHTTPState(cmd.HTTPState); err != nil {
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

	pps := ccmds.DefaultRunPS()

	_ = pps.AddOK(cdigest.PNameDigester, digest.ProcessDigester, nil, cdigest.PNameDigesterDataBase).
		AddOK(cdigest.PNameStartDigester, cdigest.ProcessStartDigester, nil, apic.PNameStartAPI)
	_ = pps.POK(launch.PNameStorage).PostAddOK(ps.Name("check-hold"), cmd.RunCommand.PCheckHold)
	_ = pps.POK(launch.PNameStates).
		PreAddOK(PNameOperationProcessorsMap, POperationProcessorsMap).
		PreAddOK(ps.Name("when-new-block-saved-in-consensus-state-func"), cmd.RunCommand.PWhenNewBlockSavedInConsensusStateFunc).
		PreAddOK(ps.Name("when-new-block-saved-in-syncing-state-func"), cmd.RunCommand.PWhenNewBlockSavedInSyncingStateFunc).
		PreAddOK(ps.Name("when-new-block-confirmed-func"), cmd.RunCommand.PWhenNewBlockConfirmed)
	_ = pps.POK(launch.PNameEncoder).
		PostAddOK(launch.PNameAddHinters, PAddHinters)
	_ = pps.POK(apic.PNameAPI).
		PostAddOK(ccmds.PNameDigestAPIHandlers, cmd.pDigestAPIHandlers)
	_ = pps.POK(cdigest.PNameDigester).
		PostAddOK(ccmds.PNameDigesterFollowUp, cdigest.PdigesterFollowUp)

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

	return cmd.RunCommand.RunNode(nctx)
}

func (cmd *RunCommand) pDigestAPIHandlers(ctx context.Context) (context.Context, error) {
	var params *launch.LocalParams
	var local base.LocalNode
	var design cdigest.YamlDigestDesign

	if err := util.LoadFromContextOK(ctx,
		launch.LocalContextKey, &local,
		launch.LocalParamsContextKey, &params,
		cdigest.ContextValueDigestDesign, &design,
	); err != nil {
		return nil, err
	}

	if design.Equal(cdigest.YamlDigestDesign{}) {
		return ctx, nil
	}

	cache, err := ccmds.LoadCache(cmd.RunCommand.Log(), ctx, design)
	if err != nil {
		return ctx, err
	}

	var dnt *apic.HTTP2Server
	if err := util.LoadFromContext(ctx, apic.ContextValueDigestNetwork, &dnt); err != nil {
		return ctx, err
	}

	router := dnt.Router()

	handlers, err := ccmds.SetDigestAPIDefaultHandlers(cmd.RunCommand.Log(), ctx, params, cache, router, dnt.Queue())
	if err != nil {
		return ctx, err
	}

	if err := handlers.Initialize(); err != nil {
		return ctx, err
	}

	handlers.SetEncoders(encs)
	handlers.SetEncoder(enc)

	apic.SetHandlers(handlers, design.Digest)
	api.SetHandlers(handlers)

	dnt.SetEncoder(encs)

	return ctx, nil
}
