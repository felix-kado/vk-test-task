package handlers

import (
	"errors"
	"regexp"
	"unicode"
)

var (
	ErrInvalidLogin    = errors.New("login must be 3-50 characters long, start with a letter, and contain only letters, numbers, and underscores")
	ErrInvalidPassword = errors.New("password must be 8-72 characters long and contain at least one uppercase letter, one lowercase letter, one number, and one special character")
)

// ValidateLogin checks if the login meets the requirements
func ValidateLogin(login string) error {
	// Check length
	if len(login) < 3 || len(login) > 50 {
		return ErrInvalidLogin
	}

	// Check first character is a letter
	if !unicode.IsLetter(rune(login[0])) {
		return ErrInvalidLogin
	}

	// Check allowed characters (letters, numbers, underscore)
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, login)
	if !matched {
		return ErrInvalidLogin
	}

	return nil
}

// ValidatePassword checks if the password meets the requirements
func ValidatePassword(password string) error {
	// Check length
	if len(password) < 8 || len(password) > 72 {
		return ErrInvalidPassword
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsNumber(c):
			hasNumber = true
		case isSpecialCharacter(c):
			hasSpecial = true
		}

		// If all requirements are met, we can break early
		if hasUpper && hasLower && hasNumber && hasSpecial {
			break
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return ErrInvalidPassword
	}

	return nil
}

func isSpecialCharacter(r rune) bool {
	specialChars := "!@#$%^&*"
	for _, c := range specialChars {
		if r == c {
			return true
		}
	}
	return false
}
