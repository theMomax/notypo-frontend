module github.com/theMomax/notypo-frontend

require (
	github.com/dennwc/dom v0.3.0
	github.com/gopherjs/gopherjs v0.0.0-20181103185306-d547d1d9531e // indirect
	github.com/gopherjs/gopherwasm v1.1.0
	github.com/gopherjs/websocket v0.0.0-20181006171635-6fe0f86d1fcb // branch wasm
	github.com/stretchr/testify v1.3.0
	github.com/tevino/abool v0.0.0-20170917061928-9b9efcf221b5
	github.com/theMomax/notypo-backend v0.1.0
	golang.org/x/text v0.3.0
)

replace github.com/theMomax/notypo-backend v0.1.0 => ../notypo-backend
