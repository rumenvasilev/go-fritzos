package auth

import (
	"errors"
	"fmt"
)

var (
	ErrSessionInvalid           = errSessionInvalid()
	ErrInvalidHeaderContentType = errInvalidHeaderContentType()
	ErrUnsupportedChallenge     = errUnsupportedChallenge()
)

func errSessionInvalid() error {
	return errors.New("login failed, session id is wrong")
}

func errInvalidHeaderContentType() error {
	return errors.New("expected xml content-type, but got something else")
}

func errUnsupportedChallenge() error {
	return errors.New("cannot solve challenge, input string is not in the expected format")
}

type BlockTimeError struct {
	Duration int
	Message  string
}

func (e *BlockTimeError) Error() string {
	return fmt.Sprintf("%s, please wait %d seconds", e.Message, e.Duration)
}
