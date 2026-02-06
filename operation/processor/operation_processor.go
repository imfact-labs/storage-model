package processor

import (
	"fmt"

	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/ProtoconNet/mitum-currency/v3/operation/extension"
	cprocessor "github.com/ProtoconNet/mitum-currency/v3/operation/processor"
	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-storage/operation/storage"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/pkg/errors"
)

const (
	DuplicationTypeSender      ctypes.DuplicationType = "sender"
	DuplicationTypeCurrency    ctypes.DuplicationType = "currency"
	DuplicationTypeContract    ctypes.DuplicationType = "contract"
	DuplicationTypeStorageData ctypes.DuplicationType = "storagedata"
)

func CheckDuplication(opr *cprocessor.OperationProcessor, op mitumbase.Operation) error {
	opr.Lock()
	defer opr.Unlock()

	var duplicationTypeSenderID string
	var duplicationTypeCurrencyID string
	var duplicationTypeStorageData []string
	var duplicationTypeContractID string
	var newAddresses []mitumbase.Address

	switch t := op.(type) {
	case currency.CreateAccount:
		fact, ok := t.Fact().(currency.CreateAccountFact)
		if !ok {
			return errors.Errorf("expected Cr has already been deletedeateAccountFact, not %T", t.Fact())
		}
		as, err := fact.Targets()
		if err != nil {
			return errors.Errorf("failed to get Addresses")
		}
		newAddresses = as
		duplicationTypeSenderID = cprocessor.DuplicationKey(fact.Sender().String(), DuplicationTypeSender)
	case currency.UpdateKey:
		fact, ok := t.Fact().(currency.UpdateKeyFact)
		if !ok {
			return errors.Errorf("expected UpdateKeyFact, not %T", t.Fact())
		}
		duplicationTypeSenderID = cprocessor.DuplicationKey(fact.Sender().String(), DuplicationTypeSender)
	case currency.Transfer:
		fact, ok := t.Fact().(currency.TransferFact)
		if !ok {
			return errors.Errorf("expected TransferFact, not %T", t.Fact())
		}
		duplicationTypeSenderID = cprocessor.DuplicationKey(fact.Sender().String(), DuplicationTypeSender)
	case currency.RegisterCurrency:
		fact, ok := t.Fact().(currency.RegisterCurrencyFact)
		if !ok {
			return errors.Errorf("expected RegisterCurrencyFact, not %T", t.Fact())
		}
		duplicationTypeCurrencyID = cprocessor.DuplicationKey(fact.Currency().Currency().String(), DuplicationTypeCurrency)
	case currency.UpdateCurrency:
		fact, ok := t.Fact().(currency.UpdateCurrencyFact)
		if !ok {
			return errors.Errorf("expected UpdateCurrencyFact, not %T", t.Fact())
		}
		duplicationTypeCurrencyID = cprocessor.DuplicationKey(fact.Currency().String(), DuplicationTypeCurrency)
	case currency.Mint:
	case extension.CreateContractAccount:
		fact, ok := t.Fact().(extension.CreateContractAccountFact)
		if !ok {
			return errors.Errorf("expected CreateContractAccountFact, not %T", t.Fact())
		}
		as, err := fact.Targets()
		if err != nil {
			return errors.Errorf("failed to get Addresses")
		}
		newAddresses = as
		duplicationTypeSenderID = cprocessor.DuplicationKey(fact.Sender().String(), DuplicationTypeSender)
		duplicationTypeContractID = cprocessor.DuplicationKey(fact.Sender().String(), DuplicationTypeContract)
	case extension.Withdraw:
		fact, ok := t.Fact().(extension.WithdrawFact)
		if !ok {
			return errors.Errorf("expected WithdrawFact, not %T", t.Fact())
		}
		duplicationTypeSenderID = cprocessor.DuplicationKey(fact.Sender().String(), DuplicationTypeSender)
	case storage.RegisterModel:
		fact, ok := t.Fact().(storage.RegisterModelFact)
		if !ok {
			return errors.Errorf("expected RegisterModelFact, not %T", t.Fact())
		}
		duplicationTypeSenderID = cprocessor.DuplicationKey(fact.Sender().String(), DuplicationTypeSender)
		duplicationTypeContractID = cprocessor.DuplicationKey(fact.Contract().String(), DuplicationTypeContract)
	case storage.CreateData:
		fact, ok := t.Fact().(storage.CreateDataFact)
		if !ok {
			return errors.Errorf("expected CreateDataFact, not %T", t.Fact())
		}
		duplicationTypeSenderID = cprocessor.DuplicationKey(fact.Sender().String(), DuplicationTypeSender)
		duplicationTypeStorageData = []string{cprocessor.DuplicationKey(
			fmt.Sprintf("%s:%s", fact.Contract().String(), fact.DataKey()), DuplicationTypeStorageData)}
	case storage.UpdateData:
		fact, ok := t.Fact().(storage.UpdateDataFact)
		if !ok {
			return errors.Errorf("expected UpdateDataFact, not %T", t.Fact())
		}
		duplicationTypeSenderID = cprocessor.DuplicationKey(fact.Sender().String(), DuplicationTypeSender)
		duplicationTypeStorageData = []string{cprocessor.DuplicationKey(
			fmt.Sprintf("%s:%s", fact.Contract().String(), fact.DataKey()), DuplicationTypeStorageData)}
	case storage.DeleteData:
		fact, ok := t.Fact().(storage.DeleteDataFact)
		if !ok {
			return errors.Errorf("expected DeleteDataFact, not %T", t.Fact())
		}
		duplicationTypeSenderID = cprocessor.DuplicationKey(fact.Sender().String(), DuplicationTypeSender)
		duplicationTypeStorageData = []string{cprocessor.DuplicationKey(
			fmt.Sprintf("%s:%s", fact.Contract().String(), fact.DataKey()), DuplicationTypeStorageData)}
	case storage.CreateDatas:
		fact, ok := t.Fact().(storage.CreateDatasFact)
		if !ok {
			return errors.Errorf("expected CreateDatasFact, not %T", t.Fact())
		}
		duplicationTypeSenderID = cprocessor.DuplicationKey(fact.Sender().String(), DuplicationTypeSender)
		var datas []string
		for _, v := range fact.Items() {
			key := cprocessor.DuplicationKey(fmt.Sprintf("%s:%s", v.Contract().String(), v.DataKey()), DuplicationTypeStorageData)
			datas = append(datas, key)
		}
		duplicationTypeStorageData = datas
	case storage.UpdateDatas:
		fact, ok := t.Fact().(storage.UpdateDatasFact)
		if !ok {
			return errors.Errorf("expected UpdateDatasFact, not %T", t.Fact())
		}
		duplicationTypeSenderID = cprocessor.DuplicationKey(fact.Sender().String(), DuplicationTypeSender)
		var datas []string
		for _, v := range fact.Items() {
			key := cprocessor.DuplicationKey(fmt.Sprintf("%s:%s", v.Contract().String(), v.DataKey()), DuplicationTypeStorageData)
			datas = append(datas, key)
		}
		duplicationTypeStorageData = datas
	default:
		return nil
	}

	if len(duplicationTypeSenderID) > 0 {
		if _, found := opr.Duplicated[duplicationTypeSenderID]; found {
			return errors.Errorf("proposal cannot have duplicated sender, %v", duplicationTypeSenderID)
		}

		opr.Duplicated[duplicationTypeSenderID] = struct{}{}
	}

	if len(duplicationTypeCurrencyID) > 0 {
		if _, found := opr.Duplicated[duplicationTypeCurrencyID]; found {
			return errors.Errorf(
				"cannot register duplicated currency id, %v within a proposal",
				duplicationTypeCurrencyID,
			)
		}

		opr.Duplicated[duplicationTypeCurrencyID] = struct{}{}
	}
	if len(duplicationTypeContractID) > 0 {
		if _, found := opr.Duplicated[duplicationTypeContractID]; found {
			return errors.Errorf(
				"cannot use a duplicated contract for registering in contract model , %v within a proposal",
				duplicationTypeSenderID,
			)
		}

		opr.Duplicated[duplicationTypeContractID] = struct{}{}
	}
	if len(duplicationTypeStorageData) > 0 {
		for _, v := range duplicationTypeStorageData {
			if _, found := opr.Duplicated[v]; found {
				return errors.Errorf(
					"cannot use a duplicated contract-datakey for storage data, %v within a proposal",
					v,
				)
			}
			opr.Duplicated[v] = struct{}{}
		}
	}

	if len(newAddresses) > 0 {
		if err := opr.CheckNewAddressDuplication(newAddresses); err != nil {
			return err
		}
	}

	return nil
}

func GetNewProcessor(opr *cprocessor.OperationProcessor, op mitumbase.Operation) (mitumbase.OperationProcessor, bool, error) {
	switch i, err := opr.GetNewProcessorFromHintset(op); {
	case err != nil:
		return nil, false, err
	case i != nil:
		return i, true, nil
	}

	switch t := op.(type) {
	case currency.CreateAccount,
		currency.UpdateKey,
		currency.Transfer,
		extension.CreateContractAccount,
		extension.Withdraw,
		currency.RegisterCurrency,
		currency.UpdateCurrency,
		currency.Mint,
		storage.RegisterModel,
		storage.UpdateData,
		storage.UpdateDatas,
		storage.CreateDatas,
		storage.DeleteData,
		storage.CreateData:
		return nil, false, errors.Errorf("%T needs SetProcessor", t)
	default:
		return nil, false, nil
	}
}
