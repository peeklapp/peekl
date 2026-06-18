package collectors

import "os"

func GetHostname() (string, error) {
	return os.Hostname()
}
