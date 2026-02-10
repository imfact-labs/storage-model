package storage

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (it UpdateDataItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":      it.Hint().String(),
			"contract":   it.contract,
			"data_key":   it.dataKey,
			"data_value": it.dataValue,
			"currency":   it.currency,
		},
	)
}

type UpdateDataItemBSONUnmarshaler struct {
	Hint      string `bson:"_hint"`
	Contract  string `bson:"contract"`
	DataKey   string `bson:"data_key"`
	DataValue string `bson:"data_value"`
	Currency  string `bson:"currency"`
}

func (it *UpdateDataItem) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var uit UpdateDataItemBSONUnmarshaler
	if err := bson.Unmarshal(b, &uit); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *it)
	}

	ht, err := hint.ParseHint(uit.Hint)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *it)
	}

	if err := it.unpack(enc, ht,
		uit.Contract,
		uit.DataKey,
		uit.DataValue,
		uit.Currency,
	); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *it)
	}

	return nil
}
