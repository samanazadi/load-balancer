package integration

import (
	"fmt"
	"github.com/samanazadi/load-balancer/configs"
	"github.com/samanazadi/load-balancer/internal/algorithm"
	"github.com/samanazadi/load-balancer/internal/app"
	"github.com/samanazadi/load-balancer/internal/checker"
	"github.com/samanazadi/load-balancer/pkg/logging"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

type TestingLogger struct {
	t *testing.T
}

func (t *TestingLogger) Print(v ...any) {
	t.t.Log(v...)
}

func (t *TestingLogger) Printf(format string, v ...any) {
	t.t.Logf(format, v...)
}

func (t *TestingLogger) Println(v ...any) {
	t.t.Log(v...)
}

func (t *TestingLogger) Fatal(v ...any) {
	t.t.Fatal(v...)
}

func (t *TestingLogger) Fatalf(format string, v ...any) {
	t.t.Fatalf(format, v...)
}

func (t *TestingLogger) Fatalln(v ...any) {
	t.t.Fatal(v...)
}

func (t *TestingLogger) Panic(v ...any) {
	t.t.Fatal(v...)
}

func (t *TestingLogger) Panicf(format string, v ...any) {
	t.t.Fatalf(format, v...)
}

func (t *TestingLogger) Panicln(v ...any) {
	t.t.Fatal(v...)
}

const N = 4 // number of nodes

func TestBigBang(t *testing.T) {
	t.Log("integration test started")

	// testing logger
	logging.Logger = &TestingLogger{t: t}

	// create mock servers
	var mocks []*httptest.Server
	for i := 0; i < N; i++ {
		mocks = append(mocks, CreateTestServer(i+1))
	}

	// config
	cfg, err := configs.New()
	if err != nil {
		t.Fatal(err)
	}

	// replace real nodes with mock nodes
	cfg.Nodes = make([]string, 0, len(mocks))
	for _, mock := range mocks {
		cfg.Nodes = append(cfg.Nodes, mock.URL)
	}

	// checker
	t.Logf("checker: %s", cfg.Checker.Name)
	chk, err := checker.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// algorithm
	t.Logf("algorithm: %s", cfg.Algorithm.Name)
	alg, err := algorithm.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// load balancer
	lb := app.New(cfg, chk, alg)
	t.Log("load balancer created")

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Port),
		Handler: lb,
	}

	t.Logf("load balancer started at port %d", cfg.Port)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			t.Fatalf("cannot start load balancer: %s", err.Error())
		}
	}()

	// test server pool size
	t.Run("ServerPoolSize", func(t *testing.T) {
		want := len(mocks)
		got := len(lb.ServerPool.Nodes)
		if got != want {
			t.Fatalf("len(ServerPool.Nodes) = %d, want = %d", got, want)
		}
	})

	// test next eligible node
	t.Run("NextEligibleNode", func(t *testing.T) {
		for i := 1; i < 1000*N; i++ {
			r := httptest.NewRequest("GET", "localhost:"+strconv.Itoa(cfg.Port), nil)
			n := lb.Algorithm.GetNextEligibleNode(r)
			if n == nil {
				t.Fatal("GetNextEligibleNode() = nil but alive nodes are available")
			}
		}
	})

	// passive health check: tcp
	cfg.Checker.Name = checker.TCPType
	cfg.HealthCheck.Passive.Timeout = 3
	chk, err = checker.New(cfg)
	if err != nil {
		t.Fatalf("cannot create checker: %s", err)
	}

	lb.ServerPool.ConnectionChecker = chk
	mocks[1].Close()              // shut down node 1
	lb.StartPassiveHealthCheck(1) // run health check to mark node 1 as dead
	time.Sleep(2 * time.Second)

	found := false
	for _, n := range lb.ServerPool.Nodes {
		if n.URL.String() == mocks[1].URL {
			found = true
			if n.IsAlive() {
				t.Fatalf("passive health check didn't mark node as  dead")
			}
		}
	}
	if !found {
		t.Fatalf("dead node not found in server pool")
	}

	// passive health check: http
	cfg.Checker.Name = checker.HTTPType
	cfg.HealthCheck.Passive.Timeout = 1
	cfg.Checker.Params = map[string]any{
		"path":      "/ping",
		"keyPhrase": "pong",
	}
	chk, err = checker.New(cfg)
	if err != nil {
		t.Fatalf("cannot create checker: %s", err)
	}

	lb.ServerPool.ConnectionChecker = chk
	mocks[2].Close()              // shut down node 2
	lb.StartPassiveHealthCheck(1) // run health check to mark node 2 as dead
	time.Sleep(2 * time.Second)

	found = false
	for _, n := range lb.ServerPool.Nodes {
		if n.URL.String() == mocks[2].URL {
			found = true
			if n.IsAlive() {
				t.Fatalf("passive health check didn't mark node as  dead")
			}
		}
	}
	if !found {
		t.Fatalf("dead node not found in server pool")
	}

	// clean up
	for _, mock := range mocks {
		mock.Close()
	}
	t.Log("integration test completed")
}

func CreateTestServer(n int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ping" {
			fmt.Fprint(rw, strings.NewReader("pong"))
		}
		fmt.Fprint(rw, strings.NewReader(fmt.Sprintf("mock server #%d", n)))
	}))
}
