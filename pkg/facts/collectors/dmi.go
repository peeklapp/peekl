package collectors

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/utils"
)

type DmiTypePair struct {
	Name string
	Type reflect.Type
}

func GetDmiData() (models.DmiData, error) {
	var dmiData models.DmiData

	// If no DMI data, return empty
	if !utils.FileExist("/sys/devices/virtual/dmi/id") {
		return dmiData, nil
	}

	rawStruct := map[string]map[string]any{}
	types := []DmiTypePair{
		{"bios", reflect.TypeFor[models.DmiBiosData]()},
		{"board", reflect.TypeFor[models.DmiBoardData]()},
		{"chassis", reflect.TypeFor[models.DmiChassisData]()},
		{"product", reflect.TypeFor[models.DmiProductData]()},
	}

	for iType := range len(types) {
		for num := range types[iType].Type.NumField() {
			field := types[iType].Type.Field(num)
			tag := field.Tag.Get("sysfs")
			data, err := os.ReadFile(filepath.Join("/sys/devices/virtual/dmi/id", tag))
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				return dmiData, fmt.Errorf("Error while trying to read file for DMI data : %s", err.Error())
			}
			rawStruct[types[iType].Name] = map[string]any{tag: string(data)}
		}
	}

	err := mapstructure.Decode(rawStruct, &dmiData)
	if err != nil {
		return dmiData, fmt.Errorf("Error while decoding raw data into struct : %s", err.Error())
	}

	return dmiData, nil
}
