package ui

import (
	"time"

	"github.com/dennwc/dom"
	"github.com/dennwc/dom/js"
	"github.com/tevino/abool"
)

// ErrorPage represents the page, which prints error-messages
type ErrorPage struct {
	page
	wrapper *dom.Element
	cleared *abool.AtomicBool
}

// InitErrorPage initializes the page, which prints error-messages
func initErrorPage() *ErrorPage {
	p := initPage("error")
	ep := &ErrorPage{
		page:    p,
		wrapper: dom.Doc.GetElementById("error_wrapper"),
		cleared: abool.NewBool(true),
	}

	reload := dom.NewButton("reload")
	reload.OnClick(func(_ dom.Event) {
		js.Get("window").Get("location").Call("reload", false)
	})
	ep.Root().AppendChild(reload)

	return ep
}

// Cleared returns whether this page can be hidden or not
func (ep *ErrorPage) Cleared() bool {
	return ep.cleared.IsSet()
}

// Clear empties the page
func (ep *ErrorPage) Clear() {
	ep.wrapper.SetInnerHTML("")
	ep.cleared.Set()
}

// Print displays the given message as a new message-block
func (ep *ErrorPage) Print(message string) {
	p := dom.NewElement("p")
	p.SetInnerHTML(message)
	ep.wrapper.AppendChild(p)
	ep.cleared.UnSet()
}

// Hide hides the error-page, but it will block, if the page isn't empty
func (ep *ErrorPage) Hide() {
	for !ep.cleared.IsSet() {
		time.Sleep(50 * time.Millisecond)
	}
	ep.page.Hide()
}
