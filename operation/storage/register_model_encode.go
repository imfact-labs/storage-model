package storage

import (
	"github.com/imfact-labs/currency-model/types"
	mitumbase "github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util/encoder"
)

func (fact *RegisterModelFact) unpack(
	enc encoder.Encoder,
	sa, ta, pid, cid string,
) error {
	fact.currency = types.CurrencyID(cid)

	sender, err := mitumbase.DecodeAddress(sa, enc)
	if err != nil {
		return err
	}
	fact.sender = sender
	contract, err := mitumbase.DecodeAddress(ta, enc)
	if err != nil {
		return err
	}
	fact.contract = contract
	fact.project = pid

	return nil
}
