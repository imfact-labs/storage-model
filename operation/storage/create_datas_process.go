package storage

import (
	"context"
	"sync"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	currencystate "github.com/ProtoconNet/mitum-currency/v3/state"
	statecurrency "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-storage/state"
	"github.com/ProtoconNet/mitum-storage/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var createDatasItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CreateDatasItemProcessor)
	},
}

var createDatasProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CreateDatasProcessor)
	},
}

func (CreateDatas) Process(
	_ context.Context, _ base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	return nil, nil, nil
}

type CreateDatasItemProcessor struct {
	h      util.Hash
	sender base.Address
	item   CreateDatasItem
}

func (ipp *CreateDatasItemProcessor) PreProcess(
	_ context.Context, _ base.Operation, getStateFunc base.GetStateFunc,
) error {
	e := util.StringError("preprocess CreateDatasItemProcessor")
	it := ipp.item

	if err := it.IsValid(nil); err != nil {
		return e.Wrap(err)
	}

	if err := currencystate.CheckExistsState(statecurrency.DesignStateKey(it.Currency()), getStateFunc); err != nil {
		return e.Wrap(common.ErrCurrencyNF.Wrap(errors.Errorf("currency id %v", it.Currency())))
	}

	if err := currencystate.CheckExistsState(state.DesignStateKey(it.Contract()), getStateFunc); err != nil {
		return e.Wrap(
			common.ErrServiceNF.Errorf("storage service state for contract account %v", it.Contract()))
	}

	if found, _ := currencystate.CheckNotExistsState(state.DataStateKey(it.Contract(), it.DataKey()), getStateFunc); found {
		return e.Wrap(
			common.ErrStateE.Errorf(
				"storage data for key %q in contract account %v",
				it.DataKey(), it.Contract(),
			))
	}

	return nil
}

func (ipp *CreateDatasItemProcessor) Process(
	_ context.Context, _ base.Operation, getStateFunc base.GetStateFunc,
) ([]base.StateMergeValue, error) {
	it := ipp.item

	var sts []base.StateMergeValue
	data := types.NewData(
		it.DataKey(), it.DataValue(),
	)
	if err := data.IsValid(nil); err != nil {
		return nil, err
	}

	sts = append(sts, currencystate.NewStateMergeValue(
		state.DataStateKey(it.Contract(), it.DataKey()),
		state.NewDataStateValue(data),
	))

	return sts, nil
}

func (ipp *CreateDatasItemProcessor) Close() {
	ipp.h = nil
	ipp.sender = nil
	ipp.item = CreateDatasItem{}

	createDatasItemProcessorPool.Put(ipp)
}

type CreateDatasProcessor struct {
	*base.BaseOperationProcessor
}

func NewCreateDatasProcessor() currencytypes.GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new CreateDatasProcessor")

		nopp := createDatasProcessorPool.Get()
		opp, ok := nopp.(*CreateDatasProcessor)
		if !ok {
			return nil, e.Errorf("expected %T, not %T", CreateDatasProcessor{}, nopp)
		}

		b, err := base.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e.Wrap(err)
		}

		opp.BaseOperationProcessor = b

		return opp, nil
	}
}

func (opp *CreateDatasProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(CreateDatasFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMTypeMismatch).
				Errorf("expected %T, not %T", CreateDatasFact{}, op.Fact())), nil
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", err)), nil
	}

	for _, it := range fact.Items() {
		ip := createDatasItemProcessorPool.Get()
		ipc, ok := ip.(*CreateDatasItemProcessor)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError(
				common.ErrMTypeMismatch.Errorf("expected %T, not %T", CreateDatasItemProcessor{}, ip)), nil
		}

		ipc.h = op.Hash()
		ipc.sender = fact.Sender()
		ipc.item = it

		if err := ipc.PreProcess(ctx, op, getStateFunc); err != nil {
			return nil, base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.Errorf("%v", err),
			), nil
		}

		ipc.Close()
	}

	return ctx, nil, nil
}

func (opp *CreateDatasProcessor) Process( // nolint:dupl
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, _ := op.Fact().(CreateDatasFact)

	var sts []base.StateMergeValue // nolint:prealloc
	for _, it := range fact.Items() {
		ip := createDatasItemProcessorPool.Get()
		ipc, _ := ip.(*CreateDatasItemProcessor)

		ipc.h = op.Hash()
		ipc.sender = fact.Sender()
		ipc.item = it

		st, err := ipc.Process(ctx, op, getStateFunc)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("failed to process CreateDatasItem; %w", err), nil
		}

		sts = append(sts, st...)

		ipc.Close()
	}

	items := make([]DataItem, len(fact.Items()))
	for i := range fact.Items() {
		items[i] = fact.Items()[i]
	}

	return sts, nil, nil
}

func (opp *CreateDatasProcessor) Close() error {
	createDatasProcessorPool.Put(opp)

	return nil
}

func calculateDIDItemsFee(getStateFunc base.GetStateFunc, items []DataItem) (
	map[currencytypes.CurrencyID]base.State, map[currencytypes.CurrencyID][2]common.Big, error) {
	feeReceiveSts := map[currencytypes.CurrencyID]base.State{}
	required := map[currencytypes.CurrencyID][2]common.Big{}

	for _, item := range items {
		rq := [2]common.Big{common.ZeroBig, common.ZeroBig}

		if k, found := required[item.Currency()]; found {
			rq = k
		}

		policy, err := currencystate.ExistsCurrencyPolicy(item.Currency(), getStateFunc)
		if err != nil {
			return nil, nil, err
		}

		switch k, err := policy.Feeer().Fee(common.ZeroBig); {
		case err != nil:
			return nil, nil, err
		case !k.OverZero():
			required[item.Currency()] = [2]common.Big{rq[0], rq[1]}
		default:
			required[item.Currency()] = [2]common.Big{rq[0].Add(k), rq[1].Add(k)}
		}

		if policy.Feeer().Receiver() == nil {
			continue
		}

		if err := currencystate.CheckExistsState(statecurrency.AccountStateKey(policy.Feeer().Receiver()), getStateFunc); err != nil {
			return nil, nil, err
		} else if st, found, err := getStateFunc(statecurrency.BalanceStateKey(policy.Feeer().Receiver(), item.Currency())); err != nil {
			return nil, nil, err
		} else if !found {
			return nil, nil, errors.Errorf("feeer receiver account not found, %s", policy.Feeer().Receiver())
		} else {
			feeReceiveSts[item.Currency()] = st
		}

	}

	return feeReceiveSts, required, nil

}
