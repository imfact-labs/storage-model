package storage

import (
	"fmt"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/operation/extras"
	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"github.com/pkg/errors"
)

type DataItem interface {
	util.Byter
	util.IsValider
	Currency() ctypes.CurrencyID
}

var CreateDataItems uint = 100

var (
	CreateDataFactHint = hint.MustNewHint("mitum-storage-create-data-operation-fact-v0.0.1")
	CreateDataHint     = hint.MustNewHint("mitum-storage-create-data-operation-v0.0.1")
)

type CreateDataFact struct {
	mitumbase.BaseFact
	sender mitumbase.Address
	items  []CreateDataItem
}

func NewCreateDataFact(
	token []byte, sender mitumbase.Address, items []CreateDataItem) CreateDataFact {
	bf := mitumbase.NewBaseFact(CreateDataFactHint, token)
	fact := CreateDataFact{
		BaseFact: bf,
		sender:   sender,
		items:    items,
	}

	fact.SetHash(fact.GenerateHash())
	return fact
}

func (fact CreateDataFact) IsValid(b []byte) error {
	if n := len(fact.items); n < 1 {
		return common.ErrFactInvalid.Wrap(common.ErrArrayLen.Wrap(errors.Errorf("empty items")))
	} else if n > int(CreateDataItems) {
		return common.ErrFactInvalid.Wrap(common.ErrArrayLen.Wrap(errors.Errorf("items, %d over max, %d", n, CreateDataItems)))
	}

	if err := util.CheckIsValiders(nil, false,
		fact.sender,
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

func (fact CreateDataFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact CreateDataFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact CreateDataFact) Bytes() []byte {
	is := make([][]byte, len(fact.items))
	for i := range fact.items {
		is[i] = fact.items[i].Bytes()
	}

	return util.ConcatBytesSlice(
		fact.Token(),
		fact.sender.Bytes(),
		util.ConcatBytesSlice(is...),
	)
}

func (fact CreateDataFact) Token() mitumbase.Token {
	return fact.BaseFact.Token()
}

func (fact CreateDataFact) Signer() mitumbase.Address {
	return fact.sender
}

func (fact CreateDataFact) Items() []CreateDataItem {
	return fact.items
}

func (fact CreateDataFact) Addresses() ([]mitumbase.Address, error) {
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

func (fact CreateDataFact) FeeBase() map[ctypes.CurrencyID][]common.Big {
	required := make(map[ctypes.CurrencyID][]common.Big)

	for i := range fact.items {
		zeroBig := common.ZeroBig
		cid := fact.items[i].Currency()
		var amsTemp []common.Big
		if ams, found := required[cid]; found {
			ams = append(ams, zeroBig)
			required[cid] = ams
		} else {
			amsTemp = append(amsTemp, zeroBig)
			required[cid] = amsTemp
		}
	}

	return required
}

func (fact CreateDataFact) FeePayer() mitumbase.Address {
	return fact.sender
}

func (fact CreateDataFact) FeeItemCount() (uint, bool) {
	return uint(len(fact.items)), extras.HasItem
}

func (fact CreateDataFact) FactUser() mitumbase.Address {
	return fact.sender
}

func (fact CreateDataFact) Sender() mitumbase.Address {
	return fact.sender
}

func (fact CreateDataFact) ActiveContractOwnerHandlerOnly() [][2]mitumbase.Address {
	var arr [][2]mitumbase.Address
	for i := range fact.items {
		arr = append(arr, [2]mitumbase.Address{fact.items[i].contract, fact.sender})
	}
	return arr
}

type CreateData struct {
	extras.ExtendedOperation
}

func NewCreateData(fact CreateDataFact) (CreateData, error) {
	return CreateData{
		ExtendedOperation: extras.NewExtendedOperation(CreateDataHint, fact),
	}, nil
}
