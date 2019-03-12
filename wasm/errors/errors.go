package errors

import "errors"

// Charactersitic codes
const (
	Terminating = iota
	Critical
	Warning
	Message

	UI
	Server

	Input
	Output
)

// Charactersitic describes an Error's characteristic
type Charactersitic int

// Error extends the standard error-interface by a Charactersitic.
type Error interface {
	error
	// Returns all Error-characteristics
	Description() []Charactersitic
	// Returns whether this Error fulfills the given characteristic or not
	Is(Charactersitic) bool
	// Returns whether this Error fulfills one of the given characteristics or
	// not
	IsOne(...Charactersitic) bool
	// Returns whether this Error fulfills all of the given characteristic or
	//not
	IsAll(...Charactersitic) bool
}

// New creates a new Error
func New(message string, characteristics ...Charactersitic) Error {
	m := make(map[Charactersitic]bool, len(characteristics))
	for _, c := range characteristics {
		m[c] = true
	}
	return &defaultError{
		error:       errors.New(message),
		description: m,
	}
}

type defaultError struct {
	error
	description map[Charactersitic]bool
}

func (e *defaultError) Description() []Charactersitic {
	return toSlice(e.description)
}

func (e *defaultError) Is(characteristic Charactersitic) bool {
	return e.description[characteristic]
}

func (e *defaultError) IsOne(characteristics ...Charactersitic) bool {
	for _, c := range characteristics {
		if e.description[c] {
			return true
		}
	}
	return false
}

func (e *defaultError) IsAll(characteristics ...Charactersitic) bool {
	for _, c := range characteristics {
		if !e.description[c] {
			return false
		}
	}
	return true
}

func toSlice(m map[Charactersitic]bool) []Charactersitic {
	s := make([]Charactersitic, len(m))
	i := 0
	for k := range m {
		s[i] = k
		i++
	}
	return s
}
