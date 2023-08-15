package checker

import (
	"fmt"
	"github.com/samanazadi/load-balancer/configs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	// TCP
	cfg := &configs.Config{Checker: configs.Checker{Name: TCPType}}
	chk, err := New(cfg)
	if _, ok := chk.(TCP); !ok {
		t.Errorf("checker.New(TCPType) != TCP")
	}
	if err != nil {
		t.Errorf("checker.New(TCPType) returns error")
	}
	// HTTP
	cfg = &configs.Config{Checker: configs.Checker{Name: HTTPType,
		Params: map[string]any{"path": "path", "keyPhrase": "keyPhrase"}}}
	chk, err = New(cfg)
	if _, ok := chk.(HTTP); !ok {
		t.Errorf("checker.New(HTTPType) != HTTP")
	}
	if err != nil {
		t.Errorf("checker.New(HTTPType) returns error")
	}
	// invalid type
	cfg = &configs.Config{Checker: configs.Checker{Name: "invalid"}}
	chk, err = New(cfg)
	if err == nil {
		t.Errorf("checker.New(HTTPType) doesn't return error")
	}
}

func TestTCPCheckAvailableServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}))
	hc := TCP{
		Timeout: 1,
	}

	u, _ := url.Parse(server.URL)
	if got := hc.Check(u); !got {
		t.Errorf("TCP{timeout: %d} should have succeeded", 1)
	}
}

func TestTCPCheckUnavailableServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}))
	hc := TCP{
		Timeout: 1,
	}

	u, _ := url.Parse(server.URL + "1")
	if got := hc.Check(u); got {
		t.Errorf("TCP{timeout: %d} should have failed", 1)
	}
}

func TestHTTPCheckAvailableServer(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		keyPhrase   string
		timeout     int
		serverPath  string
		serverResp  string
		serverDelay int
		want        bool
	}{
		{
			name:        "AllCorrect",
			path:        "/ping",
			keyPhrase:   "pong",
			timeout:     10,
			serverPath:  "/ping",
			serverResp:  "pong",
			serverDelay: 0,
			want:        true,
		},
		{
			name:        "IncorrectPath",
			path:        "/test",
			keyPhrase:   "pong",
			timeout:     10,
			serverPath:  "/ping",
			serverResp:  "pong",
			serverDelay: 0,
			want:        false,
		},
		{
			name:        "IncorrectKeyPhrase",
			path:        "/ping",
			keyPhrase:   "key",
			timeout:     10,
			serverPath:  "/ping",
			serverResp:  "pong",
			serverDelay: 0,
			want:        false,
		},
		{
			name:        "ExceedingTimeout",
			path:        "/ping",
			keyPhrase:   "key",
			timeout:     1,
			serverPath:  "/ping",
			serverResp:  "pong",
			serverDelay: 2,
			want:        false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				if r.URL.Path == test.serverPath {
					time.Sleep(time.Duration(test.serverDelay) * time.Second)
					fmt.Fprint(rw, test.serverResp)
				}
			}))
			defer server.Close()
			hc := HTTP{
				Path:      test.path,
				KeyPhrase: test.keyPhrase,
				Timeout:   test.timeout,
			}
			u, _ := url.Parse(server.URL)

			if got := hc.Check(u); got != test.want {
				var shouldFail string
				if test.want {
					shouldFail = "succeeded"
				} else {
					shouldFail = "failed"
				}

				t.Errorf(
					"HTTP{path: %s, keyPhrase: %s, timeout: %d}, Server{path: %s, response: %s, delay: %ds} should have %s",
					test.path, test.keyPhrase, test.timeout, test.serverPath, test.serverResp, test.serverDelay, shouldFail)
			}
		})
	}
}

func TestHTTPCheckUnavailableServer(t *testing.T) {
	hc := HTTP{
		Path:      "/ping",
		KeyPhrase: "pong",
		Timeout:   1,
	}
	u, _ := url.Parse("unavailable")
	if got := hc.Check(u); got {
		t.Errorf("HTTP.Check(unavailable server) = %t", got)
	}
}

func TestHTTPCheckerParamDecode(t *testing.T) {
	params := map[string]any{
		"path":      "some-path",
		"keyPhrase": "key",
	}
	path, keyPhrase := HTTPCheckerParamDecode(params)

	if path != "some-path" {
		t.Errorf("checker.HTTPCheckerParamDecode(map[path=some-path]).path = %s", path)
	}
	if keyPhrase != "key" {
		t.Errorf("checker.HTTPCheckerParamDecode(map[keyPhrase=key]).keyPhrase = %s", keyPhrase)
	}
}
