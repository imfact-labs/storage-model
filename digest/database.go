package digest

import (
	"context"
	"strconv"

	cdigest "github.com/imfact-labs/currency-model/digest"
	utilc "github.com/imfact-labs/currency-model/digest/util"
	"github.com/imfact-labs/mitum2/base"
	utilm "github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/storage-model/state"
	"github.com/imfact-labs/storage-model/types"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var maxLimit int64 = 50

var (
	DefaultColNameStorage     = "digest_storage"
	DefaultColNameStorageData = "digest_storage_data"
)

func StorageDesign(st *cdigest.Database, contract string) (types.Design, base.State, error) {
	filter := utilc.NewBSONFilter("contract", contract)
	q := filter.D()

	opt := options.FindOne().SetSort(
		utilc.NewBSONFilter("height", -1).D(),
	)
	var sta base.State
	if err := st.MongoClient().GetByFilter(
		DefaultColNameStorage,
		q,
		func(res *mongo.SingleResult) error {
			i, err := cdigest.LoadState(res.Decode, st.Encoders())
			if err != nil {
				return err
			}
			sta = i
			return nil
		},
		opt,
	); err != nil {
		return types.Design{}, nil, utilm.ErrNotFound.WithMessage(err, "storage design by contract account %v", contract)
	}

	if sta != nil {
		de, err := state.GetDesignFromState(sta)
		if err != nil {
			return types.Design{}, nil, err
		}
		return de, sta, nil
	} else {
		return types.Design{}, nil, errors.Errorf("state is nil")
	}
}

func StorageData(db *cdigest.Database, contract, key string) (*types.Data, int64, string, string, bool, error) {
	filter := utilc.NewBSONFilter("contract", contract)
	filter = filter.Add("data_key", key)
	q := filter.D()

	opt := options.FindOne().SetSort(
		utilc.NewBSONFilter("height", -1).D(),
	)
	var data *types.Data
	var height int64
	var operation string
	var timestamp string
	var deleted bool
	var err error
	if err := db.MongoClient().GetByFilter(
		DefaultColNameStorageData,
		q,
		func(res *mongo.SingleResult) error {
			data, height, operation, timestamp, deleted, err = LoadStorageData(res.Decode, db.Encoders())
			if err != nil {
				return err
			}
			return nil
		},
		opt,
	); err != nil {
		return nil, 0, "", "", false, utilm.ErrNotFound.WithMessage(err, "storage data for key %s in contract account %s", key, contract)
	}

	if data != nil {
		return data, height, operation, timestamp, deleted, nil
	} else {
		return nil, 0, "", "", false, errors.Errorf("data is nil")
	}
}

func SotrageDataHistoryByDataKey(
	st *cdigest.Database,
	contract,
	key string,
	reverse bool,
	offset string,
	limit int64,
	callback func(*types.Data, int64, string, string, bool) (bool, error),
) error {
	filter, err := buildStorageDataHistoryFilterByDataKey(contract, key, offset, reverse)
	if err != nil {
		return err
	}

	sr := 1
	if reverse {
		sr = -1
	}

	opt := options.Find().SetSort(
		utilc.NewBSONFilter("height", sr).D(),
	)

	switch {
	case limit <= 0: // no limit
	case limit > maxLimit:
		opt = opt.SetLimit(maxLimit)
	default:
		opt = opt.SetLimit(limit)
	}

	return st.MongoClient().Find(
		context.Background(),
		DefaultColNameStorageData,
		filter,
		func(cursor *mongo.Cursor) (bool, error) {
			data, height, operation, timestamp, deleted, err := LoadStorageData(cursor.Decode, st.Encoders())
			if err != nil {
				return false, err
			}
			return callback(data, height, operation, timestamp, deleted)
		},
		opt,
	)
}

func DataCountByContract(
	st *cdigest.Database,
	contract string,
	deleted bool,
) (int64, error) {
	filterA := bson.A{}

	// filter fot matching collection
	filterContract := bson.D{{"contract", bson.D{{"$in", []string{contract}}}}}
	filterA = append(filterA, filterContract)

	filter := bson.D{}
	if len(filterA) > 0 {
		filter = bson.D{
			{"$and", filterA},
		}
	}

	var docs []StorageDataDocBSONUnMarshaler
	var ctx = context.Background()

	var cursor *mongo.Cursor
	opts := options.Find().SetSort(bson.D{{"_id", 1}})
	c, err := st.MongoClient().Collection(DefaultColNameStorageData).Find(ctx, filter, opts)
	if err != nil {
		return 0, err
	} else {
		defer func() {
			_ = c.Close(ctx)
		}()

		cursor = c
	}
	if err = cursor.All(ctx, &docs); err != nil {
		return 0, err
	}

	dataMap := make(map[string]StorageDataDocBSONUnMarshaler)

	for _, doc := range docs {
		if d, found := dataMap[doc.K]; found {
			if doc.HT > d.HT {
				if deleted || !doc.DL {
					dataMap[doc.K] = doc
				} else {
					delete(dataMap, doc.K)
				}
			}
		} else if !deleted && !doc.DL {
			dataMap[doc.K] = doc
		} else if deleted {
			dataMap[doc.K] = doc
		}
	}

	return int64(len(dataMap)), nil
}

func buildStorageDataHistoryFilterByDataKey(contract, key string, offset string, reverse bool) (bson.D, error) {
	filterA := bson.A{}

	// filter for matching data key
	filterContract := bson.D{{"contract", bson.D{{"$in", []string{contract}}}}}
	filterDataKey := bson.D{{"data_key", bson.D{{"$in", []string{key}}}}}
	filterA = append(filterA, filterContract)
	filterA = append(filterA, filterDataKey)

	// if offset exist, apply offset
	if len(offset) > 0 {
		index, err := strconv.ParseUint(offset, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "invalid index of offset")
		}
		if !reverse {
			filterOffset := bson.D{
				{"height", bson.D{{"$gt", index}}},
			}
			filterA = append(filterA, filterOffset)
		} else {
			filterHeight := bson.D{
				{"height", bson.D{{"$lt", index}}},
			}
			filterA = append(filterA, filterHeight)
		}
	}

	filter := bson.D{}
	if len(filterA) > 0 {
		filter = bson.D{
			{"$and", filterA},
		}
	}

	return filter, nil
}
