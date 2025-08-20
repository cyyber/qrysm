package testing

import (
	"context"
	"errors"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/theQRL/go-zond/p2p/qnr"
)

// MockPeerManager is mock of the PeerManager interface.
type MockPeerManager struct {
	Qnr               *qnr.Record
	PID               peer.ID
	BHost             host.Host
	DiscoveryAddr     []multiaddr.Multiaddr
	FailDiscoveryAddr bool
}

// Disconnect .
func (_ *MockPeerManager) Disconnect(peer.ID) error {
	return nil
}

// PeerID .
func (m *MockPeerManager) PeerID() peer.ID {
	return m.PID
}

// Host .
func (m *MockPeerManager) Host() host.Host {
	return m.BHost
}

// QNR .
func (m MockPeerManager) QNR() *qnr.Record {
	return m.Qnr
}

// DiscoveryAddresses .
func (m MockPeerManager) DiscoveryAddresses() ([]multiaddr.Multiaddr, error) {
	if m.FailDiscoveryAddr {
		return nil, errors.New("fail")
	}
	return m.DiscoveryAddr, nil
}

// RefreshQNR .
func (_ MockPeerManager) RefreshQNR() {}

// FindPeersWithSubnet .
func (_ MockPeerManager) FindPeersWithSubnet(_ context.Context, _ string, _ uint64, _ int) (bool, error) {
	return true, nil
}

// AddPingMethod .
func (_ MockPeerManager) AddPingMethod(_ func(ctx context.Context, id peer.ID) error) {}
