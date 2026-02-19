package types

import (
	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/mitum2/util/valuehash"
	"github.com/pkg/errors"
)

var DesignHint = hint.MustNewHint("mitum-storage-design-v0.0.1")

type Design struct {
	hint.BaseHinter
	project string
}

func NewDesign(project string) Design {
	return Design{
		BaseHinter: hint.NewBaseHinter(DesignHint),
		project:    project,
	}
}

func (de Design) IsValid([]byte) error {
	if err := util.CheckIsValiders(nil, false,
		de.BaseHinter,
	); err != nil {
		return err
	}
	if len(de.project) > MaxProjectIDLen {
		return common.ErrValOOR.Wrap(errors.Errorf("project length over allowed, %d > %d", len(de.project), MaxProjectIDLen))
	}

	return nil
}

func (de Design) Bytes() []byte {
	return util.ConcatBytesSlice([]byte(de.project))
}

func (de Design) Hash() util.Hash {
	return de.GenerateHash()
}

func (de Design) GenerateHash() util.Hash {
	return valuehash.NewSHA256(de.Bytes())
}

func (de Design) Project() string {
	return de.project
}

func (de Design) Equal(cd Design) bool {
	if de.project != cd.project {
		return false
	}

	return true
}
