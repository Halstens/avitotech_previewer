package domain

import "fmt"

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func IsDomainError(err error, code string) bool {
	if domainErr, ok := err.(*Error); ok {
		return domainErr.Code == code
	}
	return false
}
