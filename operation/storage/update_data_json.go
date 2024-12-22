package storage

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/operation/extras"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

type UpdateDataFactJSONMarshaler struct {
	mitumbase.BaseFactJSONMarshaler
	Sender    mitumbase.Address `json:"sender"`
	Contract  mitumbase.Address `json:"contract"`
	DataKey   string            `json:"dataKey"`
	DataValue string            `json:"dataValue"`
	Currency  types.CurrencyID  `json:"currency"`
}

func (fact UpdateDataFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(UpdateDataFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Sender:                fact.sender,
		Contract:              fact.contract,
		DataKey:               fact.dataKey,
		DataValue:             fact.dataValue,
		Currency:              fact.currency,
	})
}

type UpdateDataFactJSONUnmarshaler struct {
	mitumbase.BaseFactJSONUnmarshaler
	Sender    string `json:"sender"`
	Contract  string `json:"contract"`
	DataKey   string `json:"dataKey"`
	DataValue string `json:"dataValue"`
	Currency  string `json:"currency"`
}

func (fact *UpdateDataFact) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var u UpdateDataFactJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	fact.BaseFact.SetJSONUnmarshaler(u.BaseFactJSONUnmarshaler)

	if err := fact.unpack(enc, u.Sender, u.Contract, u.DataKey, u.DataValue, u.Currency); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	return nil
}

func (op UpdateData) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(OperationMarshaler{
		BaseOperationJSONMarshaler:           op.BaseOperation.JSONMarshaler(),
		BaseOperationExtensionsJSONMarshaler: op.BaseOperationExtensions.JSONMarshaler(),
	})
}

func (op *UpdateData) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var ubo common.BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *op)
	}

	op.BaseOperation = ubo

	var ueo extras.BaseOperationExtensions
	if err := ueo.DecodeJSON(b, enc); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *op)
	}

	op.BaseOperationExtensions = &ueo

	return nil
}
