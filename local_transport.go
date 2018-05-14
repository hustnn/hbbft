package hbbft

import (
	"fmt"
	"sync"
)

// LocalTransport implements a local Transport. This is used to test hbbft
// without going over the network.
type LocalTransport struct {
	lock      sync.RWMutex
	peers     map[string]*LocalTransport
	consumeCh chan RPC
	addr      string
}

// NewLocalTransport returns a new LocalTransport.
func NewLocalTransport(addr string) *LocalTransport {
	return &LocalTransport{
		peers:     make(map[string]*LocalTransport),
		consumeCh: make(chan RPC),
		addr:      addr,
	}
}

// Consume implements the Transport interface.
func (t *LocalTransport) Consume() <-chan RPC {
	return t.consumeCh
}

// SendProofRequests implements the Transport interface.
func (t *LocalTransport) SendProofRequests(id uint64, reqs []*ProofRequest) error {
	i := 0
	for addr := range t.peers {
		if err := t.makeRPC(id, addr, reqs[i]); err != nil {
			return err
		}
		i++
	}
	return nil
}

// Broadcast implements the Transport interface.
func (t *LocalTransport) Broadcast(id uint64, v interface{}) error {
	for addr := range t.peers {
		if err := t.makeRPC(id, addr, v); err != nil {
			return err
		}
	}
	return nil
}

// Connect implements the Transport interface.
func (t *LocalTransport) Connect(addr string, tr Transport) {
	trans := tr.(*LocalTransport)
	t.lock.Lock()
	defer t.lock.Unlock()
	t.peers[addr] = trans
}

// Addr implements the Transport interface.
func (t *LocalTransport) Addr() string {
	return t.addr
}

func (t *LocalTransport) makeRPC(id uint64, addr string, payload interface{}) error {
	t.lock.RLock()
	peer, ok := t.peers[addr]
	t.lock.RUnlock()

	if !ok {
		return fmt.Errorf("failed to connect with %s", addr)
	}
	peer.consumeCh <- RPC{
		NodeID:  id,
		Payload: payload,
	}
	return nil
}