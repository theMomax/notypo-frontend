package main

import (
	"time"

	"github.com/dennwc/dom/js"
	"github.com/theMomax/notypo-backend/api"
	"github.com/theMomax/notypo-frontend/wasm/comparison"
	"github.com/theMomax/notypo-frontend/wasm/config"
	"github.com/theMomax/notypo-frontend/wasm/errors"
	"github.com/theMomax/notypo-frontend/wasm/game"
	"github.com/theMomax/notypo-frontend/wasm/ui"
)

func main() {
	defer func() {
		recover()
		<-time.After(3 * time.Second)
		js.Get("window").Get("location").Call("reload", false)
	}()
	starter := make(chan func())
	ui.OnPlay(func() {
		switch config.Game.StreamSupplierDescription().Type {
		case api.Random, api.Dictionary:
			starter <- handleSinglePlayerGame
		}
	})
	for {
		s := <-starter
		s()
		if !ui.EP.Cleared() {
			time.AfterFunc(3*time.Second, ui.EP.Clear)
		}
		ui.Visit(ui.CP)
	}
}

func handleSinglePlayerGame() {
	ui.Visit(ui.GP)
	var started bool
	var errorOccurred bool
	exit := make(chan interface{})
	ui.GP.OnExit(func() {
		go game.Stop()
		go func() {
			exit <- true
		}()
	})

	game.HandleGame(&config.Game,
		func() <-chan comparison.Character {
			return game.ModelInputProvider(config.Game.StreamSupplierDescription())
		},
		func() <-chan comparison.Character {
			return game.AttemptInputProvider(arrayOfCharacters(append(config.Game.StreamSupplierDescription().Charset, comparison.BS)...), func(c api.Character) {
				if !started {
					started = true
					ui.GP.SetTimer(time.Minute)
					time.AfterFunc(time.Minute, game.Stop)
				}
			})
		}, func(c comparison.Character) {
			ui.GP.CreateCharacter(c)
		}, func(c comparison.Comparison) {
			if len(c.Changes()) == 1 {
				change := c.Changes()[0]
				if change.Deletion() {
					ui.GP.DeleteChar()
				} else {
					ui.GP.TypeChar(c.State().Correct())
				}
			}
			ui.GP.SetCPM(float64(c.Statistics().CorrectCharacters()))
			ui.GP.SetWPM(float64(c.Statistics().CorrectWords()))
			ui.GP.SetFR(float64(c.Statistics().FailureRate()))
		}, func(e errors.Error) {
			errorOccurred = true
			ui.EP.Print(e.Error())
			if e.Is(errors.Critical) {
				game.Stop()
				ui.Visit(ui.EP)
			}
		})
	if !errorOccurred {
		select {
		case <-time.After(30 * time.Second):
		case <-exit:
		}
	}
	ui.GP.ClearGame()
}

func arrayOfCharacters(items ...api.BasicCharacter) []api.Character {
	a := make([]api.Character, len(items))
	for i, c := range items {
		a[i] = c
	}
	return a
}
