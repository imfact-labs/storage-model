package types

import (
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
)

type DataJSONMarshaler struct {
	hint.BaseHinter
	DataKey   string `json:"dataKey"`
	DataValue string `json:"dataValue"`
	Deleted   bool   `json:"deleted"`
}

func (d Data) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(DataJSONMarshaler{
		BaseHinter: d.BaseHinter,
		DataKey:    d.dataKey,
		DataValue:  d.dataValue,
		Deleted:    d.isDeleted,
	})
}

type DataJSONUnmarshaler struct {
	Hint      hint.Hint `json:"_hint"`
	DataKey   string    `json:"dataKey"`
	DataValue string    `json:"dataValue"`
	IsDeleted bool      `json:"deleted"`
}

func (d *Data) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("failed to decode json of Data")

	var u DataJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	return d.unmarshal(u.Hint, u.DataKey, u.DataValue, u.IsDeleted)
}
