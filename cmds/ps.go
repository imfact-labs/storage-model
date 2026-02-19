package cmds

import (
	"context"

	ccmds "github.com/imfact-labs/currency-model/app/cmds"
	cprocessor "github.com/imfact-labs/currency-model/operation/processor"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/isaac"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/mitum2/util/ps"
	"github.com/imfact-labs/storage-model/operation/storage"
)

var PNameOperationProcessorsMap = ps.Name("mitum-storage-operation-processors-map")

func POperationProcessorsMap(pctx context.Context) (context.Context, error) {
	var isaacParams *isaac.Params
	var db isaac.Database
	var opr *cprocessor.OperationProcessor
	var set *hint.CompatibleSet[isaac.NewOperationProcessorInternalFunc]

	if err := util.LoadFromContextOK(pctx,
		launch.ISAACParamsContextKey, &isaacParams,
		launch.CenterDatabaseContextKey, &db,
		ccmds.OperationProcessorContextKey, &opr,
		launch.OperationProcessorsMapContextKey, &set,
	); err != nil {
		return pctx, err
	}

	err := opr.SetCheckDuplicationFunc(cprocessor.CheckDuplication)
	if err != nil {
		return pctx, err
	}
	err = opr.SetGetNewProcessorFunc(cprocessor.GetNewProcessor)
	if err != nil {
		return pctx, err
	}

	if err := opr.SetProcessor(
		storage.RegisterModelHint,
		storage.NewRegisterModelProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		storage.DeleteDataHint,
		storage.NewDeleteDataProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		storage.CreateDataHint,
		storage.NewCreateDataProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		storage.UpdateDataHint,
		storage.NewUpdateDataProcessor(),
	); err != nil {
		return pctx, err
	}

	_ = set.Add(storage.CreateDataHint, func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
		return opr.New(
			height,
			getStatef,
			nil,
			nil,
		)
	})

	_ = set.Add(storage.RegisterModelHint, func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
		return opr.New(
			height,
			getStatef,
			nil,
			nil,
		)
	})

	_ = set.Add(storage.UpdateDataHint, func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
		return opr.New(
			height,
			getStatef,
			nil,
			nil,
		)
	})

	_ = set.Add(storage.DeleteDataHint, func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
		return opr.New(
			height,
			getStatef,
			nil,
			nil,
		)
	})

	pctx = context.WithValue(pctx, ccmds.OperationProcessorContextKey, opr)
	pctx = context.WithValue(pctx, launch.OperationProcessorsMapContextKey, set) //revive:disable-line:modifies-parameter

	return pctx, nil
}
