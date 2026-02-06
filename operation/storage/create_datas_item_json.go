package storage

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type CreateDatasItemJSONMarshaler struct {
	hint.BaseHinter
	Contract  base.Address      `json:"contract"`
	DataKey   string            `json:"dataKey"`
	DataValue string            `json:"dataValue"`
	Currency  ctypes.CurrencyID `json:"currency"`
}

func (it CreateDatasItem) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CreateDatasItemJSONMarshaler{
		BaseHinter: it.BaseHinter,
		Contract:   it.contract,
		DataKey:    it.dataKey,
		DataValue:  it.dataValue,
		Currency:   it.currency,
	})
}

type CreateDatasItemJSONUnMarshaler struct {
	Hint      hint.Hint `json:"_hint"`
	Contract  string    `json:"contract"`
	DataKey   string    `json:"dataKey"`
	DataValue string    `json:"dataValue"`
	Currency  string    `json:"currency"`
}

func (it *CreateDatasItem) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var uit CreateDatasItemJSONUnMarshaler
	if err := enc.Unmarshal(b, &uit); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *it)
	}

	if err := it.unpack(enc,
		uit.Hint,
		uit.Contract,
		uit.DataKey,
		uit.DataValue,
		uit.Currency,
	); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *it)
	}

	return nil
}
