package cmds

import (
	"context"

	currencydigest "github.com/ProtoconNet/mitum-currency/v3/digest"
	"github.com/ProtoconNet/mitum-storage/digest"
	"github.com/ProtoconNet/mitum2/isaac"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/logging"
)

func ProcessDigester(ctx context.Context) (context.Context, error) {
	var vs util.Version
	var log *logging.Logging
	var digestDesign currencydigest.YamlDigestDesign

	if err := util.LoadFromContextOK(ctx,
		launch.VersionContextKey, &vs,
		launch.LoggingContextKey, &log,
		currencydigest.ContextValueDigestDesign, &digestDesign,
	); err != nil {
		return ctx, err
	}

	if !digestDesign.Digest {
		return ctx, nil
	}

	var st *currencydigest.Database
	if err := util.LoadFromContext(ctx, currencydigest.ContextValueDigestDatabase, &st); err != nil {
		return ctx, err
	}

	if st == nil {
		return ctx, nil
	}

	var design launch.NodeDesign
	if err := util.LoadFromContext(ctx,
		launch.DesignContextKey, &design,
	); err != nil {
		return ctx, err
	}
	root := launch.LocalFSDataDirectory(design.Storage.Base)

	var newReaders func(context.Context, string, *isaac.BlockItemReadersArgs) (*isaac.BlockItemReaders, error)
	var fromRemotes isaac.RemotesBlockItemReadFunc

	if err := util.LoadFromContextOK(ctx,
		launch.NewBlockItemReadersFuncContextKey, &newReaders,
		launch.RemotesBlockItemReaderFuncContextKey, &fromRemotes,
	); err != nil {
		return ctx, err
	}

	var sourceReaders *isaac.BlockItemReaders

	switch i, err := newReaders(ctx, root, nil); {
	case err != nil:
		return ctx, err
	default:
		sourceReaders = i
	}

	di := currencydigest.NewDigester(st, root, sourceReaders, fromRemotes, design.NetworkID, vs.String(), nil)
	_ = di.SetLogging(log)

	di.PrepareFunc = []currencydigest.BlockSessionPrepareFunc{
		currencydigest.PrepareCurrencies, currencydigest.PrepareAccounts, currencydigest.PrepareDIDRegistry,
		digest.PrepareStorage,
	}

	return context.WithValue(ctx, currencydigest.ContextValueDigester, di), nil
}
