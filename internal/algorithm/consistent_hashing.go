package algorithm

import (
	"fmt"
	"github.com/samanazadi/load-balancer/configs"
	"github.com/samanazadi/load-balancer/internal/models/node"
	"hash/crc32"
	"net/http"
	"sort"
	"strconv"
)

const CRC32Type = "crc32"

type HashFunc func([]byte) uint32

type ConsistentHashing struct {
	Replicas    int                // replicas count
	HashFunc    HashFunc           // hash function
	VNodes      []int              // sorted virtual nodes
	ActualNodes map[int]*node.Node // vnode to node
	Nodes       []*node.Node       // original nodes
}

func (ch *ConsistentHashing) GetNextEligibleNode(r *http.Request) *node.Node {
	requestHash := int(ch.HashFunc([]byte(r.RemoteAddr)))
	index := sort.Search(len(ch.VNodes), func(i int) bool { return ch.VNodes[i] >= requestHash }) // binary search

	if index == len(ch.VNodes) { // spun a complete round
		index = 0
	}

	// check for dead node
	an := ch.ActualNodes[ch.VNodes[index]]
	on, i := ch.getOriginalNode(an)
	if on.IsAlive() {
		return on
	}
	// get next node
	next := i + 1
	last := next + len(ch.Nodes)
	for i := next; i < last; i++ {
		index := i % len(ch.Nodes)
		if ch.Nodes[index].IsAlive() {
			return ch.Nodes[index]
		}
	}
	return nil // no available node
}

func (ch *ConsistentHashing) getOriginalNode(n *node.Node) (*node.Node, int) {
	for i, nn := range ch.Nodes {
		if nn.URL.String() == n.URL.String() {
			return nn, i
		}
	}
	return nil, 0
}

func (ch *ConsistentHashing) SetNodes(nodes []*node.Node) {
	ch.Nodes = nodes
	// vnodes and actual nodes
	ch.ActualNodes = make(map[int]*node.Node)
	for _, n := range nodes {
		for i := 0; i < ch.Replicas; i++ {
			vn := int(ch.HashFunc([]byte(strconv.Itoa(i) + n.URL.String()))) // vnode
			ch.VNodes = append(ch.VNodes, vn)
			ch.ActualNodes[vn] = n
		}
	}
	sort.Ints(ch.VNodes)
}

func ConsistentHashingParamDecode(m map[string]any) (replicas int, hashFunc string, err error) {
	freplicas, ok := m["replicas"].(float64)
	if !ok {
		return 0, "", fmt.Errorf("consistent hasing invalid replicas ")
	}
	replicas = int(freplicas)

	hashFunc, ok = m["hashFunc"].(string)
	if !ok || hashFunc == "" {
		return 0, "", fmt.Errorf("consistent hasing invalid hashFunc ")
	}
	return
}

func NewConsistentHashing(cfg *configs.Config) (Algorithm, error) {
	replicas, hashFunc, err := ConsistentHashingParamDecode(cfg.Algorithm.Params)
	if err != nil {
		return nil, err
	}

	// replicas
	if replicas < 1 {
		return nil, fmt.Errorf("invalid replicas: %d", replicas)
	}

	// hash function
	var hf HashFunc
	switch hashFunc {
	case CRC32Type:
		hf = crc32.ChecksumIEEE

	default:
		return nil, fmt.Errorf("invalid hashFunc: %s", hashFunc)
	}
	return &ConsistentHashing{
		Replicas: replicas,
		HashFunc: hf,
	}, nil
}
