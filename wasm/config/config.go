package config

import (
	"net/url"

	"github.com/theMomax/notypo-backend/api"
)

type GameConfig struct {
	StreamSupplierDescription api.StreamSupplierDescription
}

type BackendConfig struct {
	BaseURL *url.URL
}

var Game GameConfig
var Backend BackendConfig
