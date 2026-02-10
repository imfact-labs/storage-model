package storage

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-storage/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

var CreateDataItemHint = hint.MustNewHint("mitum-storage-create-data-item-v0.0.1")

type CreateDataItem struct {
	hint.BaseHinter
	contract  base.Address
	dataKey   string
	dataValue string
	currency  ctypes.CurrencyID
}

func NewCreateDataItem(
	contract base.Address,
	key, value string,
	currency ctypes.CurrencyID,
) CreateDataItem {
	return CreateDataItem{
		BaseHinter: hint.NewBaseHinter(CreateDataItemHint),
		contract:   contract,
		dataKey:    key,
		dataValue:  value,
		currency:   currency,
	}
}

func (it CreateDataItem) Bytes() []byte {
	return util.ConcatBytesSlice(
		it.contract.Bytes(),
		[]byte(it.dataKey),
		[]byte(it.dataValue),
		it.currency.Bytes(),
	)
}

func (it CreateDataItem) IsValid([]byte) error {
	if len(it.dataKey) < 1 || len(it.dataKey) > types.MaxKeyLen {
		return common.ErrFactInvalid.Wrap(
			common.ErrValOOR.Wrap(
				errors.Errorf("invalid data key length %v < 1 or %v > %v", len(it.dataKey), len(it.dataKey), types.MaxKeyLen)))
	}

	if !ctypes.ReValidSpcecialCh.Match([]byte(it.dataKey)) {
		return common.ErrFactInvalid.Wrap(common.ErrValueInvalid.Wrap(errors.Errorf("date key %s, must match regex `^[^\\s:/?#\\[\\]$@]*$`", it.dataKey)))
	}

	if len(it.dataValue) < 1 || len(it.dataValue) > types.MaxDataLen {
		return common.ErrFactInvalid.Wrap(
			common.ErrValOOR.Wrap(
				errors.Errorf("invalid data value length %v < 1 or %v > %v", len(it.dataValue), len(it.dataValue), types.MaxDataLen)))
	}

	if err := util.CheckIsValiders(nil, false,
		it.BaseHinter,
		it.contract,
	); err != nil {
		return common.ErrItemInvalid.Wrap(err)
	}

	return nil
}

func (it CreateDataItem) Contract() base.Address {
	return it.contract
}

func (it CreateDataItem) DataKey() string {
	return it.dataKey
}

func (it CreateDataItem) DataValue() string {
	return it.dataValue
}

func (it CreateDataItem) Currency() ctypes.CurrencyID {
	return it.currency
}

func (it CreateDataItem) Addresses() []base.Address {
	ad := make([]base.Address, 1)

	ad[0] = it.contract

	return ad
}
