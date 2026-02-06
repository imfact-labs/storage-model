package state

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-storage/types"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type DesignStateValueJSONMarshaler struct {
	hint.BaseHinter
	Design types.Design `json:"design"`
}

func (sv DesignStateValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(
		DesignStateValueJSONMarshaler(sv),
	)
}

type DesignStateValueJSONUnmarshaler struct {
	Hint   hint.Hint       `json:"_hint"`
	Design json.RawMessage `json:"design"`
}

func (sv *DesignStateValue) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("failed to decode json of DesignStateValue")

	var u DesignStateValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	sv.BaseHinter = hint.NewBaseHinter(u.Hint)

	var sd types.Design
	if err := sd.DecodeJSON(u.Design, enc); err != nil {
		return e.Wrap(err)
	}
	sv.Design = sd

	return nil
}

type StorageDataStateValueJSONMarshaler struct {
	hint.BaseHinter
	Data types.Data `json:"storage_data"`
}

func (sv DataStateValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(
		StorageDataStateValueJSONMarshaler(sv),
	)
}

type DataStateValueJSONUnmarshaler struct {
	Hint        hint.Hint       `json:"_hint"`
	StorageData json.RawMessage `json:"storage_data"`
}

func (sv *DataStateValue) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("decode json of DataStateValue")

	var u DataStateValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	sv.BaseHinter = hint.NewBaseHinter(u.Hint)

	var t types.Data
	if err := t.DecodeJSON(u.StorageData, enc); err != nil {
		return e.Wrap(err)
	}
	sv.Data = t

	return nil
}
