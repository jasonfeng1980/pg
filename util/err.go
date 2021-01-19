package util

import (
	"errors"
	"fmt"
)

func NewError(code int64, msg string) error{
	return errors.New(fmt.Sprintf("%d,%s", code, msg))
}