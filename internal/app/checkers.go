package app

import (
	"github.com/samanazadi/load-balancer/pkg/logging"
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

// TCPChecker checks by establishing a tcp connection.
type TCPChecker struct {
	Timeout int
}

func (c TCPChecker) Check(url *url.URL) bool {
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

// HTTPChecker checks by making a get HTTP request
type HTTPChecker struct {
	Path      string
	KeyPhrase string
	Timeout   int
}

func (c HTTPChecker) Check(url *url.URL) bool {
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

func HTTPCheckerParamDecode(m map[string]any) (path, keyPhrase string) {
	path = m["path"].(string)
	keyPhrase = m["keyPhrase"].(string)
	return
}
