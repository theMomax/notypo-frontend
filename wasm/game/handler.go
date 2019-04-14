package game

import (
	"github.com/tevino/abool"
	"github.com/theMomax/notypo-backend/api"
	com "github.com/theMomax/notypo-frontend/wasm/communication"
	"github.com/theMomax/notypo-frontend/wasm/comparison"
	"github.com/theMomax/notypo-frontend/wasm/config"
	"github.com/theMomax/notypo-frontend/wasm/errors"

	"sort"
	"syscall/js"
)

// errors
var (
	ErrAlreadyRunning             = errors.New("the game is already running", errors.Critical)
	ErrUnknownPanicCause          = errors.New("something unexpected happened", errors.Critical)
	ErrBackendCommunicationFailed = errors.New("a critical error occurred while communicating with the backend", errors.Critical, errors.Server)
)

var (
	running = abool.New()
	onStop  = make([]func(), 0)
)

// HandleGame starts a game with the given configuration. If there is already a
// game running, this function panics with ErrAlreadyRunning
func HandleGame(config *config.GameConfig, modelOpener, attemptOpener func() <-chan comparison.Character, modelOutputHandler func(comparison.Character), comparisonOutputHandler func(comparison.Comparison), errorHandler func(errors.Error)) {
	if !running.SetToIf(false, true) {
		panic(ErrAlreadyRunning)
	}
	{
		defer handlePanics(errorHandler)
		cmp := make(chan comparison.Comparison)

		go func() {
			defer handlePanics(errorHandler)
			for {
				c, ok := <-cmp
				if !ok {
					break
				}
				comparisonOutputHandler(c)
			}
		}()

		model := modelOpener()
		mCopy := make(chan comparison.Character, cap(model))
		go func() {
			defer handlePanics(errorHandler)
			for {
				c, ok := <-model
				if !ok {
					Stop()
					break
				}
				{
					defer func() {
						recover() // send on closed channel
					}()
					mCopy <- c
				}
				modelOutputHandler(c)
			}
		}()

		onstop(func() {
			close(mCopy)
		})
		defer handlePanics(errorHandler)
		comparison.Compare(mCopy, attemptOpener(), cmp)
	}
	Stop()
}

// Stop aborts the game currently running
func Stop() {
	if running.SetToIf(true, false) {
		for _, f := range onStop {
			f()
		}
		onStop = make([]func(), 0)
	}
}

// AttemptInputProvider registers EventListeners for the given charset and pipes
// the events into the returned channel asynchronously. If the given charset is
// nil, any key is accepted. If there is already a game running, this function
// panics with ErrAlreadyRunning. In addition all callbacks are called
// asynchronously after each input
func AttemptInputProvider(charset []rune, callbacks ...func(rune)) <-chan comparison.Character {
	sort.Slice(charset, func(i, j int) bool {
		return charset[i] < charset[j]
	})
	apt := make(chan comparison.Character, 5)
	var attemptKeyboardListener js.Func = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		var c rune
		name := args[0].Get("key").String()
		if name == "Backspace" {
			c = comparison.BS
		} else {
			c = []rune(name)[0]
		}
		if charset == nil || contains(charset, c) {
			apt <- character(c)
			for _, f := range callbacks {
				go f(c)
			}
		}
		return nil
	})
	js.Global().Call("addEventListener", "keypress", attemptKeyboardListener)
	// remove attemptKeyboardListener and close channel
	onstop(func() {
		close(apt)
		js.Global().Call("removeEventListener", "keypress", attemptKeyboardListener)
	})
	return apt
}

// ModelInputProvider creates and subscribes to a Character-Stream using the
// backend-api and the given description. It panics, if the server responses
// with a critical error, or, if description is invalid
func ModelInputProvider(description *api.StreamSupplierDescription) <-chan comparison.Character {
	streamID, err := com.CreateRandomStream(config.Backend.BaseURL, description)
	if err != nil {
		panic(err)
	}
	streamConnectionID, err := com.OpenStreamConnection(config.Backend.BaseURL, *streamID)
	if err != nil {
		panic(err)
	}

	done := make(chan bool)
	mod, err := com.ReadStreamConnection(config.Backend.BaseURL, 50, *streamConnectionID, done)
	if err != nil {
		panic(err)
	}

	onstop(func() {
		done <- true
		close(done)
	})

	onstop(func() {
		err = com.CloseStreamConnection(config.Backend.BaseURL, *streamConnectionID)
		if err != nil {
			panic(err)
		}
	})
	return mod
}

type character rune

func (c character) Rune() rune {
	return rune(c)
}

func contains(s []rune, c rune) bool {
	i := sort.Search(len(s), func(i int) bool {
		return s[i] >= c
	})
	if i < len(s) && s[i] == c {
		return true
	}
	return false
}

func onstop(f func()) {
	onStop = append(onStop, f)
}

func handlePanics(h func(errors.Error)) {
	func() {
		e := recover()
		if e == nil {
			return
		}
		err, ok := e.(errors.Error)
		if !ok {
			h(ErrUnknownPanicCause.Append(err.Error()))
		} else {
			if err != nil {
				h(err)
			}
			if err.Is(errors.Critical) {
				Stop()
			}
		}
	}()
}
