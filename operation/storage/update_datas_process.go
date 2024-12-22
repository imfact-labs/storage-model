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

var updateDatasItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(UpdateDatasItemProcessor)
	},
}

var updateDatasProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(UpdateDatasProcessor)
	},
}

func (UpdateDatas) Process(
	_ context.Context, _ base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	return nil, nil, nil
}

type UpdateDatasItemProcessor struct {
	h      util.Hash
	sender base.Address
	item   UpdateDatasItem
}

func (ipp *UpdateDatasItemProcessor) PreProcess(
	_ context.Context, _ base.Operation, getStateFunc base.GetStateFunc,
) error {
	e := util.StringError("preprocess UpdatDatasItemProcessor")
	it := ipp.item

	if err := it.IsValid(nil); err != nil {
		return e.Wrap(err)
	}

	if err := currencystate.CheckExistsState(statecurrency.DesignStateKey(it.Currency()), getStateFunc); err != nil {
		return e.Wrap(common.ErrCurrencyNF.Wrap(errors.Errorf("currency id %v", it.Currency())))
	}

	if err := currencystate.CheckExistsState(state.DesignStateKey(it.Contract()), getStateFunc); err != nil {
		return e.Wrap(
			common.ErrServiceNF.Errorf("storage service in contract account %v", it.Contract()))
	}

	if err := currencystate.CheckExistsState(state.DataStateKey(it.Contract(), it.DataKey()), getStateFunc); err != nil {
		return e.Wrap(
			common.ErrStateNF.Errorf(
				"storage data for key %q in contract account %v",
				it.DataKey(), it.Contract(),
			))
	}

	return nil
}

func (ipp *UpdateDatasItemProcessor) Process(
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

func (ipp *UpdateDatasItemProcessor) Close() {
	ipp.h = nil
	ipp.sender = nil
	ipp.item = UpdateDatasItem{}

	updateDatasItemProcessorPool.Put(ipp)
}

type UpdateDatasProcessor struct {
	*base.BaseOperationProcessor
}

func NewUpdateDatasProcessor() currencytypes.GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new UpdateDatasProcessor")

		nopp := updateDatasProcessorPool.Get()
		opp, ok := nopp.(*UpdateDatasProcessor)
		if !ok {
			return nil, e.Errorf("expected %T, not %T", UpdateDatasProcessor{}, nopp)
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

func (opp *UpdateDatasProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(UpdateDatasFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMTypeMismatch).
				Errorf("expected %T, not %T", UpdateDatasFact{}, op.Fact())), nil
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", err)), nil
	}

	for _, it := range fact.Items() {
		ip := updateDatasItemProcessorPool.Get()
		ipc, ok := ip.(*UpdateDatasItemProcessor)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError(
				common.ErrMTypeMismatch.Errorf("expected %T, not %T", UpdateDatasItemProcessor{}, ip)), nil
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

func (opp *UpdateDatasProcessor) Process( // nolint:dupl
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, _ := op.Fact().(UpdateDatasFact)

	var sts []base.StateMergeValue // nolint:prealloc
	for _, it := range fact.Items() {
		ip := updateDatasItemProcessorPool.Get()
		ipc, _ := ip.(*UpdateDatasItemProcessor)

		ipc.h = op.Hash()
		ipc.sender = fact.Sender()
		ipc.item = it

		st, err := ipc.Process(ctx, op, getStateFunc)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("failed to process UpdateDatasItem; %w", err), nil
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

func (opp *UpdateDatasProcessor) Close() error {
	updateDatasProcessorPool.Put(opp)

	return nil
}
