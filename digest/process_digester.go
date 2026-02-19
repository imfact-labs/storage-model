package digest

import (
	"context"

	cdigest "github.com/imfact-labs/currency-model/digest"
	"github.com/imfact-labs/mitum2/isaac"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/logging"
)

func ProcessDigester(ctx context.Context) (context.Context, error) {
	var vs util.Version
	var log *logging.Logging
	var digestDesign cdigest.YamlDigestDesign

	if err := util.LoadFromContextOK(ctx,
		launch.VersionContextKey, &vs,
		launch.LoggingContextKey, &log,
		cdigest.ContextValueDigestDesign, &digestDesign,
	); err != nil {
		return ctx, err
	}

	if !digestDesign.Digest {
		return ctx, nil
	}

	var st *cdigest.Database
	if err := util.LoadFromContext(ctx, cdigest.ContextValueDigestDatabase, &st); err != nil {
		return ctx, err
	}

	if st == nil {
		return ctx, nil
	}

	if err := st.CreateIndex(DefaultIndexes); err != nil {
		return ctx, err
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

	di := cdigest.NewDigester(st, root, sourceReaders, fromRemotes, design.NetworkID, vs.String(), nil)
	_ = di.SetLogging(log)

	di.PrepareFunc = []cdigest.BlockSessionPrepareFunc{
		cdigest.PrepareCurrencies, cdigest.PrepareAccounts, cdigest.PrepareDIDRegistry,
		PrepareStorage,
	}

	return context.WithValue(ctx, cdigest.ContextValueDigester, di), nil
}
