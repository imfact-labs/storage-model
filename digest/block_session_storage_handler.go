package digest

import (
	currencydigest "github.com/ProtoconNet/mitum-currency/v3/digest"
	"github.com/ProtoconNet/mitum-storage/state"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"go.mongodb.org/mongo-driver/mongo"
)

func PrepareStorage(bs *currencydigest.BlockSession, st mitumbase.State) (string, []mongo.WriteModel, error) {

	switch {
	case state.IsDesignStateKey(st.Key()):
		j, err := handleStorageDesignState(bs, st)
		if err != nil {
			return "", nil, err
		}

		return DefaultColNameStorage, j, nil
	case state.IsDataStateKey(st.Key()):
		j, err := handleStorageDataState(bs, st)
		if err != nil {
			return "", nil, err
		}

		return DefaultColNameStorageData, j, nil
	}

	return "", nil, nil
}

func handleStorageDesignState(bs *currencydigest.BlockSession, st mitumbase.State) ([]mongo.WriteModel, error) {
	if storageDesignDoc, err := NewStorageDesignDoc(st, bs.Database().Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(storageDesignDoc),
		}, nil
	}
}

func handleStorageDataState(bs *currencydigest.BlockSession, st mitumbase.State) ([]mongo.WriteModel, error) {
	if StorageDataDoc, err := NewStorageDataDoc(st, bs.BlockMap().Manifest().ProposedAt(), bs.Database().Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(StorageDataDoc),
		}, nil
	}
}
