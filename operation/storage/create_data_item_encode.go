package storage

import (
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
)

func (it *CreateDataItem) unpack(enc encoder.Encoder, ht hint.Hint,
	cAdr, dataKey, dataValue string,
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

	return nil
}
