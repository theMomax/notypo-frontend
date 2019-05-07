package errors

import (
	"errors"
	"strings"
)

// characteristics codes
const (
	Critical = iota
	Warning
	Message

	UI
	Server

	Input
	Output
)

// Characteristics describes an Error's characteristics
type Characteristics int

// Error extends the standard error-interface by a Characteristics.
type Error interface {
	error
	// Returns all Error-characteristics
	Description() []Characteristics
	// Returns whether this Error fulfills the given characteristics or not
	Is(Characteristics) bool
	// Returns whether this Error fulfills one of the given characteristics or
	// not
	IsOne(...Characteristics) bool
	// Returns whether this Error fulfills all of the given characteristics or
	//not
	IsAll(...Characteristics) bool
	// Returns a new Error, that contains all information contained in this
	// Error plus the new Characteristics
	Elaborate(...Characteristics) Error
	// Returns a new Error, that contains all information contained in this
	// Error plus the new Descriptions
	Append(...string) Error
	// Returns a new Error, that states the given string as the reason and
	// contains the original Error as a Description
	Prepend(string) Error
	// Returns the most basic error-value
	Type() error
}

// New creates a new Error
func New(message string, characteristics ...Characteristics) Error {
	m := make(map[Characteristics]bool, len(characteristics))
	for _, c := range characteristics {
		m[c] = true
	}
	e := errors.New(message)
	return &defaultError{
		error:       e,
		basicType:   e,
		description: m,
	}
}

type defaultError struct {
	error
	basicType   error
	description map[Characteristics]bool
}

func (e *defaultError) Description() []Characteristics {
	return toSlice(e.description)
}

func (e *defaultError) Is(characteristics Characteristics) bool {
	return e.description[characteristics]
}

func (e *defaultError) IsOne(characteristics ...Characteristics) bool {
	for _, c := range characteristics {
		if e.description[c] {
			return true
		}
	}
	return false
}

func (e *defaultError) IsAll(characteristics ...Characteristics) bool {
	for _, c := range characteristics {
		if !e.description[c] {
			return false
		}
	}
	return true
}

func (e *defaultError) Elaborate(characteristics ...Characteristics) Error {
	n := &defaultError{
		error:       e.error,
		basicType:   e.basicType,
		description: make(map[Characteristics]bool),
	}
	for _, c := range characteristics {
		n.description[c] = true
	}
	for c := range e.description {
		n.description[c] = true
	}
	return n
}

func (e *defaultError) Append(descriptions ...string) Error {
	n := &defaultError{
		error:       errors.New(e.Error() + " (" + strings.Join(descriptions, ") (") + ")"),
		basicType:   e.basicType,
		description: e.description,
	}
	return n
}

func (e *defaultError) Prepend(explanation string) Error {
	n := &defaultError{
		error:       errors.New(explanation + " (" + e.Error() + ")"),
		basicType:   e.basicType,
		description: e.description,
	}
	return n
}

func (e *defaultError) Type() error {
	return e.basicType
}

func toSlice(m map[Characteristics]bool) []Characteristics {
	s := make([]Characteristics, len(m))
	i := 0
	for k := range m {
		s[i] = k
		i++
	}
	return s
}
