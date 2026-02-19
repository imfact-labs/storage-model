package types

import (
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
)

func (de *Design) unmarshal(
	_ encoder.Encoder,
	ht hint.Hint,
	pid string,
) error {
	de.BaseHinter = hint.NewBaseHinter(ht)
	de.project = pid

	return nil
}
