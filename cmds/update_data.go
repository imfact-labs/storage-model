package cmds

import (
	"context"

	ccmds "github.com/imfact-labs/currency-model/app/cmds"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/storage-model/operation/storage"
	"github.com/pkg/errors"
)

type UpdateDataCommand struct {
	BaseCommand
	ccmds.OperationFlags
	Sender   ccmds.AddressFlag    `arg:"" name:"sender" help:"sender address" required:"true"`
	Contract ccmds.AddressFlag    `arg:"" name:"contract" help:"contract address" required:"true"`
	Key      string               `arg:"" name:"key" help:"key" required:"true"`
	Value    string               `arg:"" name:"value" help:"value" required:"true"`
	Currency ccmds.CurrencyIDFlag `arg:"" name:"currency" help:"currency id" required:"true"`
	sender   base.Address
	contract base.Address
}

func (cmd *UpdateDataCommand) Run(pctx context.Context) error { // nolint:dupl
	if _, err := cmd.prepare(pctx); err != nil {
		return err
	}

	if err := cmd.parseFlags(); err != nil {
		return err
	}

	op, err := cmd.createOperation()
	if err != nil {
		return err
	}

	ccmds.PrettyPrint(cmd.Out, op)

	return nil
}

func (cmd *UpdateDataCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	a, err := cmd.Sender.Encode(cmd.Encoders.JSON())
	if err != nil {
		return errors.Wrapf(err, "invalid sender format, %q", cmd.Sender)
	} else {
		cmd.sender = a
	}

	a, err = cmd.Contract.Encode(cmd.Encoders.JSON())
	if err != nil {
		return errors.Wrapf(err, "invalid contract format, %q", cmd.Contract)
	} else {
		cmd.contract = a
	}

	if len(cmd.Key) < 1 {
		return errors.Errorf("invalid Key, %s", cmd.Key)
	}

	if len(cmd.Value) < 1 {
		return errors.Errorf("invalid value, %s", cmd.Value)
	}

	return nil
}

func (cmd *UpdateDataCommand) createOperation() (base.Operation, error) { // nolint:dupl
	e := util.StringError("failed to create update data operation")

	item := storage.NewUpdateDataItem(cmd.contract, cmd.Key, cmd.Value)

	fact := storage.NewUpdateDataFact([]byte(cmd.Token), cmd.sender, []storage.UpdateDataItem{item}, cmd.Currency.CID)

	op, err := storage.NewUpdateData(fact)
	if err != nil {
		return nil, e.Wrap(err)
	}
	err = op.Sign(cmd.Privatekey, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, e.Wrap(err)
	}

	return op, nil
}
