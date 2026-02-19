package storage

import (
	"context"
	"sync"

	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/state"
	statec "github.com/imfact-labs/currency-model/state/currency"
	ctypes "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	statestrg "github.com/imfact-labs/storage-model/state"
	"github.com/imfact-labs/storage-model/types"
	"github.com/pkg/errors"
)

var createDataItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CreateDataItemProcessor)
	},
}

var createDataProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CreateDataProcessor)
	},
}

func (CreateData) Process(
	_ context.Context, _ base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	return nil, nil, nil
}

type CreateDataItemProcessor struct {
	h      util.Hash
	sender base.Address
	item   CreateDataItem
}

func (ipp *CreateDataItemProcessor) PreProcess(
	_ context.Context, _ base.Operation, getStateFunc base.GetStateFunc,
) error {
	e := util.StringError("preprocess CreateDataItemProcessor")
	it := ipp.item

	if err := it.IsValid(nil); err != nil {
		return e.Wrap(err)
	}

	if err := state.CheckExistsState(statec.DesignStateKey(it.Currency()), getStateFunc); err != nil {
		return e.Wrap(common.ErrCurrencyNF.Wrap(errors.Errorf("currency id %v", it.Currency())))
	}

	if err := state.CheckExistsState(statestrg.DesignStateKey(it.Contract()), getStateFunc); err != nil {
		return e.Wrap(
			common.ErrServiceNF.Errorf("storage service state for contract account %v", it.Contract()))
	}

	if found, _ := state.CheckNotExistsState(statestrg.DataStateKey(it.Contract(), it.DataKey()), getStateFunc); found {
		return e.Wrap(
			common.ErrStateE.Errorf(
				"storage data for key %q in contract account %v",
				it.DataKey(), it.Contract(),
			))
	}

	return nil
}

func (ipp *CreateDataItemProcessor) Process(
	_ context.Context, _ base.Operation, _ base.GetStateFunc,
) ([]base.StateMergeValue, error) {
	it := ipp.item

	var sts []base.StateMergeValue
	data := types.NewData(
		it.DataKey(), it.DataValue(),
	)
	if err := data.IsValid(nil); err != nil {
		return nil, err
	}

	sts = append(sts, state.NewStateMergeValue(
		statestrg.DataStateKey(it.Contract(), it.DataKey()),
		statestrg.NewDataStateValue(data),
	))

	return sts, nil
}

func (ipp *CreateDataItemProcessor) Close() {
	ipp.h = nil
	ipp.sender = nil
	ipp.item = CreateDataItem{}

	createDataItemProcessorPool.Put(ipp)
}

type CreateDataProcessor struct {
	*base.BaseOperationProcessor
}

func NewCreateDataProcessor() ctypes.GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new CreateDataProcessor")

		nopp := createDataProcessorPool.Get()
		opp, ok := nopp.(*CreateDataProcessor)
		if !ok {
			return nil, e.Errorf("expected %T, not %T", CreateDataProcessor{}, nopp)
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

func (opp *CreateDataProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(CreateDataFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMTypeMismatch).
				Errorf("expected %T, not %T", CreateDataFact{}, op.Fact())), nil
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", err)), nil
	}

	for _, it := range fact.Items() {
		ip := createDataItemProcessorPool.Get()
		ipc, ok := ip.(*CreateDataItemProcessor)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError(
				common.ErrMTypeMismatch.Errorf("expected %T, not %T", CreateDataItemProcessor{}, ip)), nil
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

func (opp *CreateDataProcessor) Process( // nolint:dupl
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, _ := op.Fact().(CreateDataFact)

	var sts []base.StateMergeValue // nolint:prealloc
	for _, it := range fact.Items() {
		ip := createDataItemProcessorPool.Get()
		ipc, _ := ip.(*CreateDataItemProcessor)

		ipc.h = op.Hash()
		ipc.sender = fact.Sender()
		ipc.item = it

		st, err := ipc.Process(ctx, op, getStateFunc)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("failed to process CreateDataItem; %w", err), nil
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

func (opp *CreateDataProcessor) Close() error {
	createDataProcessorPool.Put(opp)

	return nil
}
