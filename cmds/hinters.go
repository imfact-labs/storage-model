package cmds

import (
	ccmds "github.com/ProtoconNet/mitum-currency/v3/cmds"
	"github.com/ProtoconNet/mitum-storage/operation/storage"
	"github.com/ProtoconNet/mitum-storage/state"
	"github.com/ProtoconNet/mitum-storage/types"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

var Hinters []encoder.DecodeDetail
var SupportedProposalOperationFactHinters []encoder.DecodeDetail

var AddedHinters = []encoder.DecodeDetail{
	// revive:disable-next-line:line-length-limit

	{Hint: types.DesignHint, Instance: types.Design{}},
	{Hint: types.DataHint, Instance: types.Data{}},

	{Hint: storage.CreateDataHint, Instance: storage.CreateData{}},
	{Hint: storage.UpdateDataHint, Instance: storage.UpdateData{}},
	{Hint: storage.DeleteDataHint, Instance: storage.DeleteData{}},
	{Hint: storage.RegisterModelHint, Instance: storage.RegisterModel{}},
	{Hint: storage.CreateDataItemHint, Instance: storage.CreateDataItem{}},
	{Hint: storage.UpdateDataItemHint, Instance: storage.UpdateDataItem{}},

	{Hint: state.DataStateValueHint, Instance: state.DataStateValue{}},
	{Hint: state.DesignStateValueHint, Instance: state.DesignStateValue{}},
}

var AddedSupportedHinters = []encoder.DecodeDetail{
	{Hint: storage.CreateDataFactHint, Instance: storage.CreateDataFact{}},
	{Hint: storage.UpdateDataFactHint, Instance: storage.UpdateDataFact{}},
	{Hint: storage.DeleteDataFactHint, Instance: storage.DeleteDataFact{}},
	{Hint: storage.RegisterModelFactHint, Instance: storage.RegisterModelFact{}},
}

func init() {
	Hinters = append(Hinters, ccmds.Hinters...)
	Hinters = append(Hinters, AddedHinters...)

	SupportedProposalOperationFactHinters = append(SupportedProposalOperationFactHinters, ccmds.SupportedProposalOperationFactHinters...)
	SupportedProposalOperationFactHinters = append(SupportedProposalOperationFactHinters, AddedSupportedHinters...)
}

func LoadHinters(encs *encoder.Encoders) error {
	for i := range Hinters {
		if err := encs.AddDetail(Hinters[i]); err != nil {
			return errors.Wrap(err, "add hinter to encoder")
		}
	}

	for i := range SupportedProposalOperationFactHinters {
		if err := encs.AddDetail(SupportedProposalOperationFactHinters[i]); err != nil {
			return errors.Wrap(err, "add supported proposal operation fact hinter to encoder")
		}
	}

	return nil
}
