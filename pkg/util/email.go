package util

import "regexp"

// IsValidEmail ...
func IsValidEmail(email string) bool {
	r := regexp.MustCompile(`^[\w-\.]+@([\w-]+\.)+[\w-]{2,}$`)
	return r.MatchString(email)
}
