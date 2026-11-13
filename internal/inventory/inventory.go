package inventory

import (
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/peeklapp/peekl/internal/models"
)

type entity interface {
	GetName() string
	Resolve() error
	SetSourcePath(string)
}

func loadEntities[T any, PT interface {
	*T
	entity
}](codeRoot *os.Root, dir string, decode func(filename string, src []byte) (T, error)) (map[string]T, error) {
	entities := map[string]T{}

	entitiesFiles, err := fs.Glob(codeRoot.FS(), path.Join(dir, "*.hcl"))
	if err != nil {
		return nil, err
	}

	for _, entityFile := range entitiesFiles {
		entityInventoryFileRaw, err := codeRoot.ReadFile(entityFile)
		if err != nil {
			return nil, err
		}

		value, err := decode(entityFile, entityInventoryFileRaw)
		if err != nil {
			return nil, err
		}
		e := PT(&value)

		expectedFileName := path.Join(dir, e.GetName()+".hcl")
		if entityFile != expectedFileName {
			return nil, fmt.Errorf("'%s' found in file '%s'. The file should be named '%s'", e.GetName(), entityFile, expectedFileName)
		}

		if err := e.Resolve(); err != nil {
			return nil, err
		}
		e.SetSourcePath(entityFile)

		entities[e.GetName()] = value
	}

	return entities, nil
}

func decodeNode(filename string, src []byte) (models.NodeInventory, error) {
	var n models.NodeInventoryFile
	err := hclsimple.Decode(filename, src, nil, &n)
	return n.Node, err
}

func decodeGroup(filename string, src []byte) (models.GroupInventory, error) {
	var g models.GroupInventoryFile
	err := hclsimple.Decode(filename, src, nil, &g)
	return g.Group, err
}

func LoadInventory(codeRoot *os.Root) (map[string]models.NodeInventory, map[string]models.GroupInventory, error) {
	nodes, err := loadEntities(codeRoot, "inventory/nodes", decodeNode)
	if err != nil {
		return nil, nil, err
	}

	groups, err := loadEntities(codeRoot, "inventory/groups", decodeGroup)
	if err != nil {
		return nil, nil, err
	}

	return nodes, groups, nil
}
