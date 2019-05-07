// Package ui provides helper-functions for DOM-manipulation
package ui

// shortcuts to the html-pages
var (
	LD page
	CP *ConfigPage
	GP *GamePage
	EP *ErrorPage
)

var (
	pages []page
)

func init() {
	pages = make([]page, 0, 4)
	EP = initErrorPage()
	pages = append(pages, EP)
	LD = initPage("loading")
	pages = append(pages, LD)
	CP = initConfigPage()
	pages = append(pages, CP)
	GP = initGamePage()
	pages = append(pages, GP)

	Visit(CP)
}

// Visit makes target visible and hides the other pages
func Visit(target page) {
	for _, p := range pages {
		if p != target {
			p.Hide()
		} else {
			p.Show()
		}
	}
}

// OnPlay registers a callback function, which is called, when the user hits
// the play-button in the config-page
func OnPlay(callback func()) {
	CP.registerOnPlay(callback)
}
