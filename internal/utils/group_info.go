package utils

import (
	"os/user"
	"strconv"
)

func GetGroupGidFromName(name string) (int, error) {
	group, err := user.LookupGroup(name)
	if err != nil {
		return 0, err
	}

	// We ignore the error because we know for sure the the value of
	// `group.Uid` is always going to be an integer at least on Linux
	intGid, _ := strconv.Atoi(group.Gid)
	return intGid, nil
}

func GetGroupNameFromGid(gid int) (string, error) {
	group, err := user.LookupGroupId(strconv.Itoa(gid))
	if err != nil {
		return "", err
	}
	return group.Name, nil
}
