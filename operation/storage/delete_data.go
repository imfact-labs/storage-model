package storage

import (
	"fmt"

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
	DeleteDataFactHint = hint.MustNewHint("mitum-storage-delete-data-operation-fact-v0.0.1")
	DeleteDataHint     = hint.MustNewHint("mitum-storage-delete-data-operation-v0.0.1")
)

type DeleteDataFact struct {
	mitumbase.BaseFact
	sender   mitumbase.Address
	contract mitumbase.Address
	dataKey  string
	currency ctypes.CurrencyID
}

func NewDeleteDataFact(
	token []byte, sender, contract mitumbase.Address,
	key string, currency ctypes.CurrencyID) DeleteDataFact {
	bf := mitumbase.NewBaseFact(DeleteDataFactHint, token)
	fact := DeleteDataFact{
		BaseFact: bf,
		sender:   sender,
		contract: contract,
		dataKey:  key,
		currency: currency,
	}

	fact.SetHash(fact.GenerateHash())
	return fact
}

func (fact DeleteDataFact) IsValid(b []byte) error {
	if len(fact.dataKey) < 1 || len(fact.dataKey) > types.MaxKeyLen {
		return common.ErrFactInvalid.Wrap(
			common.ErrValOOR.Wrap(
				errors.Errorf("invalid data key length %v < 1 or %v > %v", len(fact.dataKey), len(fact.dataKey), types.MaxKeyLen)))
	}

	if !ctypes.ReValidSpcecialCh.Match([]byte(fact.dataKey)) {
		return common.ErrFactInvalid.Wrap(common.ErrValueInvalid.Wrap(errors.Errorf("date key %s, must match regex `^[^\\s:/?#\\[\\]$@]*$`", fact.dataKey)))
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

func (fact DeleteDataFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact DeleteDataFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact DeleteDataFact) Bytes() []byte {
	return util.ConcatBytesSlice(
		fact.Token(),
		fact.sender.Bytes(),
		fact.contract.Bytes(),
		[]byte(fact.dataKey),
		fact.currency.Bytes(),
	)
}

func (fact DeleteDataFact) Token() mitumbase.Token {
	return fact.BaseFact.Token()
}

func (fact DeleteDataFact) Sender() mitumbase.Address {
	return fact.sender
}

func (fact DeleteDataFact) Contract() mitumbase.Address {
	return fact.contract
}

func (fact DeleteDataFact) DataKey() string {
	return fact.dataKey
}

func (fact DeleteDataFact) Currency() ctypes.CurrencyID {
	return fact.currency
}

func (fact DeleteDataFact) Addresses() ([]mitumbase.Address, error) {
	as := []mitumbase.Address{fact.sender}

	return as, nil
}

func (fact DeleteDataFact) FeeBase() map[ctypes.CurrencyID][]common.Big {
	required := make(map[ctypes.CurrencyID][]common.Big)
	required[fact.Currency()] = []common.Big{common.ZeroBig}

	return required
}

func (fact DeleteDataFact) FeePayer() mitumbase.Address {
	return fact.sender
}

func (fact DeleteDataFact) FeeItemCount() (uint, bool) {
	return extras.ZeroItem, extras.HasNoItem
}

func (fact DeleteDataFact) FactUser() mitumbase.Address {
	return fact.sender
}

func (fact DeleteDataFact) Signer() mitumbase.Address {
	return fact.sender
}

func (fact DeleteDataFact) ActiveContractOwnerHandlerOnly() [][2]mitumbase.Address {
	return [][2]mitumbase.Address{{fact.contract, fact.sender}}
}

func (fact DeleteDataFact) DupKey() (map[ctypes.DuplicationKeyType][]string, error) {
	r := make(map[ctypes.DuplicationKeyType][]string)
	r[extras.DuplicationKeyTypeSender] = []string{fact.sender.String()}
	r[DuplicationTypeStorageData] = []string{fmt.Sprintf("%s:%s", fact.Contract().String(), fact.DataKey())}

	return r, nil
}

type DeleteData struct {
	extras.ExtendedOperation
}

func NewDeleteData(fact DeleteDataFact) (DeleteData, error) {
	return DeleteData{
		ExtendedOperation: extras.NewExtendedOperation(DeleteDataHint, fact),
	}, nil
}
