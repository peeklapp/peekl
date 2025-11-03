package utils

import (
	"os/user"
	"strconv"
)

func GetUserUidFromUsername(name string) (int, error) {
	user, err := user.Lookup(name)
	if err != nil {
		return 0, err
	}

	// We ignore the error because we know for sure the the value of
	// `user.Uid` is always going to be an integer at least on Linux
	intUid, _ := strconv.Atoi(user.Uid)
	return intUid, nil
}

func GetUserUsernameFromUid(uid int) (string, error) {
	user, err := user.LookupId(strconv.Itoa(uid))
	if err != nil {
		return "", err
	}
	return user.Name, nil
}
