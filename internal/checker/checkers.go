package checker

import (
	"github.com/samanazadi/load-balancer/internal/logging"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	TCP  = "tcp"
	HTTP = "http"
)

// ConnectionChecker checks for establishment of a connection
type ConnectionChecker interface {
	Check(*url.URL) bool
}

type TCPChecker struct {
	timeout int
}

// Check checks for a destinations by establishing a tcp connection. Should be thread-safe.
func (c TCPChecker) Check(url *url.URL) bool {
	timeout := time.Second * time.Duration(c.timeout)
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

type HTTPChecker struct {
	path      string
	keyPhrase string
	timeout   int
}

func (c HTTPChecker) Check(url *url.URL) bool {
	client := http.Client{
		Timeout: time.Second * time.Duration(c.timeout),
	}
	res, err := client.Get(url.String() + c.path)
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

	return strings.Contains(string(body), c.keyPhrase)
}
