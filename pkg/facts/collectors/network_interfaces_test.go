package collectors

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNetworkInterfacesFromDaata(t *testing.T) {
	data, _ := os.ReadFile("testdata/ip_output.json")

	res, err := processIpOutput(string(data))
	if err != nil {
		t.Errorf("Should not have returned an error : %s", err.Error())
	}

	assert.Equal(t, 3, len(res))
	assert.Equal(t, "00:00:00:00:00:00", res[0].MacAddress)
	assert.Equal(t, "lo", res[0].Name)
}
