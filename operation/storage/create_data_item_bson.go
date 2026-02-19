package storage

import (
	"github.com/imfact-labs/currency-model/common"
	bsonenc "github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (it CreateDataItem) MarshalBSON() ([]byte, error) {
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

type CreateDataItemBSONUnmarshaler struct {
	Hint      string `bson:"_hint"`
	Contract  string `bson:"contract"`
	DataKey   string `bson:"data_key"`
	DataValue string `bson:"data_value"`
	Currency  string `bson:"currency"`
}

func (it *CreateDataItem) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var uit CreateDataItemBSONUnmarshaler
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
