package storage

import (
	"github.com/imfact-labs/currency-model/common"
	ctypes "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *CreateDataFact) unpack(enc encoder.Encoder, sa string, bit []byte, cid string) error {
	switch a, err := base.DecodeAddress(sa, enc); {
	case err != nil:
		return err
	default:
		fact.sender = a
	}
	fact.currency = ctypes.CurrencyID(cid)

	hit, err := enc.DecodeSlice(bit)
	if err != nil {
		return err
	}

	items := make([]CreateDataItem, len(hit))
	for i := range hit {
		j, ok := hit[i].(CreateDataItem)
		if !ok {
			return common.ErrTypeMismatch.Wrap(errors.Errorf("expected %T, not %T", CreateDataItem{}, hit[i]))
		}

		items[i] = j
	}
	fact.items = items

	return nil
}
