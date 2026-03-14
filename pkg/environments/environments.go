package environments

import "regexp"

func EnvironmentNameIsValid(environmentName string) bool {
	r, _ := regexp.Compile("^[A-Za-z0-9_]")
	return r.MatchString(environmentName)
}
