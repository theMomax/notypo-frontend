package config

import (
	"math/rand"
	"net/url"

	"github.com/theMomax/notypo-backend/api"
)

func init() {
	Backend.BaseURL = &url.URL{
		Scheme: "http",
		Host:   "localhost:4000",
	}
	Game.modificators = make(map[int64]*modificator)

}

type GameConfig struct {
	modificators map[int64]*modificator
	sst          api.StreamSourceType
}

type BackendConfig struct {
	BaseURL *url.URL
}

type modificator struct {
	sst              api.StreamSourceType
	isMultiplicative bool
	action           func(*api.StreamSupplierDescription)
}

var Game GameConfig
var Backend BackendConfig

// StreamSupplierDescription builds and returns a api.StreamSupplierDescription
// based on the configured modificators
func (gc *GameConfig) StreamSupplierDescription() *api.StreamSupplierDescription {
	ssd := &api.StreamSupplierDescription{
		Type:    gc.sst,
		Charset: make([]api.BasicCharacter, 0),
	}
	multiplicatives := make([]func(*api.StreamSupplierDescription), 0)
	for _, m := range gc.modificators {
		if m.sst == gc.sst {
			if m.isMultiplicative {
				multiplicatives = append(multiplicatives, m.action)
			} else {
				m.action(ssd)
			}
		}
	}
	for _, a := range multiplicatives {
		a(ssd)
	}
	return ssd
}

// AddModificator registers a function mod, which can modify the
// StreamSupplierDescription returned by StreamSupplierDescription(), if t
// matches the api.StreamSourceType currently set via SetType. Functions
// registered using the multiplicative flag are executed at last
func (gc *GameConfig) AddModificator(t api.StreamSourceType, mod func(*api.StreamSupplierDescription), multiplicative ...bool) (id int64) {
	id = rand.Int63()
	var m bool
	if len(multiplicative) == 1 {
		m = multiplicative[0]
	}
	gc.modificators[id] = &modificator{
		sst:              t,
		isMultiplicative: m,
		action:           mod,
	}
	return
}

// RemoveModificator removes the modificator with the given id
func (gc *GameConfig) RemoveModificator(id int64) {
	delete(gc.modificators, id)
}

// SetType configures the Type of StreamSupplierDescription()'s return-value and
// the modificators used
func (gc *GameConfig) SetType(t api.StreamSourceType) {
	gc.sst = t
}
