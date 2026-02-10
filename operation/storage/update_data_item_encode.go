package storage

import (
	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (it *UpdateDataItem) unpack(enc encoder.Encoder, ht hint.Hint,
	cAdr, dataKey, dataValue, cid string,
) error {
	it.BaseHinter = hint.NewBaseHinter(ht)

	switch a, err := base.DecodeAddress(cAdr, enc); {
	case err != nil:
		return err
	default:
		it.contract = a
	}

	it.dataKey = dataKey
	it.dataValue = dataValue
	it.currency = ctypes.CurrencyID(cid)

	return nil
}
