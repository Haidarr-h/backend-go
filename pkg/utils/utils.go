package utils

import "strings"

func IsEmail(str string) bool {
	return strings.Contains(str, "@")
}
