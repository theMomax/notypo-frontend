// Package comparison contains the logic for comparing two character-streams.
// One "model" stream, that only contains printable Characters, and a
// user-generated attempt, that may also contain backspaces.
// This comparison results in another stream, containing detailed information on
// the stream's state, the current changes and further statistics
package comparison

import (
	"time"
	"unicode"

	"github.com/theMomax/notypo-frontend/wasm/errors"
)

// BS stands for backspace
const BS = '\u0008'

// ErrIllegalModelInput is thrown, if the "model" stream contains a
// non-printable Character as defined by the unicode.IsPrint method
var ErrIllegalModelInput = errors.New("the model-input-stream contained an illegal character", errors.Critical, errors.Input)

// Character is the required type for the input-streams
type Character interface {
	// Rune returns the character's utf-8 representation
	Rune() rune
}

// Comparison is the comparison's output-stream type. It holds the stream's current
// state, the latest changes and some statistics
type Comparison interface {
	State() State
	Changes() []Modification
	Statistics() Statistics
}

// State contains information about whether the input-streams match or not and
// whether this is due to the latest modifications or not
type State interface {
	Correct() bool
	StatusChanged() bool
}

// Modification represents the change of a single Character
type Modification interface {
	Character
	// Position returns the index of the Character that is modified, added or
	// deleted
	Position() int
	// Deletion is true, if this Modification's Character is a backspace
	Deletion() bool
	// Correct returns true, if the considered Character equals its model. This
	// value is always true, if the Modification is a deletion
	Correct() bool
}

// Statistics contains information on the total amount of Characters, words and
// misses, as well as the failure-quote
type Statistics interface {
	// TotalCharacters returns the total amount of Characters minus the amount
	// of backspaces
	TotalCharacters() int
	// CorrectCharacters returns the total amount of correct Characters
	CorrectCharacters() int
	// CorrectWords returns the total amount of correct spaces
	CorrectWords() int
	// TotalMisses returns the total amount of Characters (excluding backspaces)
	// , that do not match their model-equivalent
	TotalMisses() int
	// FailureRate returns the TotalMisses per TotalCharacters
	FailureRate() float64
}

// Compare compares the model and attempt Character-streams and channels the
// result of every attempt-input to the comp output channel. The comparison
// ends, when ether model or attempt is closed, or the (optional) timeout
// duration has passed. The function panics with ErrIllegalModelInput, if the
// "model" stream contains an illegal Character
func Compare(model, attempt <-chan Character, comp chan<- Comparison, timeout ...time.Duration) {
	done := make(chan bool)
	var stop bool

	if len(timeout) == 1 {
		time.AfterFunc(timeout[0], func() {
			done <- true
		})
	}
	go func() {
		<-done
		stop = true
		close(done)
		close(comp)
	}()

	index := 0
	indexOfLastCorrectState := -1
	var oldState State
	oldState = &state{
		correct:       true,
		statusChanged: false,
	}
	var stats Statistics
	stats = &statistics{
		totalCharacters: 0,
		correctWords:    0,
		totalMisses:     0,
		failureRate:     0,
	}
	cap := 1000
	if len(timeout) == 1 {
		cap = int(timeout[0].Seconds() * 10.0)
	}
	mbuf := characterbuffer{
		buffer: make([]Character, 0, cap),
		src:    model,
	}

	for {
		a, aok := <-attempt
		if !aok {
			done <- true
			return
		}

		c := comparison{}
		if a.Rune() == BS {
			if index == 0 {
				continue
			}
			c.state = &state{
				correct:       index-2 <= indexOfLastCorrectState,
				statusChanged: index-2 == indexOfLastCorrectState,
			}
			correctCharacters := stats.CorrectCharacters()
			if oldState.Correct() {
				correctCharacters--
			}
			c.statistics = &statistics{
				totalCharacters:   index - 1,
				correctCharacters: correctCharacters,
				correctWords:      stats.CorrectWords(),
				totalMisses:       stats.TotalMisses(),
				failureRate:       float64(stats.TotalMisses()) / float64(index-1),
			}

			c.changes = []Modification{
				&modification{
					Character: a,
					position:  index - 1,
					deletion:  true,
					correct:   true,
				},
			}

			index--
			if indexOfLastCorrectState == index {
				indexOfLastCorrectState--
			}

		} else {
			m, mok := mbuf.get(index)
			if !mok {
				done <- true
				return
			}
			if !unicode.IsPrint(m.Rune()) {
				done <- true
				r := m.Rune()
				if r == BS {
					r = '\u2190'
				}
				panic(ErrIllegalModelInput.Append(string(r)))
			}

			if m.Rune() == a.Rune() {
				if oldState.Correct() {
					indexOfLastCorrectState = index
					c.state = &state{
						correct:       true,
						statusChanged: false,
					}

					words := stats.CorrectWords()
					if unicode.IsSpace(a.Rune()) {
						words++
					}
					c.statistics = &statistics{
						totalCharacters:   index + 1,
						correctCharacters: stats.CorrectCharacters() + 1,
						correctWords:      words,
						totalMisses:       stats.TotalMisses(),
						failureRate:       float64(stats.TotalMisses()) / float64(index+1),
					}

					c.changes = []Modification{
						&modification{
							Character: a,
							position:  index,
							deletion:  false,
							correct:   true,
						},
					}
					index++
				} else {
					c.state = &state{
						correct:       false,
						statusChanged: false,
					}

					c.statistics = &statistics{
						totalCharacters:   index + 1,
						correctCharacters: stats.CorrectCharacters(),
						correctWords:      stats.CorrectWords(),
						totalMisses:       stats.TotalMisses(),
						failureRate:       float64(stats.TotalMisses()) / float64(index+1),
					}

					c.changes = []Modification{
						&modification{
							Character: a,
							position:  index,
							deletion:  false,
							correct:   true,
						},
					}
					index++
				}
			} else {
				c.state = &state{
					correct:       false,
					statusChanged: oldState.Correct(),
				}

				c.statistics = &statistics{
					totalCharacters:   index + 1,
					correctCharacters: stats.CorrectCharacters(),
					correctWords:      stats.CorrectWords(),
					totalMisses:       stats.TotalMisses() + 1,
					failureRate:       float64(stats.TotalMisses()+1) / float64(index+1),
				}

				c.changes = []Modification{
					&modification{
						Character: a,
						position:  index,
						deletion:  false,
						correct:   false,
					},
				}
				index++
			}
		}
		if stop {
			return
		}
		comp <- &c
		oldState = c.state
		stats = c.statistics
	}
}

type comparison struct {
	state      State
	changes    []Modification
	statistics Statistics
}

type state struct {
	correct       bool
	statusChanged bool
}

type modification struct {
	Character
	position int
	deletion bool
	correct  bool
}

type statistics struct {
	totalCharacters   int
	correctCharacters int
	correctWords      int
	totalMisses       int
	failureRate       float64
}

func (c *comparison) State() State {
	return c.state
}

func (c *comparison) Changes() []Modification {
	return c.changes
}

func (c *comparison) Statistics() Statistics {
	return c.statistics
}

func (g *state) Correct() bool {
	return g.correct
}

func (g *state) StatusChanged() bool {
	return g.statusChanged
}

func (g *modification) Position() int {
	return g.position
}

func (g *modification) Deletion() bool {
	return g.deletion
}

func (g *modification) Correct() bool {
	return g.correct
}

func (g *statistics) TotalCharacters() int {
	return g.totalCharacters
}

func (g *statistics) CorrectCharacters() int {
	return g.correctCharacters
}

func (g *statistics) CorrectWords() int {
	return g.correctWords
}

func (g *statistics) TotalMisses() int {
	return g.totalMisses
}

func (g *statistics) FailureRate() float64 {
	return g.failureRate
}

type characterbuffer struct {
	buffer []Character
	src    <-chan Character
}

func (c *characterbuffer) get(i int) (value Character, ok bool) {
	for i >= len(c.buffer) {
		ch, ok := <-c.src
		if !ok {
			return nil, false
		}
		c.buffer = append(c.buffer, ch)
	}
	return c.buffer[i], true
}
