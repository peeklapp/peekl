package models

import (
	"testing"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

func TestCtyToMapNumbers(t *testing.T) {
	src := []byte(`
name  = "nginx"
port  = 80
ratio = 1.5
ports = [80, 443]
`)
	vars := map[string]cty.Value{}
	if err := hclsimple.Decode("vars.hcl", src, nil, &vars); err != nil {
		t.Fatalf("Should not have returned an error : %s", err.Error())
	}

	result, err := CtyToMap(cty.ObjectVal(vars))
	if err != nil {
		t.Fatalf("Should not have returned an error : %s", err.Error())
	}
	assert.Equal(t, map[string]any{
		"name":  "nginx",
		"port":  80,
		"ratio": 1.5,
		"ports": []any{80, 443},
	}, result)
}
