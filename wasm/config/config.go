package config

import (
	"net/url"
	"time"

	"github.com/theMomax/notypo-backend/api"
)

type GameConfig struct {
	StreamSupplierDescription api.StreamSupplierDescription
	Timeout                   *time.Duration
}

type BackendConfig struct {
	BaseURL *url.URL
}

var Game GameConfig
var Backend BackendConfig
