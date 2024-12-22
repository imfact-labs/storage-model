package storage

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/state"
	crtypes "github.com/ProtoconNet/mitum-currency/v3/types"
	statetstr "github.com/ProtoconNet/mitum-storage/state"
	"github.com/ProtoconNet/mitum-storage/types"
	"sync"

	statecurrency "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
)

var updateDataProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(UpdateDataProcessor)
	},
}

func (UpdateData) Process(
	_ context.Context, _ mitumbase.GetStateFunc,
) ([]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error) {
	return nil, nil, nil
}

type UpdateDataProcessor struct {
	*mitumbase.BaseOperationProcessor
}

func NewUpdateDataProcessor() crtypes.GetNewProcessor {
	return func(
		height mitumbase.Height,
		getStateFunc mitumbase.GetStateFunc,
		newPreProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
	) (mitumbase.OperationProcessor, error) {
		e := util.StringError("failed to create new UpdateDataProcessor")

		nopp := updateDataProcessorPool.Get()
		opp, ok := nopp.(*UpdateDataProcessor)
		if !ok {
			return nil, e.Errorf("expected UpdateDataProcessor, not %T", nopp)
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

func (opp *UpdateDataProcessor) PreProcess(
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
) (context.Context, mitumbase.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(UpdateDataFact)
	if !ok {
		return ctx, mitumbase.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMTypeMismatch).
				Errorf("expected %T, not %T", UpdateDataFact{}, op.Fact())), nil
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

	if err := state.CheckExistsState(statetstr.DesignStateKey(fact.Contract()), getStateFunc); err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMServiceNF).Errorf("storage service for contract account %v",
				fact.Contract(),
			)), nil
	}

	if err := state.CheckExistsState(statetstr.DataStateKey(fact.Contract(), fact.DataKey()), getStateFunc); err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMStateNF).Errorf(
				"storage data for key %v in contract account %v",
				fact.DataKey(), fact.Contract(),
			)), nil
	}

	return ctx, nil, nil
}

func (opp *UpdateDataProcessor) Process( // nolint:dupl
	_ context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc) (
	[]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error,
) {
	e := util.StringError("failed to process UpdateData")

	fact, ok := op.Fact().(UpdateDataFact)
	if !ok {
		return nil, nil, e.Errorf("expected UpdateDataFact, not %T", op.Fact())
	}

	stData := types.NewData(
		fact.DataKey(),
		fact.DataValue(),
	)
	if err := stData.IsValid(nil); err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError("invalid storage data; %w", err), nil
	}

	var sts []mitumbase.StateMergeValue // nolint:prealloc
	sts = append(sts, state.NewStateMergeValue(
		statetstr.DataStateKey(fact.Contract(), fact.DataKey()),
		statetstr.NewDataStateValue(stData),
	))

	return sts, nil, nil
}

func (opp *UpdateDataProcessor) Close() error {
	updateDataProcessorPool.Put(opp)

	return nil
}
