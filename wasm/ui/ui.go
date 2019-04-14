// Package ui provides helper-functions for DOM-manipulation
package ui

import (
	"github.com/dennwc/dom"
)

// Portrayable defines a object, that has a raw- or html-representation
type Portrayable interface {
	Portrayal() string
}

// shortcuts to the html elements with the corresponding ID
var (
	GP *GamePage
)

func init() {
	GP = InitGamePage()
}

func Create(parent *dom.Element, item Portrayable) *dom.Element {
	e := dom.Doc.CreateElement("span")
	e.SetInnerHTML(item.Portrayal())
	parent.AppendChild(e)
	return e
}
