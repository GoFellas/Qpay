package utils

import (
	"errors"
	"unicode"
)

func CheckGateway(name string) error {
	err := hasDigits(name)
	if err != nil {
		return err
	}
	return nil
}
func hasDigits(name string) error {
	hasDigits := false

	for _, char := range name {
		if unicode.IsDigit(char) {
			hasDigits = true
			break
		}
	}

	if hasDigits {
		return errors.New("the gateway name must be string")
	}
	return nil
}
