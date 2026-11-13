package domain

import (
	"errors"
	"fmt"
	"strings"
)

type Kind int

const (
	KindInternal Kind = iota
	KindUnavailable
	KindNotFound
	KindInvalidArgument
	KindConflict
	KindTooLarge
)

func (k Kind) String() string {
	switch k {
	case KindUnavailable:
		return "unavailable"
	case KindNotFound:
		return "not_found"
	case KindInvalidArgument:
		return "invalid_argument"
	case KindConflict:
		return "conflict"
	case KindTooLarge:
		return "too_large"
	default:
		return "internal"
	}
}

type AppError struct {
	kind    Kind
	message string
	err     error
}

func (ae *AppError) Error() string {
	if ae.err == nil {
		return ae.message
	}

	cause := ae.err.Error()
	if cause == ae.message {
		return ae.message
	}

	if rest, ok := strings.CutPrefix(cause, ae.message+": "); ok {
		cause = rest
	}

	return fmt.Sprintf("%s: %s", ae.message, cause)
}

func (ae *AppError) Unwrap() error {
	return ae.err
}

func (ae *AppError) Message() string {
	return ae.message
}

func (ae *AppError) Kind() Kind {
	return ae.kind
}

func NewAppError(kind Kind, message string, err error) *AppError {
	return &AppError{
		kind:    kind,
		message: message,
		err:     err,
	}
}

func NewInvalidArgumentError(message string, err error) *AppError {
	return NewAppError(KindInvalidArgument, message, err)
}

func NewUnavailableError(message string, err error) *AppError {
	return NewAppError(KindUnavailable, message, err)
}

func NewNotFoundError(cause error) *AppError {
	return NewAppError(KindNotFound, cause.Error(), cause)
}

func KindOf(err error) Kind {
	if e, ok := errors.AsType[*AppError](err); ok {
		return e.kind
	}

	return KindInternal
}
