package ui

import (
	"net/url"
	"sort"
	"strings"
	"unicode"

	"github.com/dennwc/dom"
	"github.com/dennwc/dom/js"
	"github.com/theMomax/notypo-backend/api"
	com "github.com/theMomax/notypo-frontend/wasm/communication"
	"github.com/theMomax/notypo-frontend/wasm/config"
	"golang.org/x/text/language"
)

// ConfigPage represents the page, which contains settings and options
type ConfigPage struct {
	page
	lang          language.Tag
	typeWrapper   *dom.Element
	optionWrapper *dom.Element
	startWrapper  *dom.Element
	playButton    *dom.Button
	onPlay        []func()
	optionpages   []page
}

type option interface {
	Description() string
	EnabledByDefault() bool
	OnEnable() func()
	OnDisable() func()
}

type setting interface {
	Name() string
	Description() string
	Options() []option
}

type gameType interface {
	SST() api.StreamSourceType
	Name() string
	Description() string
	Settings() []setting
}

var settings map[api.StreamSourceType]gameType

const defaultStreamSourceType = api.Random

// initConfigPage initializes the page, which displays settings and options
func initConfigPage() *ConfigPage {
	cp := &ConfigPage{
		page:          initPage("config"),
		typeWrapper:   dom.Doc.GetElementById("game_types"),
		optionWrapper: dom.Doc.GetElementById("game_options"),
		startWrapper:  dom.Doc.GetElementById("game_start"),
		playButton:    dom.NewButton("Play"),
		optionpages:   make([]page, 0),
	}
	u, err := url.Parse(js.Get("window").Get("location").Get("href").String())
	if err != nil {
		EP.Print(err.Error())
		Visit(EP)
	}
	m := language.NewMatcher([]language.Tag{
		language.English,
		language.German,
	})
	langstring := language.English.String()
	if len(u.Query()["lang"]) > 0 {
		langstring = u.Query()["lang"][0]
	}
	tag, err := language.Parse(langstring)
	if err != nil {
		EP.Print(err.Error())
		tag = language.English
	}
	cp.lang, _, _ = m.Match(tag)

	settings = make(map[api.StreamSourceType]gameType)
	settings[api.Random] = &random{cp.lang}

	cp.playButton.OnClick(func(e dom.Event) {
		for _, c := range cp.onPlay {
			c()
		}
	})
	cp.startWrapper.AppendChild(cp.playButton)

	types, err := com.StreamOptions(config.Backend.BaseURL)
	if err != nil {
		EP.Print(err.Error())
		Visit(EP)
	}
	relevantTypes := make([]gameType, 0)
	for _, t := range types {
		if s, ok := settings[t]; ok {
			relevantTypes = append(relevantTypes, s)
		}
	}
	cp.buildPage(relevantTypes)
	return cp
}

func (cp *ConfigPage) buildPage(relevantTypes []gameType) {
	if len(relevantTypes) == 0 {
		EP.Print("no game-modes available")
		Visit(EP)
	}
	for _, t := range relevantTypes {
		b := dom.NewButton(t.Name())
		p := initTabFromElement(&b.Element, cp.buildOptionsPage(t))
		b.OnClick(func(dom.Event) {
			cp.visit(p)
			config.Game.SetType(t.SST())
		})
		if t.SST() == defaultStreamSourceType {
			config.Game.SetType(t.SST())
		}
		cp.typeWrapper.AppendChild(b)
		cp.optionpages = append(cp.optionpages)
	}
}

func (cp *ConfigPage) buildOptionsPage(t gameType) page {
	p := dom.NewElement("div")
	description := dom.NewElement("div")
	description.ClassList().Add("game_description")
	description.SetInnerHTML(t.Description())
	p.AppendChild(description)
	for _, s := range t.Settings() {
		settings := dom.NewElement("div")
		p.AppendChild(settings)
		sn := dom.NewElement("div")
		sn.SetInnerHTML(s.Name())
		sn.ClassList().Add("name")
		settings.AppendChild(sn)
		sd := dom.NewElement("div")
		sd.SetInnerHTML(s.Description())
		sd.ClassList().Add("description")
		settings.AppendChild(sd)
		for _, o := range s.Options() {
			opt := dom.NewButton(o.Description())
			if o.EnabledByDefault() {
				opt.ClassList().Add("active")
				o.OnEnable()()
			}
			func(o option) {
				opt.OnClick(func(e dom.Event) {
					if strings.Contains(" "+opt.GetAttribute("class").String()+" ", " active ") {
						opt.ClassList().Remove("active")
						o.OnDisable()()
					} else {
						opt.ClassList().Add("active")
						o.OnEnable()()
					}
				})
			}(o)
			settings.AppendChild(opt)
		}
	}
	cp.optionWrapper.AppendChild(p)
	return initPageFromElement(p)
}

func (cp *ConfigPage) visit(target page) {
	for _, p := range cp.optionpages {
		if p != target {
			p.Hide()
		} else {
			p.Show()
		}
	}
}

func (cp *ConfigPage) registerOnPlay(callback func()) {
	cp.onPlay = append(cp.onPlay, callback)
}

type random struct {
	lang language.Tag
}

func (r *random) SST() api.StreamSourceType {
	return api.Random
}

func (r *random) Name() string {
	return "Random"
}

func (r *random) Description() string {
	return "Speed-Test with random characters."
}

func (r *random) Settings() []setting {
	return []setting{&charset{api.Random, r.lang}}
}

type charset struct {
	sst  api.StreamSourceType
	lang language.Tag
}

func (c *charset) Name() string {
	return "Charset"
}

func (c *charset) Description() string {
	return "The characters contained in the model-text."
}

func (c *charset) Options() []option {
	return []option{
		&charsetoption{inner{}, c.sst, c.lang, nil},
		&charsetoption{indexfingers{}, c.sst, c.lang, nil},
		&charsetoption{middlefingers{}, c.sst, c.lang, nil},
		&charsetoption{ringfingers{}, c.sst, c.lang, nil},
		&charsetoption{littlefingers{}, c.sst, c.lang, nil},
		&charsetoption{thumbs{}, c.sst, c.lang, nil},
		&shift{c.sst, nil},
	}
}

type charsetoption struct {
	charsetprovider
	sst         api.StreamSourceType
	lang        language.Tag
	modificator *int64
}

func (c *charsetoption) Description() string {
	return c.description(c.lang) + ": " + charsetDescription(c.chars(c.lang))
}

func (c *charsetoption) EnabledByDefault() bool {
	return true
}

func (c *charsetoption) OnEnable() func() {
	return func() {
		if c.modificator == nil {
			id := config.Game.AddModificator(c.sst, func(ssd *api.StreamSupplierDescription) {
				ssd.Charset = extend(ssd.Charset, c.chars(c.lang))
			})
			c.modificator = &id
		}
	}
}

func (c *charsetoption) OnDisable() func() {
	return func() {
		if c.modificator != nil {
			config.Game.RemoveModificator(*c.modificator)
			c.modificator = nil
		}
	}
}

type charsetprovider interface {
	chars(language.Tag) []api.BasicCharacter
	description(language.Tag) string
}

type inner struct{}

func (i inner) chars(l language.Tag) []api.BasicCharacter {
	switch l {
	case language.German:
		return []api.BasicCharacter{'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', 'ö'}
	default:
		return []api.BasicCharacter{'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l'}
	}
}

func (i inner) description(l language.Tag) string {
	return "Inner Line"
}

type indexfingers struct{}

func (i indexfingers) chars(l language.Tag) []api.BasicCharacter {
	switch l {
	case language.German:
		return []api.BasicCharacter{'r', 'f', 'z', 'u', 'c', 'v', 'b', 'n', 'm'}
	default:
		return []api.BasicCharacter{'r', 'f', 'y', 'u', 'c', 'v', 'b', 'n', 'm'}
	}
}

func (i indexfingers) description(l language.Tag) string {
	return "Index Fingers"
}

type middlefingers struct{}

func (m middlefingers) chars(l language.Tag) []api.BasicCharacter {
	switch l {
	case language.German:
		return []api.BasicCharacter{'e', 'd', 'i', 'k', ','}
	default:
		return []api.BasicCharacter{'e', 'd', 'i', 'k', ','}
	}
}

func (m middlefingers) description(l language.Tag) string {
	return "Middle Fingers"
}

type ringfingers struct{}

func (r ringfingers) chars(l language.Tag) []api.BasicCharacter {
	switch l {
	case language.German:
		return []api.BasicCharacter{'w', 's', 'x', 'o', 'l'}
	default:
		return []api.BasicCharacter{'w', 's', 'x', 'o', 'l'}
	}
}

func (r ringfingers) description(l language.Tag) string {
	return "Ring Fingers"
}

type littlefingers struct{}

func (li littlefingers) chars(l language.Tag) []api.BasicCharacter {
	switch l {
	case language.German:
		return []api.BasicCharacter{'q', 'y', 'ß', 'p', 'ü', 'ö', 'ä', '.'}
	default:
		return []api.BasicCharacter{'q', 'z', 'p', '.'}
	}
}

func (li littlefingers) description(l language.Tag) string {
	return "Little Fingers"
}

type thumbs struct{}

func (t thumbs) chars(l language.Tag) []api.BasicCharacter {
	switch l {
	case language.German:
		return []api.BasicCharacter{' '}
	default:
		return []api.BasicCharacter{' '}
	}
}

func (t thumbs) description(l language.Tag) string {
	return "Thumbs"
}

type shift struct {
	sst         api.StreamSourceType
	modificator *int64
}

func (s *shift) Description() string {
	return "Upper-Case-Letters"
}

func (s *shift) EnabledByDefault() bool {
	return true
}

func (s *shift) OnEnable() func() {
	return func() {
		if s.modificator == nil {
			id := config.Game.AddModificator(s.sst, func(ssd *api.StreamSupplierDescription) {
				ssd.Charset = extend(ssd.Charset, upperCase(ssd.Charset))
			}, true)
			s.modificator = &id
		}
	}
}

func (s *shift) OnDisable() func() {
	return func() {
		if s.modificator != nil {
			config.Game.RemoveModificator(*s.modificator)
			s.modificator = nil
		}
	}
}

func charsetDescription(cs []api.BasicCharacter) (d string) {
	for i, c := range cs {
		if i != 0 {
			d += ", "
		}
		d += "'" + string(c) + "'"
	}
	return
}

func extend(dst, src []api.BasicCharacter) []api.BasicCharacter {
	sort.Slice(dst, func(i, j int) bool {
		return dst[i].Rune() < dst[j].Rune()
	})
	for _, c := range src {
		if !contains(dst, c) {
			dst = append(dst, c)
		}
	}
	return dst
}

func contains(slice []api.BasicCharacter, item api.BasicCharacter) bool {
	i := sort.Search(len(slice), func(i int) bool {
		return slice[i].Rune() >= item.Rune()
	})
	if i < len(slice) && slice[i].Rune() == item.Rune() {
		return true
	}
	return false
}

func upperCase(m []api.BasicCharacter) (copy []api.BasicCharacter) {
	copy = make([]api.BasicCharacter, 0, len(m))
	for _, c := range m {
		u := api.BasicCharacter(unicode.ToUpper(c.Rune()))
		if u.Rune() != c.Rune() {
			copy = append(copy, u)
		}
	}
	return
}
