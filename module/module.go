package module

import (
	apic "github.com/imfact-labs/currency-model/api"
	"github.com/imfact-labs/currency-model/app/modulekit"
	modapi "github.com/imfact-labs/storage-model/api"
	"github.com/imfact-labs/storage-model/runtime/spec"
	"github.com/imfact-labs/storage-model/runtime/steps"
)

const ID = "storage"

type Module struct{}

var _ modulekit.ModelModule = Module{}

func (Module) ID() string {
	return ID
}

func (Module) Register(reg *modulekit.Registry) error {
	if err := reg.AddHinters(ID, spec.AddedHinters...); err != nil {
		return err
	}

	if err := reg.AddSupportedFacts(ID, spec.AddedSupportedHinters...); err != nil {
		return err
	}

	if err := reg.AddOperationProcessors(ID, modulekit.OperationProcessors{
		Name:      steps.PNameOperationProcessorsMap,
		Func:      steps.POperationProcessorsMap,
		SupportsA: true,
		SupportsB: false,
	}); err != nil {
		return err
	}

	if err := reg.AddAPIRoutes(
		ID,
		modulekit.APIRoute{Path: modapi.HandlerPathStorageDesign, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: modapi.HandlerPathStorageData, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: modapi.HandlerPathStorageDataHistory, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: modapi.HandlerPathStorageDataCount, Methods: []string{"GET"}},
	); err != nil {
		return err
	}

	if err := reg.AddAPIHandlers(ID, modulekit.APIHandlerInitializer{
		Key: "storage.api.handlers",
		Register: func(hd *apic.Handlers, _ bool) {
			modapi.SetHandlers(hd)
		},
	}); err != nil {
		return err
	}

	return reg.AddCLICommands(
		ID,
		modulekit.CLICommand{Key: "operation.storage", Description: "storage operation"},
	)
}
