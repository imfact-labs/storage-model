package storage

import (
	"context"
	"sync"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/state"
	stateextension "github.com/ProtoconNet/mitum-currency/v3/state/extension"
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	statestorage "github.com/ProtoconNet/mitum-storage/state"
	"github.com/ProtoconNet/mitum-storage/types"

	statecurrency "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var registerModelProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(RegisterModelProcessor)
	},
}

type RegisterModelProcessor struct {
	*mitumbase.BaseOperationProcessor
}

func NewRegisterModelProcessor() currencytypes.GetNewProcessor {
	return func(
		height mitumbase.Height,
		getStateFunc mitumbase.GetStateFunc,
		newPreProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
	) (mitumbase.OperationProcessor, error) {
		e := util.StringError("failed to create new RegisterModelProcessor")

		nopp := registerModelProcessorPool.Get()
		opp, ok := nopp.(*RegisterModelProcessor)
		if !ok {
			return nil, errors.Errorf("expected RegisterModelProcessor, not %T", nopp)
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

func (opp *RegisterModelProcessor) PreProcess(
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
) (context.Context, mitumbase.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(RegisterModelFact)
	if !ok {
		return ctx, mitumbase.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMTypeMismatch).
				Errorf("expected %T, not %T", RegisterModelFact{}, op.Fact())), nil
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", err)), nil
	}

	if err := state.CheckExistsState(statecurrency.DesignStateKey(fact.Currency()), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMCurrencyNF).Errorf("currency id, %v", fact.Currency())), nil
	}

	if found, _ := state.CheckNotExistsState(statestorage.DesignStateKey(fact.Contract()), getStateFunc); found {
		return ctx, mitumbase.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMServiceE).Errorf("storage service for contract account %v",
				fact.Contract(),
			)), nil
	}

	return ctx, nil, nil
}

func (opp *RegisterModelProcessor) Process(
	_ context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc) (
	[]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error,
) {
	e := util.StringError("failed to process RegisterModel")

	fact, ok := op.Fact().(RegisterModelFact)
	if !ok {
		return nil, nil, e.Errorf("expected RegisterModelFact, not %T", op.Fact())
	}

	var sts []mitumbase.StateMergeValue

	design := types.NewDesign(fact.Project())
	if err := design.IsValid(nil); err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError("invalid storage design, %q; %w", fact.Contract(), err), nil
	}

	sts = append(sts, state.NewStateMergeValue(
		statestorage.DesignStateKey(fact.Contract()),
		statestorage.NewDesignStateValue(design),
	))

	st, err := state.ExistsState(stateextension.StateKeyContractAccount(fact.Contract()), "contract account", getStateFunc)
	if err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError("contract account not found, %q; %w", fact.Contract(), err), nil
	}

	ca, err := stateextension.StateContractAccountValue(st)
	if err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError("failed to get state value of contract account, %q; %w", fact.Contract(), err), nil
	}
	nca := ca.SetActive(true)

	sts = append(sts, state.NewStateMergeValue(
		stateextension.StateKeyContractAccount(fact.Contract()),
		stateextension.NewContractAccountStateValue(nca),
	))

	return sts, nil, nil
}

func (opp *RegisterModelProcessor) Close() error {
	registerModelProcessorPool.Put(opp)

	return nil
}
