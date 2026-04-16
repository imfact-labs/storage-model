package storage

import (
	"fmt"

	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/operation/extras"
	"github.com/imfact-labs/currency-model/types"
	mitumbase "github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/mitum2/util/valuehash"
	"github.com/pkg/errors"
)

var UpdateDataItems uint = 100

var (
	UpdateDataFactHint = hint.MustNewHint("mitum-storage-update-data-operation-fact-v0.0.1")
	UpdateDataHint     = hint.MustNewHint("mitum-storage-update-data-operation-v0.0.1")
)

type UpdateDataFact struct {
	mitumbase.BaseFact
	sender   mitumbase.Address
	items    []UpdateDataItem
	currency types.CurrencyID
}

func NewUpdateDataFact(
	token []byte,
	sender mitumbase.Address,
	items []UpdateDataItem,
	currency types.CurrencyID,
) UpdateDataFact {
	bf := mitumbase.NewBaseFact(UpdateDataFactHint, token)
	fact := UpdateDataFact{
		BaseFact: bf,
		sender:   sender,
		items:    items,
		currency: currency,
	}

	fact.SetHash(fact.GenerateHash())
	return fact
}

func (fact UpdateDataFact) IsValid(b []byte) error {
	if n := len(fact.items); n < 1 {
		return common.ErrFactInvalid.Wrap(common.ErrArrayLen.Wrap(errors.Errorf("empty items")))
	} else if n > int(UpdateDataItems) {
		return common.ErrFactInvalid.Wrap(common.ErrArrayLen.Wrap(errors.Errorf("items, %d over max, %d", n, UpdateDataItems)))
	}

	if err := util.CheckIsValiders(nil, false,
		fact.sender,
		fact.currency,
	); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	founds := map[string]struct{}{}
	for _, it := range fact.items {
		if err := it.IsValid(nil); err != nil {
			return common.ErrFactInvalid.Wrap(err)
		}

		if it.contract.Equal(fact.sender) {
			return common.ErrFactInvalid.Wrap(common.ErrSelfTarget.Wrap(errors.Errorf("sender %v is same with contract account", fact.sender)))
		}

		k := fmt.Sprintf("%s-%s", it.contract, it.dataKey)

		if _, found := founds[k]; found {
			return common.ErrFactInvalid.Wrap(common.ErrDupVal.Wrap(errors.Errorf("dataKey %v for contract account %v", it.DataKey(), it.Contract())))
		}

		founds[k] = struct{}{}
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
	is := make([][]byte, len(fact.items))
	for i := range fact.items {
		is[i] = fact.items[i].Bytes()
	}

	return util.ConcatBytesSlice(
		fact.Token(),
		fact.sender.Bytes(),
		fact.currency.Bytes(),
		util.ConcatBytesSlice(is...),
	)
}

func (fact UpdateDataFact) Token() mitumbase.Token {
	return fact.BaseFact.Token()
}

func (fact UpdateDataFact) Sender() mitumbase.Address {
	return fact.sender
}

func (fact UpdateDataFact) Items() []UpdateDataItem {
	return fact.items
}

func (fact UpdateDataFact) Currency() types.CurrencyID {
	return fact.currency
}

func (fact UpdateDataFact) Addresses() ([]mitumbase.Address, error) {
	var as []mitumbase.Address

	adrMap := make(map[string]struct{})
	for i := range fact.items {
		for j := range fact.items[i].Addresses() {
			if _, found := adrMap[fact.items[i].Addresses()[j].String()]; !found {
				adrMap[fact.items[i].Addresses()[j].String()] = struct{}{}
				as = append(as, fact.items[i].Addresses()[j])
			}
		}
	}
	as = append(as, fact.sender)

	return as, nil
}

func (fact UpdateDataFact) FeeBase() (types.CurrencyID, int, int, bool) {
	return fact.Currency(), len(fact.items), len(fact.Bytes()), extras.HasItem
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
	var arr [][2]mitumbase.Address
	for i := range fact.items {
		arr = append(arr, [2]mitumbase.Address{fact.items[i].contract, fact.sender})
	}
	return arr
}

func (fact UpdateDataFact) DupKey() (map[types.DuplicationKeyType][]string, error) {
	r := make(map[types.DuplicationKeyType][]string)
	r[extras.DuplicationKeyTypeSender] = []string{fact.sender.String()}
	for _, item := range fact.items {
		r[DuplicationTypeStorageData] = append(
			r[DuplicationTypeStorageData], fmt.Sprintf("%s:%s", item.Contract().String(), item.DataKey()))
	}

	return r, nil
}

type UpdateData struct {
	extras.ExtendedOperation
}

func NewUpdateData(fact UpdateDataFact) (UpdateData, error) {
	return UpdateData{
		ExtendedOperation: extras.NewExtendedOperation(UpdateDataHint, fact),
	}, nil
}
