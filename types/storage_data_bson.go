package types

import (
	bsonenc "github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (d Data) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bson.M{
		"_hint":      d.Hint().String(),
		"data_key":   d.dataKey,
		"data_value": d.dataValue,
		"deleted":    d.isDeleted,
	})
}

type DataBSONUnmarshaler struct {
	Hint      string `bson:"_hint"`
	DataKey   string `bson:"data_key"`
	DataValue string `bson:"data_value"`
	IsDeleted bool   `bson:"deleted"`
}

func (d *Data) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of Item")

	var u DataBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return d.unmarshal(ht, u.DataKey, u.DataValue, u.IsDeleted)
}
