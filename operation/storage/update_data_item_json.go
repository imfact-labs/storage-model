package storage

import (
	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
)

type UpdateDataItemJSONMarshaler struct {
	hint.BaseHinter
	Contract  base.Address `json:"contract"`
	DataKey   string       `json:"dataKey"`
	DataValue string       `json:"dataValue"`
}

func (it UpdateDataItem) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(UpdateDataItemJSONMarshaler{
		BaseHinter: it.BaseHinter,
		Contract:   it.contract,
		DataKey:    it.dataKey,
		DataValue:  it.dataValue,
	})
}

type UpdateDataItemJSONUnMarshaler struct {
	Hint      hint.Hint `json:"_hint"`
	Contract  string    `json:"contract"`
	DataKey   string    `json:"dataKey"`
	DataValue string    `json:"dataValue"`
}

func (it *UpdateDataItem) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var uit UpdateDataItemJSONUnMarshaler
	if err := enc.Unmarshal(b, &uit); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *it)
	}

	if err := it.unpack(enc,
		uit.Hint,
		uit.Contract,
		uit.DataKey,
		uit.DataValue,
	); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *it)
	}

	return nil
}
