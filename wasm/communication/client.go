package communication

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gopherjs/websocket"
	"github.com/theMomax/notypo-backend/api"
	"github.com/theMomax/notypo-frontend/wasm/comparison"
	"github.com/theMomax/notypo-frontend/wasm/errors"
)

// errors
var (
	ErrServerConnectionFailed   = errors.New("the connection to the api-backend failed", errors.Critical, errors.Server)
	ErrTestBuild                = errors.New("the backend-api is only a test- or development-build", errors.Warning, errors.Server)
	ErrUnexpectedResponseFormat = errors.New("the returned response doesn't match the expected type", errors.Critical, errors.Server)
	ErrUnexpectedArgumentFormat = errors.New("the given arguments don't match the requirements", errors.Critical)
	ErrStreamNotImplemented     = errors.New("the server doesn't know the reqested stream-type", errors.Critical, errors.Server)
	ErrStreamNotFound           = errors.New("the server couldn't find a stream with the given id", errors.Critical, errors.Server)
	ErrIllegalConfiguration     = errors.New("the configuration is not valid for this type of stream", errors.Critical)
)

// Version requests the backend-api's version-information
func Version(baseURL *url.URL) (*api.VersionResponse, errors.Error) {
	resp, err := http.Get(baseURL.String() + api.PathVersion)
	if err != nil {
		return nil, ErrServerConnectionFailed.Append(err.Error())
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		var v api.VersionResponse
		err = json.NewDecoder(resp.Body).Decode(&v)
		if err != nil {
			return nil, ErrUnexpectedResponseFormat.Append(err.Error())
		}
		return &v, nil
	case 503:
		return nil, ErrTestBuild
	default:
		return nil, ErrUnexpectedResponseFormat.Append(resp.Status)
	}
}

// StreamOptions requests the available StreamTypes
func StreamOptions(baseURL *url.URL) (api.StreamOptionsResponse, errors.Error) {
	req, err := http.NewRequest("GET", baseURL.String()+api.PathStreamOptions, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, ErrServerConnectionFailed.Append(err.Error())
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		var s api.StreamOptionsResponse
		err = json.NewDecoder(resp.Body).Decode(&s)
		if err != nil {
			return nil, ErrUnexpectedResponseFormat.Append(err.Error())
		}
		return s, nil
	default:
		return nil, ErrUnexpectedResponseFormat.Append(resp.Status)
	}
}

// CreateRandomStream creates a Stream based on the given description and
// returns its StreamID
func CreateRandomStream(baseURL *url.URL, description *api.StreamSupplierDescription) (*int64, errors.Error) {
	b, err := json.Marshal(description)
	if err != nil {
		return nil, ErrUnexpectedArgumentFormat.Append(err.Error())
	}
	resp, err := http.Post(baseURL.String()+api.PathCreateStream, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, ErrServerConnectionFailed.Append(err.Error())
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		var streamID int64
		json.NewDecoder(resp.Body).Decode(&streamID)
		if err != nil {
			return nil, ErrUnexpectedResponseFormat.Append(err.Error())
		}
		return &streamID, nil
	case 400:
		return nil, ErrIllegalConfiguration
	case 501:
		return nil, ErrStreamNotImplemented
	default:
		return nil, ErrUnexpectedResponseFormat.Append(resp.Status)
	}
}

// OpenStreamConnection opens a connection to the given Stream and returns its
// StreamConnectionID
func OpenStreamConnection(baseURL *url.URL, streamID int64) (*int64, errors.Error) {
	resp, err := http.Get(baseURL.String() + strings.TrimSuffix(api.PathOpenStreamConnection, "{id}") + strconv.FormatInt(streamID, 10))
	if err != nil {
		return nil, ErrServerConnectionFailed.Append(err.Error())
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		var streamConnectionID int64
		json.NewDecoder(resp.Body).Decode(&streamConnectionID)
		if err != nil {
			return nil, ErrUnexpectedResponseFormat.Append(err.Error())
		}
		return &streamConnectionID, nil
	case 404:
		return nil, ErrStreamNotFound
	default:
		return nil, ErrUnexpectedResponseFormat.Append(resp.Status)
	}
}

// ReadStreamConnection returns a channel of Characters with the given buffer
// and spawns a new goroutine, which requests the Stream's content and pipes it
// into the returned channel. This process ends, when the connection is closed
// by the server, or when the goroutine receives a signal from the given done
// channel
func ReadStreamConnection(baseURL *url.URL, buffer uint, streamConnectionID int64, done <-chan bool) (<-chan comparison.Character, errors.Error) {
	mod := make(chan comparison.Character, int(buffer))
	var c net.Conn
	u := url.URL{Scheme: "ws", Host: baseURL.Host, Path: strings.TrimSuffix(api.PathEstablishWebsocketToStream, "{id}") + strconv.FormatInt(streamConnectionID, 10)}
	c, err := websocket.Dial(u.String())
	if err != nil {
		return nil, ErrServerConnectionFailed.Append(err.Error())
	}

	go func() {
		charRequests := make(chan uint, 10)
		reqLimit := buffer / 10
		if reqLimit == 0 {
			reqLimit = 1
		}
		charRequests <- reqLimit
		// request characters
		go func() {
			reqAmount := uint(0)
			for {
				select {
				case <-done:
					c.Close()
					close(mod)
					close(charRequests)
					return
				case n := <-charRequests:
					reqAmount += n
					if reqAmount >= reqLimit {
						b, err := json.Marshal(&reqAmount)
						if err != nil {
							c.Close()
							return
						}
						_, err = c.Write(b)
						if err != nil {
							c.Close()
							return
						}
						reqAmount = 0
					}
				}
			}
		}()
		// process stream
		done := false
	outer:
		for !done {
			var char api.BasicCharacter
			b := make([]byte, 100)
			n, err := c.Read(b)
			if err != nil {
				c.Close()
				break outer
			}
			err = json.Unmarshal(b[:n], &char)
			if err != nil {
				c.Close()
				break outer
			}
			// triggered, by write on closed channel, when request-routine
			// receives from done channel and returns instread of second close()
			defer func() {
				p := recover()
				if p != nil {
					done = true
				}
			}()
			mod <- char
			charRequests <- 1
		}
		close(charRequests)
		close(mod)
	}()
	return mod, nil
}

// CloseStreamConnection closes the connection with the given id
func CloseStreamConnection(baseURL *url.URL, streamConnectionID int64) errors.Error {
	req, err := http.NewRequest("DELETE", baseURL.String()+strings.TrimSuffix(api.PathCloseStreamConnection, "{id}")+strconv.FormatInt(streamConnectionID, 10), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ErrServerConnectionFailed.Append(err.Error())
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		return nil
	default:
		return ErrUnexpectedResponseFormat.Append(resp.Status)
	}
}
