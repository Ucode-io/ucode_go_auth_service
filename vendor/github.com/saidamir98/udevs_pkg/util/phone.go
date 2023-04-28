package util

import "regexp"

// IsValidPhone checks if the given string is a valid uzbek phone number.
// It returns true if the string is a valid phone, false otherwise.
//
// Note: This function does not check if the phone is already in use.
func IsValidPhone(phone string) bool {
	r := regexp.MustCompile(`^\+998[0-9]{2}[0-9]{7}$`)
	s := regexp.MustCompile(`^\(?[0-9]{2}\)? [0-9]{3}-[0-9]{2}-[0-9]{2}$`)
	if r.MatchString(phone) {
		return true
	} else if s.MatchString(phone) {
		return true
	}

	return false
}
