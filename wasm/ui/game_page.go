package ui

import (
	"strconv"
	"time"

	"github.com/dennwc/dom"
	"github.com/theMomax/notypo-frontend/wasm/comparison"
)

// GamePage represents the page, which displays the actual game
type GamePage struct {
	Page
	Stats  *dom.Element
	Done   *dom.Element
	Todo   *dom.Element
	Errors *dom.Element
	CPM    *dom.Element
	WPM    *dom.Element
	FR     *dom.Element
	Time   *dom.Element
	Cursor *dom.Element
}

// InitGamePage initializes the page, which displays the actual game
func InitGamePage() *GamePage {
	return &GamePage{
		Page:   NewPage("game"),
		Stats:  dom.Doc.GetElementById("stats"),
		Done:   dom.Doc.GetElementById("done"),
		Todo:   dom.Doc.GetElementById("todo"),
		Errors: dom.Doc.GetElementById("errors"),
		CPM:    dom.Doc.GetElementById("cpm_val"),
		WPM:    dom.Doc.GetElementById("wpm_val"),
		FR:     dom.Doc.GetElementById("fr_val"),
		Time:   dom.Doc.GetElementById("time_scale"),
		Cursor: dom.Doc.GetElementById("cursor"),
	}
}

func (gp *GamePage) ClearGame() {
	gp.Done.SetInnerHTML("")
	gp.Todo.SetInnerHTML("")
	gp.Errors.SetInnerHTML("")
	gp.CPM.SetInnerHTML("0")
	gp.WPM.SetInnerHTML("0")
	gp.FR.SetInnerHTML("0%")
	gp.Time.ClassList().Remove("timer")
	gp.Time.SetAttribute("style", "")
}

func (gp *GamePage) CreateCharacter(item comparison.Character) *dom.Element {
	e := dom.Doc.CreateElement("span")
	e.SetInnerHTML(string(item.Rune()))
	gp.Todo.AppendChild(e)
	return e
}

func (gp *GamePage) TypeChar(correct bool) {
	gp.PauseCursor()
	c := gp.Todo.ChildNodes()[0]
	c = gp.Todo.RemoveChild(c).(*dom.Element)
	if correct {
		c.SetClassName("correct")
	} else {
		c.SetClassName("wrong")
	}
	gp.Done.AppendChild(c)
}

func (gp *GamePage) DeleteChar() {
	gp.PauseCursor()
	c := gp.Done.ChildNodes()[len(gp.Done.ChildNodes())-1]
	gp.Done.RemoveChild(c)
	c.SetClassName("")
	gp.Todo.SetInnerHTML(c.OuterHTML() + gp.Todo.InnerHTML())
}

func (gp *GamePage) SetWPM(value float64) {
	gp.WPM.SetInnerHTML(strconv.FormatFloat(value, 'f', 2, 64))
}

func (gp *GamePage) SetCPM(value float64) {
	gp.CPM.SetInnerHTML(strconv.FormatFloat(value, 'f', 2, 64))
}

func (gp *GamePage) SetFR(value float64) {
	gp.FR.SetInnerHTML(strconv.FormatFloat(100*value, 'f', 2, 64) + "%")
}

func (gp *GamePage) PauseCursor() {
	gp.Cursor.ParentNode().ReplaceChild(gp.Cursor, gp.Cursor)
}

func (gp *GamePage) SetTimer(duration time.Duration) {
	gp.Time.ClassList().Add("timer")
	gp.Time.SetAttribute("style", "animation-duration:"+strconv.FormatFloat(duration.Seconds(), 'f', 2, 64)+"s")
}
