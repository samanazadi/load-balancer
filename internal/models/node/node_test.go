package node

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func TestSetAlive(t *testing.T) {
	_, tests := CreateFakeNodes()

	for _, test := range tests {
		name := fmt.Sprintf("%sTo%s", AliveToString(test.Alive), AliveToString(test.SetTo))
		t.Run(name, func(t *testing.T) {
			test.Node.SetAlive(test.SetTo)
			if test.Node.alive != test.Want {
				t.Error("Node.IsAlive() = true")
			}
			if test.Node.URL.String() != test.URL.String() {
				t.Error("Node.IsAlive() changed URL")
			}
		})
	}
}

func TestIsAlive(t *testing.T) {
	urlStr := "localhost:8001"
	uu, _ := url.Parse(urlStr)
	node := &Node{
		URL:   uu,
		alive: false,
	}

	if node.IsAlive() {
		t.Error("Node.IsAlive() = true")
	}
	if node.URL.String() != urlStr {
		t.Error("Node.IsAlive() changed URL")
	}

	node = &Node{
		URL:   uu,
		alive: true,
	}

	if !node.IsAlive() {
		t.Error("Node.IsAlive() = false")
	}
	if node.URL.String() != urlStr {
		t.Error("Node.IsAlive() changed URL")
	}
}

func TestNewAlive(t *testing.T) {
	urlStr := "localhost:8001"
	uu, _ := url.Parse(urlStr)
	node := New(uu, true, nil, nil)

	if !node.alive {
		t.Error("node.New(*, true, *, *).alive = false")
	}
	if node.URL.String() != urlStr {
		t.Errorf("node.New(%s, true, *, *).URL = %s", urlStr, node.URL.String())
	}
}

func TestNewDead(t *testing.T) {
	urlStr := "localhost:8001"
	uu, _ := url.Parse(urlStr)
	node := New(uu, false, nil, nil)

	if node.alive {
		t.Error("node.New(*, false, *, *).alive = true")
	}
	if node.URL.String() != urlStr {
		t.Errorf("node.New(%s, false, *, *).URL = %s", urlStr, node.URL.String())
	}
}

func TestContextRetryCount(t *testing.T) {
	r := &http.Request{}
	want := 1
	if got := getRetryCountFromContext(r); got != want {
		t.Errorf("getRetryCountFromContext(request without RetryCount value) = %d, want %d", got, want)
	}

	want = 2
	ctx := context.WithValue(context.Background(), RetryCount, want)
	r = r.WithContext(ctx)
	if got := getRetryCountFromContext(r); got != want {
		t.Errorf("getRetryCountFromContext(request with RetryCount = %d) = %d, want %d", want, got, want)
	}
}

func TestAliveMethods(t *testing.T) {
	urlStr := "localhost:8001"
	uu, _ := url.Parse(urlStr)
	node := New(uu, false, nil, nil)
	node.SetAlive(node.IsAlive()) // no change

	if node.IsAlive() {
		t.Error("Node.IsAlive() = false")
	}
	if node.URL.String() != urlStr {
		t.Error("Node.IsAlive() changed URL")
	}
}
