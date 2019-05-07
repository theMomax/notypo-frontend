package ui

import (
	"github.com/dennwc/dom"
)

// tab implements the page-interface for tabs
type tab struct {
	page
	button *dom.Element
}

func initTab(buttonid string, page page) page {
	return &tab{
		page:   page,
		button: dom.Doc.GetElementById(buttonid),
	}
}

func initTabFromElement(button *dom.Element, page page) page {
	return &tab{
		page:   page,
		button: button,
	}
}

func (p *tab) Show() {
	p.button.ClassList().Remove("active")
	p.page.Show()
}

func (p *tab) Hide() {
	p.button.ClassList().Add("active")
	p.page.Hide()
}

func (p *tab) Root() *dom.Element {
	return p.page.Root()
}
