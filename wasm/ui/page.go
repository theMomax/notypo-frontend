package ui

import (
	"github.com/dennwc/dom"
)

// page represents a html-page-view
type page interface {
	Show()
	Hide()
	Root() *dom.Element
}

// pageimpl is the base-implementation of Page
type pageimpl struct {
	root *dom.Element
}

// initPage creates a basic Page from the html-element with the given id
func initPage(id string) page {
	return &pageimpl{
		root: dom.Doc.GetElementById(id),
	}
}

func initPageFromElement(e *dom.Element) page {
	return &pageimpl{
		root: e,
	}
}

func (p *pageimpl) Show() {
	p.root.ClassList().Remove("hidden")
}

func (p *pageimpl) Hide() {
	p.root.ClassList().Add("hidden")
}

func (p *pageimpl) Root() *dom.Element {
	return p.root
}
