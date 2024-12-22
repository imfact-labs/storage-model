package storage

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/operation/extras"
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-storage/types"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"github.com/pkg/errors"
)

var (
	UpdateDataFactHint = hint.MustNewHint("mitum-storage-update-data-operation-fact-v0.0.1")
	UpdateDataHint     = hint.MustNewHint("mitum-storage-update-data-operation-v0.0.1")
)

type UpdateDataFact struct {
	mitumbase.BaseFact
	sender    mitumbase.Address
	contract  mitumbase.Address
	dataKey   string
	dataValue string
	currency  currencytypes.CurrencyID
}

func NewUpdateDataFact(
	token []byte, sender, contract mitumbase.Address,
	key, value string, currency currencytypes.CurrencyID) UpdateDataFact {
	bf := mitumbase.NewBaseFact(UpdateDataFactHint, token)
	fact := UpdateDataFact{
		BaseFact:  bf,
		sender:    sender,
		contract:  contract,
		dataKey:   key,
		dataValue: value,
		currency:  currency,
	}

	fact.SetHash(fact.GenerateHash())
	return fact
}

func (fact UpdateDataFact) IsValid(b []byte) error {
	if len(fact.dataKey) < 1 || len(fact.dataKey) > types.MaxKeyLen {
		return common.ErrFactInvalid.Wrap(
			common.ErrValOOR.Wrap(
				errors.Errorf("invalid data key length %v < 1 or %v > %v", len(fact.dataKey), len(fact.dataKey), types.MaxKeyLen)))
	}

	if !currencytypes.ReValidSpcecialCh.Match([]byte(fact.dataKey)) {
		return common.ErrFactInvalid.Wrap(common.ErrValueInvalid.Wrap(errors.Errorf("date key %s, must match regex `^[^\\s:/?#\\[\\]$@]*$`", fact.dataKey)))
	}

	if len(fact.dataValue) < 1 || len(fact.dataValue) > types.MaxDataLen {
		return common.ErrFactInvalid.Wrap(
			common.ErrValOOR.Wrap(
				errors.Errorf("invalid data value length %v < 1 or %v > %v", len(fact.dataValue), len(fact.dataValue), types.MaxDataLen)))
	}

	if fact.sender.Equal(fact.contract) {
		return common.ErrFactInvalid.Wrap(
			common.ErrSelfTarget.Wrap(errors.Errorf("sender %v is same with contract account", fact.sender)))
	}

	if err := util.CheckIsValiders(nil, false,
		fact.BaseHinter,
		fact.sender,
		fact.contract,
		fact.currency,
	); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if err := common.IsValidOperationFact(fact, b); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	return nil
}

func (fact UpdateDataFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact UpdateDataFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact UpdateDataFact) Bytes() []byte {
	return util.ConcatBytesSlice(
		fact.Token(),
		fact.sender.Bytes(),
		fact.contract.Bytes(),
		[]byte(fact.dataKey),
		[]byte(fact.dataValue),
		fact.currency.Bytes(),
	)
}

func (fact UpdateDataFact) Token() mitumbase.Token {
	return fact.BaseFact.Token()
}

func (fact UpdateDataFact) Sender() mitumbase.Address {
	return fact.sender
}

func (fact UpdateDataFact) Contract() mitumbase.Address {
	return fact.contract
}

func (fact UpdateDataFact) DataKey() string {
	return fact.dataKey
}

func (fact UpdateDataFact) DataValue() string {
	return fact.dataValue
}

func (fact UpdateDataFact) Currency() currencytypes.CurrencyID {
	return fact.currency
}

func (fact UpdateDataFact) Addresses() ([]mitumbase.Address, error) {
	as := []mitumbase.Address{fact.sender}

	return as, nil
}

func (fact UpdateDataFact) FeeBase() map[currencytypes.CurrencyID][]common.Big {
	required := make(map[currencytypes.CurrencyID][]common.Big)
	required[fact.Currency()] = []common.Big{common.ZeroBig}

	return required
}

func (fact UpdateDataFact) FeePayer() mitumbase.Address {
	return fact.sender
}

func (fact UpdateDataFact) FactUser() mitumbase.Address {
	return fact.sender
}

func (fact UpdateDataFact) Signer() mitumbase.Address {
	return fact.sender
}

func (fact UpdateDataFact) ActiveContractOwnerHandlerOnly() [][2]mitumbase.Address {
	return [][2]mitumbase.Address{{fact.contract, fact.sender}}
}

type UpdateData struct {
	extras.ExtendedOperation
}

func NewUpdateData(fact UpdateDataFact) (UpdateData, error) {
	return UpdateData{
		ExtendedOperation: extras.NewExtendedOperation(UpdateDataHint, fact),
	}, nil
}
