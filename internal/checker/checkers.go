package checker

import (
	"github.com/samanazadi/load-balancer/configs"
	"github.com/samanazadi/load-balancer/internal/logging"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ConnectionChecker checks for establishment of a connection
type ConnectionChecker interface {
	Check(*url.URL) bool
}

type NetChecker struct {
}

// Check checks for a destinations by establishing a tcp connection. Should be thread-safe.
func (c NetChecker) Check(url *url.URL) bool {
	timeout := time.Second * time.Duration(configs.Config.HealthCheck.Passive.Timeout)
	conn, err := net.DialTimeout("tcp", url.Host, timeout)
	defer func() {
		if connErr := conn.Close(); connErr != nil {
			logging.Logger.Printf("Cannot close connection: %s", url.String())
		}
	}()

	return err == nil
}

type HTTPChecker struct {
}

func (c HTTPChecker) Check(url *url.URL) bool {
	path := configs.Config.Checker.Params["httpChecker"].(map[string]any)["path"].(string)
	keyPhrase := configs.Config.Checker.Params["httpChecker"].(map[string]any)["keyPhrase"].(string)

	client := http.Client{
		Timeout: time.Second * time.Duration(configs.Config.HealthCheck.Passive.Timeout),
	}
	res, err := client.Get(url.String() + path)
	if err != nil {
		return false
	}
	body, err := io.ReadAll(res.Body)

	defer func() {
		err = res.Body.Close()
		if err != nil {
			logging.Logger.Printf("Cannot close body: %s", url.String())
		}
	}()

	if res.StatusCode > 299 {
		logging.Logger.Printf("HTTP checker failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
		return false
	}

	return strings.Contains(string(body), keyPhrase)
}
