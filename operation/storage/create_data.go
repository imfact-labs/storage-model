package storage

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/operation/extras"
	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-storage/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"github.com/pkg/errors"
)

var (
	CreateDataFactHint = hint.MustNewHint("mitum-storage-create-data-operation-fact-v0.0.1")
	CreateDataHint     = hint.MustNewHint("mitum-storage-create-data-operation-v0.0.1")
)

type CreateDataFact struct {
	base.BaseFact
	sender    base.Address
	contract  base.Address
	dataKey   string
	dataValue string
	currency  ctypes.CurrencyID
}

func NewCreateDataFact(
	token []byte, sender, contract base.Address,
	key, value string, currency ctypes.CurrencyID) CreateDataFact {
	bf := base.NewBaseFact(CreateDataFactHint, token)
	fact := CreateDataFact{
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

func (fact CreateDataFact) IsValid(b []byte) error {
	if len(fact.dataKey) < 1 || len(fact.dataKey) > types.MaxKeyLen {
		return common.ErrFactInvalid.Wrap(
			common.ErrValOOR.Wrap(
				errors.Errorf("invalid data key length %v < 1 or %v > %v", len(fact.dataKey), len(fact.dataKey), types.MaxKeyLen)))
	}

	if !ctypes.ReValidSpcecialCh.Match([]byte(fact.dataKey)) {
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

func (fact CreateDataFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact CreateDataFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact CreateDataFact) Bytes() []byte {
	return util.ConcatBytesSlice(
		fact.Token(),
		fact.sender.Bytes(),
		fact.contract.Bytes(),
		[]byte(fact.dataKey),
		[]byte(fact.dataValue),
		fact.currency.Bytes(),
	)
}

func (fact CreateDataFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact CreateDataFact) Sender() base.Address {
	return fact.sender
}

func (fact CreateDataFact) Contract() base.Address {
	return fact.contract
}

func (fact CreateDataFact) DataKey() string {
	return fact.dataKey
}

func (fact CreateDataFact) DataValue() string {
	return fact.dataValue
}

func (fact CreateDataFact) Currency() ctypes.CurrencyID {
	return fact.currency
}

func (fact CreateDataFact) Addresses() ([]base.Address, error) {
	as := []base.Address{fact.sender}

	return as, nil
}

func (fact CreateDataFact) FeeBase() map[ctypes.CurrencyID][]common.Big {
	required := make(map[ctypes.CurrencyID][]common.Big)
	required[fact.Currency()] = []common.Big{common.ZeroBig}

	return required
}

func (fact CreateDataFact) FeePayer() base.Address {
	return fact.sender
}

func (fact CreateDataFact) FeeItemCount() (uint, bool) {
	return extras.ZeroItem, extras.HasNoItem
}

func (fact CreateDataFact) FactUser() base.Address {
	return fact.sender
}

func (fact CreateDataFact) Signer() base.Address {
	return fact.sender
}

func (fact CreateDataFact) ActiveContractOwnerHandlerOnly() [][2]base.Address {
	return [][2]base.Address{{fact.contract, fact.sender}}
}

type CreateData struct {
	extras.ExtendedOperation
}

func NewCreateData(fact CreateDataFact) (CreateData, error) {
	return CreateData{
		ExtendedOperation: extras.NewExtendedOperation(CreateDataHint, fact),
	}, nil
}
