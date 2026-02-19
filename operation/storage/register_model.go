package storage

import (
	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/operation/extras"
	ctypes "github.com/imfact-labs/currency-model/types"
	mitumbase "github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/mitum2/util/valuehash"
	"github.com/imfact-labs/storage-model/types"
	"github.com/pkg/errors"
)

var (
	RegisterModelFactHint = hint.MustNewHint("mitum-storage-register-model-operation-fact-v0.0.1")
	RegisterModelHint     = hint.MustNewHint("mitum-storage-register-model-operation-v0.0.1")
)

type RegisterModelFact struct {
	mitumbase.BaseFact
	sender   mitumbase.Address
	contract mitumbase.Address
	project  string
	currency ctypes.CurrencyID
}

func NewRegisterModelFact(token []byte, sender, contract mitumbase.Address, project string, currency ctypes.CurrencyID) RegisterModelFact {
	bf := mitumbase.NewBaseFact(RegisterModelFactHint, token)
	fact := RegisterModelFact{
		BaseFact: bf,
		sender:   sender,
		contract: contract,
		project:  project,
		currency: currency,
	}
	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact RegisterModelFact) IsValid(b []byte) error {
	if err := fact.BaseHinter.IsValid(nil); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if fact.sender.Equal(fact.contract) {
		return common.ErrFactInvalid.Wrap(common.ErrSelfTarget.Wrap(errors.Errorf("sender %v is same with contract account", fact.sender)))
	}

	if err := util.CheckIsValiders(nil, false,
		fact.sender,
		fact.contract,
		fact.currency,
	); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if len(fact.project) > types.MaxProjectIDLen {
		return common.ErrValOOR.Wrap(
			errors.Errorf("project length over allowed, %d > %d", len(fact.project), types.MaxProjectIDLen))
	}

	if err := common.IsValidOperationFact(fact, b); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	return nil
}

func (fact RegisterModelFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact RegisterModelFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact RegisterModelFact) Bytes() []byte {
	return util.ConcatBytesSlice(
		fact.Token(),
		fact.sender.Bytes(),
		fact.contract.Bytes(),
		[]byte(fact.project),
		fact.currency.Bytes(),
	)
}

func (fact RegisterModelFact) Token() mitumbase.Token {
	return fact.BaseFact.Token()
}

func (fact RegisterModelFact) Sender() mitumbase.Address {
	return fact.sender
}

func (fact RegisterModelFact) Contract() mitumbase.Address {
	return fact.contract
}

func (fact RegisterModelFact) Addresses() ([]mitumbase.Address, error) {
	return []mitumbase.Address{fact.sender, fact.contract}, nil
}

func (fact RegisterModelFact) Project() string {
	return fact.project
}

func (fact RegisterModelFact) Currency() ctypes.CurrencyID {
	return fact.currency
}

func (fact RegisterModelFact) FeeBase() map[ctypes.CurrencyID][]common.Big {
	required := make(map[ctypes.CurrencyID][]common.Big)
	required[fact.Currency()] = []common.Big{common.ZeroBig}

	return required
}

func (fact RegisterModelFact) FeePayer() mitumbase.Address {
	return fact.sender
}

func (fact RegisterModelFact) FeeItemCount() (uint, bool) {
	return extras.ZeroItem, extras.HasNoItem
}

func (fact RegisterModelFact) FactUser() mitumbase.Address {
	return fact.sender
}

func (fact RegisterModelFact) Signer() mitumbase.Address {
	return fact.sender
}

func (fact RegisterModelFact) InActiveContractOwnerHandlerOnly() [][2]mitumbase.Address {
	return [][2]mitumbase.Address{{fact.contract, fact.sender}}
}

func (fact RegisterModelFact) DupKey() (map[ctypes.DuplicationKeyType][]string, error) {
	r := make(map[ctypes.DuplicationKeyType][]string)
	r[extras.DuplicationKeyTypeSender] = []string{fact.sender.String()}
	r[extras.DuplicationKeyTypeContractStatus] = []string{fact.contract.String()}

	return r, nil
}

type RegisterModel struct {
	extras.ExtendedOperation
}

func NewRegisterModel(fact RegisterModelFact) RegisterModel {
	return RegisterModel{
		ExtendedOperation: extras.NewExtendedOperation(RegisterModelHint, fact),
	}
}
