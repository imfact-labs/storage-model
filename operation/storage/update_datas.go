package storage

import (
	"fmt"
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/operation/extras"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"github.com/pkg/errors"
)

var UpdateDatasItems uint = 100

var (
	UpdateDatasFactHint = hint.MustNewHint("mitum-storage-update-datas-operation-fact-v0.0.1")
	UpdateDatasHint     = hint.MustNewHint("mitum-storage-update-datas-operation-v0.0.1")
)

type UpdateDatasFact struct {
	mitumbase.BaseFact
	sender mitumbase.Address
	items  []UpdateDatasItem
}

func NewUpdateDatasFact(
	token []byte, sender mitumbase.Address, items []UpdateDatasItem) UpdateDatasFact {
	bf := mitumbase.NewBaseFact(UpdateDatasFactHint, token)
	fact := UpdateDatasFact{
		BaseFact: bf,
		sender:   sender,
		items:    items,
	}

	fact.SetHash(fact.GenerateHash())
	return fact
}

func (fact UpdateDatasFact) IsValid(b []byte) error {
	if n := len(fact.items); n < 1 {
		return common.ErrFactInvalid.Wrap(common.ErrArrayLen.Wrap(errors.Errorf("empty items")))
	} else if n > int(UpdateDatasItems) {
		return common.ErrFactInvalid.Wrap(common.ErrArrayLen.Wrap(errors.Errorf("items, %d over max, %d", n, UpdateDatasItems)))
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

func (fact UpdateDatasFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact UpdateDatasFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact UpdateDatasFact) Bytes() []byte {
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

func (fact UpdateDatasFact) Token() mitumbase.Token {
	return fact.BaseFact.Token()
}

func (fact UpdateDatasFact) Sender() mitumbase.Address {
	return fact.sender
}

func (fact UpdateDatasFact) Items() []UpdateDatasItem {
	return fact.items
}

func (fact UpdateDatasFact) Addresses() ([]mitumbase.Address, error) {
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

func (fact UpdateDatasFact) FeeBase() map[types.CurrencyID][]common.Big {
	required := make(map[types.CurrencyID][]common.Big)

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

func (fact UpdateDatasFact) FeePayer() mitumbase.Address {
	return fact.sender
}

func (fact UpdateDatasFact) FactUser() mitumbase.Address {
	return fact.sender
}

func (fact UpdateDatasFact) Signer() mitumbase.Address {
	return fact.sender
}

func (fact UpdateDatasFact) ActiveContractOwnerHandlerOnly() [][2]mitumbase.Address {
	var arr [][2]mitumbase.Address
	for i := range fact.items {
		arr = append(arr, [2]mitumbase.Address{fact.items[i].contract, fact.sender})
	}
	return arr
}

type UpdateDatas struct {
	extras.ExtendedOperation
}

func NewUpdateDatas(fact UpdateDatasFact) (UpdateDatas, error) {
	return UpdateDatas{
		ExtendedOperation: extras.NewExtendedOperation(UpdateDatasHint, fact),
	}, nil
}
