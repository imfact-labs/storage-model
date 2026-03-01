package spec

import (
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/storage-model/operation/storage"
	"github.com/imfact-labs/storage-model/state"
	"github.com/imfact-labs/storage-model/types"
)

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
