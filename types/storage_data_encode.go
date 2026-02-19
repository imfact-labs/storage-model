package types

import (
	"github.com/imfact-labs/mitum2/util/hint"
)

func (d *Data) unmarshal(
	ht hint.Hint,
	dataKey,
	dataValue string,
	deleted bool,
) error {
	d.BaseHinter = hint.NewBaseHinter(ht)
	d.dataKey = dataKey
	d.dataValue = dataValue
	d.isDeleted = deleted

	return nil
}
