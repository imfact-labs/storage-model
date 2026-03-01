package cmds

import (
	"sync"

	ccmodule "github.com/imfact-labs/currency-model/app/module"
	"github.com/imfact-labs/currency-model/app/modulekit"
	smodule "github.com/imfact-labs/storage-model/module"
)

var composedModules = []modulekit.ModelModule{
	ccmodule.Module{},
	smodule.Module{},
}

var (
	moduleRegistryOnce sync.Once
	moduleRegistry     *modulekit.Registry
	moduleRegistryErr  error
)

func loadModuleRegistry() (*modulekit.Registry, error) {
	moduleRegistryOnce.Do(func() {
		moduleRegistry, moduleRegistryErr = buildModuleRegistry()
	})

	return moduleRegistry, moduleRegistryErr
}

func buildModuleRegistry() (*modulekit.Registry, error) {
	registry := modulekit.NewRegistry()

	for i := range composedModules {
		if err := registry.Register(composedModules[i]); err != nil {
			return nil, err
		}
	}

	for i := range composedModules {
		if err := registry.ValidateModuleContract(composedModules[i].ID()); err != nil {
			return nil, err
		}
	}

	return registry, nil
}

func mustBuildModuleRegistry() *modulekit.Registry {
	registry, err := loadModuleRegistry()
	if err != nil {
		panic(err)
	}

	return registry
}
