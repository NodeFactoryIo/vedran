package util

import "strconv"

func IsValidPortAsInt(port int32) bool {
	return port <= 0 && port > 49151
}

func IsValidPortAsStr(port string) bool {
	intPort, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	return IsValidPortAsInt(int32(intPort))
}