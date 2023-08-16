package checker

import (
	"fmt"
	"github.com/samanazadi/load-balancer/configs"
	"github.com/samanazadi/load-balancer/pkg/logging"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	TCPType  = "tcp"
	HTTPType = "http"
)

// ConnectionChecker checks for establishment of a connection
type ConnectionChecker interface {
	Check(*url.URL) bool
}

func New(cfg *configs.Config) (ConnectionChecker, error) {
	var chk ConnectionChecker
	switch cfg.Checker.Name {
	case TCPType:
		chk = TCP{
			Timeout: cfg.HealthCheck.Passive.Timeout,
		}
		return chk, nil

	case HTTPType:
		path, keyPhrase, err := HTTPCheckerParamDecode(cfg.Checker.Params)
		if err != nil {
			return nil, err
		}
		chk = HTTP{
			Path:      path,
			KeyPhrase: keyPhrase,
			Timeout:   cfg.HealthCheck.Passive.Timeout,
		}
		return chk, nil
	default:
		return nil, fmt.Errorf("invalid checker: %s", cfg.Checker.Name)
	}
}

// TCP checks by establishing a tcp connection.
type TCP struct {
	Timeout int
}

func (c TCP) Check(url *url.URL) bool {
	timeout := time.Second * time.Duration(c.Timeout)
	conn, err := net.DialTimeout("tcp", url.Host, timeout)
	defer func() {
		if err == nil {
			if connErr := conn.Close(); connErr != nil {
				logging.Logger.Printf("cannot close connection: %s", url.String())
			}
		}
	}()

	return err == nil
}

// HTTP checks by making a get HTTP request
type HTTP struct {
	Path      string
	KeyPhrase string
	Timeout   int
}

func (c HTTP) Check(url *url.URL) bool {
	client := http.Client{
		Timeout: time.Second * time.Duration(c.Timeout),
	}
	res, err := client.Get(url.String() + c.Path)
	if err != nil {
		return false
	}
	body, err := io.ReadAll(res.Body)

	defer func() {
		err = res.Body.Close()
		if err != nil {
			logging.Logger.Printf("cannot close body: %s", url.String())
		}
	}()

	if res.StatusCode > 299 {
		logging.Logger.Printf("HTTP checker failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
		return false
	}

	return strings.Contains(string(body), c.KeyPhrase)
}

func HTTPCheckerParamDecode(m map[string]any) (path, keyPhrase string, err error) {
	path, ok := m["path"].(string)
	if !ok {
		return "", "", fmt.Errorf("http checker invalid path ")
	}
	keyPhrase, ok = m["keyPhrase"].(string)
	if !ok || keyPhrase == "" {
		return "", "", fmt.Errorf("http checker invalid keyPhrase ")
	}
	return
}
