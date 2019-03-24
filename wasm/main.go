package main

import (
	"net/url"
	"strconv"
	"time"

	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/theMomax/notypo-backend/api"
	"github.com/theMomax/notypo-frontend/wasm/comparison"
	"github.com/theMomax/notypo-frontend/wasm/config"
	"github.com/theMomax/notypo-frontend/wasm/errors"
	"github.com/theMomax/notypo-frontend/wasm/game"
)

func main() {
	config.Backend.BaseURL = &url.URL{
		Scheme: "http",
		Host:   "localhost:4000",
	}
	timeout := time.Minute
	config.Game = config.GameConfig{
		StreamSupplierDescription: api.StreamSupplierDescription{
			Type: api.Random,
			Charset: []rune{
				'a', 'b', 'c', 'd', 'e', 'f', 'd', 'e', 'f', 'd',
			},
		},
		Timeout: &timeout,
	}

	vecty.SetTitle("notypo")
	gv := &GameView{
		errors:     make([]errors.Error, 0, 1),
		characters: make([]character, 0, 500),
	}

	go func() {
		for {
			var stats comparison.Statistics
			gv.message = ""
			vecty.Rerender(gv)
			game.HandleGame(&config.Game,
				func() <-chan comparison.Character {
					return game.ModelInputProvider(&config.Game.StreamSupplierDescription)
				},
				func() <-chan comparison.Character {
					return game.AttemptInputProvider(nil)
				}, func(c comparison.Character) {
					gv.characters = append(gv.characters, character{c.Rune(), StateUntreated})
					vecty.Rerender(gv)
				}, func(c comparison.Comparison) {
					if len(c.Changes()) == 1 {
						change := c.Changes()[0]
						if change.Deletion() {
							gv.characters[change.Position()].state = StateUntreated
						} else {
							if c.State().Correct() {
								gv.characters[change.Position()].state = StateCorrect
							} else {
								if change.Correct() {
									gv.characters[change.Position()].state = StateInvalid
								} else {
									gv.characters[change.Position()].state = StateWrong
								}
							}
						}
						vecty.Rerender(gv)
					}
					stats = c.Statistics()
				}, func(e errors.Error) {
					gv.errors = append(gv.errors, e)
					vecty.Rerender(gv)
				})
			game.Stop()

			gv.message = strconv.Itoa(stats.CorrectCharacters())
			vecty.Rerender(gv)
			time.Sleep(5 * time.Second)
			gv.errors = make([]errors.Error, 0, 1)
			gv.characters = make([]character, 0, 500)
		}
	}()

	vecty.RenderBody(gv)
}

type GameView struct {
	vecty.Core
	errors     []errors.Error
	characters []character
	message    string
}

func (g *GameView) Render() vecty.ComponentOrHTML {
	chars := make([]vecty.MarkupOrChild, 0, len(g.characters))
	for _, c := range g.characters {
		switch c.state {
		case StateUntreated:
			chars = append(chars, elem.Span(vecty.Text(string(c.Rune()))))
		case StateCorrect:
			chars = append(chars, elem.Span(vecty.Text(string(c.Rune())), vecty.Markup(
				vecty.Style("color", "green"),
			)))
		case StateInvalid:
			chars = append(chars, elem.Span(vecty.Text(string(c.Rune())), vecty.Markup(
				vecty.Style("color", "orange"),
			)))
		case StateWrong:
			chars = append(chars, elem.Span(vecty.Text(string(c.Rune())), vecty.Markup(
				vecty.Style("color", "red"),
			)))
		}
	}
	errs := make([]vecty.MarkupOrChild, 0, len(g.errors))
	for _, e := range g.errors {
		errs = append(errs, elem.Span(vecty.Text(e.Error())))
	}
	return elem.Body(
		elem.Div(
			elem.Paragraph(chars...),
		),
		elem.Div(elem.Span(vecty.Text(g.message))),
		elem.Div(errs...),
	)
}

type character struct {
	r     rune
	state charstate
}

func (c character) Rune() rune {
	return c.r
}

type charstate int

const (
	StateUntreated charstate = iota
	StateCorrect
	StateInvalid
	StateWrong
)
