package ui

import "github.com/dennwc/dom"

// Page represents a html-page-view
type Page interface {
	Show()
	Hide()
}

// page is the base-implementation of Page
type page struct {
	root *dom.Element
}

// NewPage greates a basic Page from the html-element with the given id
func NewPage(id string) Page {
	return &page{
		root: dom.Doc.GetElementById("id"),
	}
}

func (p *page) Show() {
	p.root.ClassList().Remove("hidden")
}

func (p *page) Hide() {
	p.root.ClassList().Add("hidden")
}
