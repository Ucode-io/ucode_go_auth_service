package util

import "regexp"

// IsValidEmail checks if the given string is a valid email address.
//
// It returns true if the string is a valid email address, false otherwise.
//
// Note: This function does not check if the email address is already in use.
func IsValidEmail(email string) bool {
	if len(email) > 254 {
		return false
	}
	r := regexp.MustCompile(`^[\w-\.]+@([\w-]+\.)+[\w-]{2,10}$`)
	return r.MatchString(email)
}

func IsValidEmailNew(email string) bool {
	// Define the regular expression pattern for a valid email address
	// This is a basic pattern and may not cover all edge cases
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	// Compile the regular expression
	re := regexp.MustCompile(emailRegex)

	// Use the MatchString method to check if the email matches the pattern
	return re.MatchString(email)
}
