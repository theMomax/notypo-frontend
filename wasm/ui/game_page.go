package ui

import (
	"strconv"
	"time"

	"github.com/dennwc/dom"
	"github.com/gopherjs/gopherwasm/js"
	"github.com/theMomax/notypo-frontend/wasm/comparison"
)

// GamePage represents the page, which displays the actual game
type GamePage struct {
	page
	stats          *dom.Element
	done           *dom.Element
	todo           *dom.Element
	cpm            *dom.Element
	wpm            *dom.Element
	fr             *dom.Element
	time           *dom.Element
	cursor         *dom.Element
	escapeListener js.Func
	onExit         func()
}

// InitGamePage initializes the page, which displays the actual game
func initGamePage() *GamePage {
	gp := &GamePage{
		page:   initPage("game"),
		stats:  dom.Doc.GetElementById("stats"),
		done:   dom.Doc.GetElementById("done"),
		todo:   dom.Doc.GetElementById("todo"),
		cpm:    dom.Doc.GetElementById("cpm_val"),
		wpm:    dom.Doc.GetElementById("wpm_val"),
		fr:     dom.Doc.GetElementById("fr_val"),
		time:   dom.Doc.GetElementById("time_scale"),
		cursor: dom.Doc.GetElementById("cursor"),
	}

	gp.escapeListener = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if args[0].Get("key").String() == "Escape" {
			gp.onExit()
		}
		return nil
	})

	return gp
}

func (gp *GamePage) Hide() {
	gp.page.Hide()
	js.Global().Call("removeEventListener", "keypress", gp.escapeListener)
}

func (gp *GamePage) Show() {
	gp.page.Show()
	js.Global().Call("addEventListener", "keypress", gp.escapeListener)
}

func (gp *GamePage) OnExit(callback func()) {
	gp.onExit = callback
}

func (gp *GamePage) ClearGame() {
	gp.done.SetInnerHTML("")
	gp.todo.SetInnerHTML("")
	gp.cpm.SetInnerHTML("0")
	gp.wpm.SetInnerHTML("0")
	gp.fr.SetInnerHTML("0%")
	gp.time.ClassList().Remove("timer")
	gp.time.SetAttribute("style", "")
}

func (gp *GamePage) CreateCharacter(item comparison.Character) *dom.Element {
	e := dom.Doc.CreateElement("span")
	s := string((item.Rune()))
	if s == " " {
		s = "&nbsp;"
	}
	e.SetInnerHTML(s)
	gp.todo.AppendChild(e)
	return e
}

func (gp *GamePage) TypeChar(correct bool) {
	gp.PauseCursor()
	c := gp.todo.ChildNodes()[0]
	c = gp.todo.RemoveChild(c).(*dom.Element)
	if correct {
		c.SetClassName("correct")
	} else {
		c.SetClassName("wrong")
	}
	gp.done.AppendChild(c)
}

func (gp *GamePage) DeleteChar() {
	gp.PauseCursor()
	c := gp.done.ChildNodes()[len(gp.done.ChildNodes())-1]
	gp.done.RemoveChild(c)
	c.SetClassName("")
	gp.todo.SetInnerHTML(c.OuterHTML() + gp.todo.InnerHTML())
}

func (gp *GamePage) SetWPM(value float64) {
	gp.wpm.SetInnerHTML(strconv.FormatFloat(value, 'f', 0, 64))
}

func (gp *GamePage) SetCPM(value float64) {
	gp.cpm.SetInnerHTML(strconv.FormatFloat(value, 'f', 0, 64))
}

func (gp *GamePage) SetFR(value float64) {
	gp.fr.SetInnerHTML(strconv.FormatFloat(100*value, 'f', 2, 64) + "%")
}

func (gp *GamePage) PauseCursor() {
	gp.cursor.ParentNode().ReplaceChild(gp.cursor, gp.cursor)
}

func (gp *GamePage) SetTimer(duration time.Duration) {
	gp.time.ClassList().Add("timer")
	gp.time.SetAttribute("style", "animation-duration:"+strconv.FormatFloat(duration.Seconds(), 'f', 2, 64)+"s")
}
