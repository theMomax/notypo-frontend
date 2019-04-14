package main

import (
	"log"
	"net/url"
	"time"

	"github.com/theMomax/notypo-backend/api"
	"github.com/theMomax/notypo-frontend/wasm/comparison"
	"github.com/theMomax/notypo-frontend/wasm/config"
	"github.com/theMomax/notypo-frontend/wasm/errors"
	"github.com/theMomax/notypo-frontend/wasm/game"
	"github.com/theMomax/notypo-frontend/wasm/ui"
)

func main() {
	config.Backend.BaseURL = &url.URL{
		Scheme: "http",
		Host:   "localhost:4000",
	}
	config.Game = config.GameConfig{
		StreamSupplierDescription: api.StreamSupplierDescription{
			Type: api.Random,
			Charset: []rune{
				'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
			},
		},
	}

	for {
		var started bool
		game.HandleGame(&config.Game,
			func() <-chan comparison.Character {
				return game.ModelInputProvider(&config.Game.StreamSupplierDescription)
			},
			func() <-chan comparison.Character {
				return game.AttemptInputProvider(append(config.Game.StreamSupplierDescription.Charset, comparison.BS), func(r rune) {
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
				log.Println(e.Error())
			})

		time.Sleep(5 * time.Second)
		ui.GP.ClearGame()
	}
}
