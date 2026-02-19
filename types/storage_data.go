package types

import (
	"github.com/imfact-labs/currency-model/common"
	ctypes "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/pkg/errors"
)

var (
	MaxProjectIDLen = 10
	MaxKeyLen       = 200
	MaxDataLen      = 20000
)

var DataHint = hint.MustNewHint("mitum-storage-data-v0.0.1")

type Data struct {
	hint.BaseHinter
	dataKey   string
	dataValue string
	isDeleted bool
}

func NewData(
	key string,
	value string,
) Data {
	return Data{
		BaseHinter: hint.NewBaseHinter(DataHint),
		dataKey:    key,
		dataValue:  value,
		isDeleted:  false,
	}
}

func (d Data) IsValid([]byte) error {
	if len(d.dataKey) < 1 || len(d.dataKey) > MaxKeyLen {
		return errors.Errorf("invalid key length %v < 1 or %v > %v", len(d.dataKey), len(d.dataKey), MaxKeyLen)
	}

	if !ctypes.ReValidSpcecialCh.Match([]byte(d.dataKey)) {
		return common.ErrFactInvalid.Wrap(common.ErrValueInvalid.Wrap(errors.Errorf("date key %s, must match regex `^[^\\s:/?#\\[\\]$@]*$`", d.dataKey)))
	}

	if len(d.dataValue) > MaxDataLen {
		return errors.Errorf("invalid value length %v > %v", len(d.dataValue), MaxDataLen)
	}

	return nil
}

func (d Data) Bytes() []byte {
	bs := []byte{}

	if d.isDeleted {
		bs = append(bs, 1)
	} else {
		bs = append(bs, 0)
	}

	return util.ConcatBytesSlice(
		[]byte(d.dataKey),
		[]byte(d.dataValue),
		bs,
	)
}

func (d Data) DataKey() string {
	return d.dataKey
}

func (d Data) DataValue() string {
	return d.dataValue
}

func (d Data) IsDeleted() bool {
	return d.isDeleted
}

func (d *Data) SetDeleted() {
	d.isDeleted = true
}

func (d Data) Equal(ct Data) bool {
	if d.dataKey != ct.dataKey {
		return false
	}

	if d.dataValue != ct.dataValue {
		return false
	}

	if d.isDeleted != ct.isDeleted {
		return false
	}

	return true
}
