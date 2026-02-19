package state

import (
	"fmt"
	"github.com/imfact-labs/currency-model/common"
	mitumbase "github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/storage-model/types"
	"github.com/pkg/errors"
	"strings"
)

var (
	DesignStateValueHint  = hint.MustNewHint("mitum-storage-design-state-value-v0.0.1")
	StorageStateKeyPrefix = "storage"
	DesignStateKeySuffix  = "design"
)

func StorageStateKey(addr mitumbase.Address) string {
	return fmt.Sprintf("%s:%s", StorageStateKeyPrefix, addr.String())
}

type DesignStateValue struct {
	hint.BaseHinter
	Design types.Design
}

func NewDesignStateValue(design types.Design) DesignStateValue {
	return DesignStateValue{
		BaseHinter: hint.NewBaseHinter(DesignStateValueHint),
		Design:     design,
	}
}

func (sv DesignStateValue) Hint() hint.Hint {
	return sv.BaseHinter.Hint()
}

func (sv DesignStateValue) IsValid([]byte) error {
	e := util.ErrInvalid.Errorf("invalid DesignStateValue")

	if err := sv.BaseHinter.IsValid(DesignStateValueHint.Type().Bytes()); err != nil {
		return e.Wrap(err)
	}

	if err := sv.Design.IsValid(nil); err != nil {
		return e.Wrap(err)
	}

	return nil
}

func (sv DesignStateValue) HashBytes() []byte {
	return sv.Design.Bytes()
}

func GetDesignFromState(st mitumbase.State) (types.Design, error) {
	v := st.Value()
	if v == nil {
		return types.Design{}, errors.Errorf("state value is nil")
	}

	d, ok := v.(DesignStateValue)
	if !ok {
		return types.Design{}, errors.Errorf("expected DesignStateValue but %T", v)
	}

	return d.Design, nil
}

func IsDesignStateKey(key string) bool {
	return strings.HasPrefix(key, StorageStateKeyPrefix) && strings.HasSuffix(key, DesignStateKeySuffix)
}

func DesignStateKey(addr mitumbase.Address) string {
	return fmt.Sprintf("%s:%s", StorageStateKey(addr), DesignStateKeySuffix)
}

var (
	DataStateValueHint = hint.MustNewHint("mitum-storage-data-state-value-v0.0.1")
	DataStateKeySuffix = "data"
)

type DataStateValue struct {
	hint.BaseHinter
	Data types.Data
}

func NewDataStateValue(data types.Data) DataStateValue {
	return DataStateValue{
		BaseHinter: hint.NewBaseHinter(DataStateValueHint),
		Data:       data,
	}
}

func (sv DataStateValue) Hint() hint.Hint {
	return sv.BaseHinter.Hint()
}

func (sv DataStateValue) IsValid([]byte) error {
	e := util.ErrInvalid.Errorf("invalid DataStateValue")

	if err := sv.BaseHinter.IsValid(DataStateValueHint.Type().Bytes()); err != nil {
		return e.Wrap(err)
	}

	if err := sv.Data.IsValid(nil); err != nil {
		return e.Wrap(err)
	}

	return nil
}

func (sv DataStateValue) HashBytes() []byte {
	return sv.Data.Bytes()
}

func GetDataFromState(st mitumbase.State) (types.Data, error) {
	v := st.Value()
	if v == nil {
		return types.Data{}, errors.Errorf("State value is nil")
	}

	ts, ok := v.(DataStateValue)
	if !ok {
		return types.Data{}, common.ErrTypeMismatch.Wrap(errors.Errorf("expected DataStateValue found, %T", v))
	}

	return ts.Data, nil
}

func IsDataStateKey(key string) bool {
	return strings.HasPrefix(key, StorageStateKeyPrefix) && strings.HasSuffix(key, DataStateKeySuffix)
}

func DataStateKey(addr mitumbase.Address, key string) string {
	return fmt.Sprintf("%s:%s:%s", StorageStateKey(addr), key, DataStateKeySuffix)
}
