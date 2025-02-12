package storage

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/state"
	crtypes "github.com/ProtoconNet/mitum-currency/v3/types"
	statestr "github.com/ProtoconNet/mitum-storage/state"
	"github.com/ProtoconNet/mitum-storage/types"
	"sync"

	statecurrency "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
)

var deleteDataProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(DeleteDataProcessor)
	},
}

func (DeleteData) Process(
	_ context.Context, _ mitumbase.GetStateFunc,
) ([]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error) {
	return nil, nil, nil
}

type DeleteDataProcessor struct {
	*mitumbase.BaseOperationProcessor
}

func NewDeleteDataProcessor() crtypes.GetNewProcessor {
	return func(
		height mitumbase.Height,
		getStateFunc mitumbase.GetStateFunc,
		newPreProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
	) (mitumbase.OperationProcessor, error) {
		e := util.StringError("failed to create new DeleteDataProcessor")

		nopp := deleteDataProcessorPool.Get()
		opp, ok := nopp.(*DeleteDataProcessor)
		if !ok {
			return nil, e.Errorf("expected DeleteDataProcessor, not %T", nopp)
		}

		b, err := mitumbase.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e.Wrap(err)
		}

		opp.BaseOperationProcessor = b

		return opp, nil
	}
}

func (opp *DeleteDataProcessor) PreProcess(
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
) (context.Context, mitumbase.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(DeleteDataFact)
	if !ok {
		return ctx, mitumbase.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMTypeMismatch).
				Errorf("expected %T, not %T", DeleteDataFact{}, op.Fact())), nil
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", err)), nil
	}

	if err := state.CheckExistsState(statecurrency.DesignStateKey(fact.Currency()), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMCurrencyNF).Errorf("currency id %v", fact.Currency())), nil
	}

	if err := state.CheckExistsState(statestr.DesignStateKey(fact.Contract()), getStateFunc); err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMServiceNF).Errorf("storage service state for contract account %v",
				fact.Contract(),
			)), nil
	}

	if st, err := state.ExistsState(statestr.DataStateKey(fact.Contract(), fact.DataKey()), "storage data", getStateFunc); err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMStateNF).Errorf("storage data for key %v in contract account %v", fact.DataKey(),
				fact.Contract(),
			)), nil
	} else if d, err := statestr.GetDataFromState(st); err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMStateValInvalid).Errorf(
				"storage data for key %v in contract account %v", fact.DataKey(),
				fact.Contract(),
			)), nil
	} else if d.IsDeleted() {
		return nil, mitumbase.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMValueInvalid).Errorf(
				"storage data for key %v in contract account %v has already been deleted",
				fact.DataKey(), fact.Contract(),
			)), nil
	}

	if err := state.CheckExistsState(statestr.DataStateKey(fact.Contract(), fact.DataKey()), getStateFunc); err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMStateNF).Errorf(
				"storage data for key %v in contract account %v has already been deleted",
				fact.DataKey(), fact.Contract(),
			)), nil
	}

	return ctx, nil, nil
}

func (opp *DeleteDataProcessor) Process( // nolint:dupl
	_ context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc) (
	[]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error,
) {
	e := util.StringError("failed to process DeleteData")

	fact, ok := op.Fact().(DeleteDataFact)
	if !ok {
		return nil, nil, e.Errorf("expected DeleteDataFact, not %T", op.Fact())
	}

	stData := types.NewData(
		fact.DataKey(),
		"",
	)
	stData.SetDeleted()

	if err := stData.IsValid(nil); err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError("invalid storage data; %w", err), nil
	}

	var sts []mitumbase.StateMergeValue // nolint:prealloc
	sts = append(sts, state.NewStateMergeValue(
		statestr.DataStateKey(fact.Contract(), fact.DataKey()),
		statestr.NewDataStateValue(stData),
	))

	return sts, nil, nil
}

func (opp *DeleteDataProcessor) Close() error {
	deleteDataProcessorPool.Put(opp)

	return nil
}
