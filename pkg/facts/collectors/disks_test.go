package collectors

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGatherDisksFromOutput(t *testing.T) {
	data, _ := os.ReadFile("testdata/lsblk_output.json")

	res, err := processLsblkOutput(string(data))
	if err != nil {
		t.Errorf("Should not have been an error : %s", err.Error())
	}

	assert.Equal(t, 1, len(res))
	assert.Equal(t, 3, len(res[0].Partitions))
}
