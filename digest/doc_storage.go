package digest

import (
	mongodb "github.com/ProtoconNet/mitum-currency/v3/digest/mongodb"
	bsonutil "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	cstate "github.com/ProtoconNet/mitum-currency/v3/state"
	"github.com/ProtoconNet/mitum-storage/state"
	"github.com/ProtoconNet/mitum-storage/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type StorageDesignDoc struct {
	mongodb.BaseDoc
	st     base.State
	design types.Design
}

// NewStorageDesignDoc get the State of Storage Design
func NewStorageDesignDoc(st base.State, enc encoder.Encoder) (StorageDesignDoc, error) {
	design, err := state.GetDesignFromState(st)

	if err != nil {
		return StorageDesignDoc{}, err
	}

	b, err := mongodb.NewBaseDoc(nil, st, enc)
	if err != nil {
		return StorageDesignDoc{}, err
	}

	return StorageDesignDoc{
		BaseDoc: b,
		st:      st,
		design:  design,
	}, nil
}

func (doc StorageDesignDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	parsedKey, err := cstate.ParseStateKey(doc.st.Key(), state.StorageStateKeyPrefix, 3)

	m["contract"] = parsedKey[1]
	m["height"] = doc.st.Height()

	return bsonutil.Marshal(m)
}

type StorageDataDoc struct {
	mongodb.BaseDoc
	st        base.State
	timestamp time.Time
	data      types.Data
}

func NewStorageDataDoc(st base.State, timestamp time.Time, enc encoder.Encoder) (StorageDataDoc, error) {
	data, err := state.GetDataFromState(st)
	if err != nil {
		return StorageDataDoc{}, err
	}

	b, err := mongodb.NewBaseDoc(nil, data, enc)
	if err != nil {
		return StorageDataDoc{}, err
	}

	return StorageDataDoc{
		BaseDoc:   b,
		st:        st,
		timestamp: timestamp,
		data:      data,
	}, nil
}

func (doc StorageDataDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	parsedKey, err := cstate.ParseStateKey(doc.st.Key(), state.StorageStateKeyPrefix, 4)
	if err != nil {
		return nil, err
	}

	m["contract"] = parsedKey[1]
	m["data_key"] = doc.data.DataKey()
	m["height"] = doc.st.Height()
	m["operation"] = doc.st.Operations()[0].String()
	m["timestamp"] = doc.timestamp.Format(time.RFC3339Nano)
	m["deleted"] = doc.data.IsDeleted()

	return bsonutil.Marshal(m)
}

type StorageDataDocBSONUnMarshaler struct {
	I  bson.Raw      `bson:"_id,omitempty"`
	E  string        `bson:"_e"`
	D  bson.RawValue `bson:"d"`
	H  bool          `bson:"_hinted"`
	K  string        `bson:"data_key"`
	HT int64         `bson:"height"`
	O  string        `bson:"operation"`
	T  string        `bson:"timestamp"`
	DL bool          `bson:"deleted"`
}
