package state

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum-storage/types"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (sv DesignStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":  sv.Hint().String(),
			"design": sv.Design,
		},
	)
}

type DesignStateValueBSONUnmarshaler struct {
	Hint   string   `bson:"_hint"`
	Design bson.Raw `bson:"design"`
}

func (sv *DesignStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of DesignStateValue")

	var u DesignStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	sv.BaseHinter = hint.NewBaseHinter(ht)

	var sd types.Design
	if err := sd.DecodeBSON(u.Design, enc); err != nil {
		return e.Wrap(err)
	}
	sv.Design = sd

	return nil
}

func (sv DataStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":        sv.Hint().String(),
			"storage_data": sv.Data,
		},
	)
}

type DataStateValueBSONUnmarshaler struct {
	Hint        string   `bson:"_hint"`
	StorageData bson.Raw `bson:"storage_data"`
}

func (sv *DataStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of DataStateValue")

	var u DataStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	sv.BaseHinter = hint.NewBaseHinter(ht)

	var n types.Data
	if err := n.DecodeBSON(u.StorageData, enc); err != nil {
		return e.Wrap(err)
	}
	sv.Data = n

	return nil
}
