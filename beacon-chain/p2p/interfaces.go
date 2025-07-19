package p2p

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/theQRL/go-zond/p2p/qnr"
	"github.com/theQRL/qrysm/beacon-chain/p2p/encoder"
	"github.com/theQRL/qrysm/beacon-chain/p2p/peers"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/proto/qrysm/v1alpha1/metadata"
	"google.golang.org/protobuf/proto"
)

// P2P represents the full p2p interface composed of all of the sub-interfaces.
type P2P interface {
	Broadcaster
	SetStreamHandler
	PubSubProvider
	PubSubTopicUser
	SenderEncoder
	PeerManager
	ConnectionHandler
	PeersProvider
	MetadataProvider
}

// Broadcaster broadcasts messages to peers over the p2p pubsub protocol.
type Broadcaster interface {
	Broadcast(context.Context, proto.Message) error
	BroadcastAttestation(ctx context.Context, subnet uint64, att *zondpb.Attestation) error
	BroadcastSyncCommitteeMessage(ctx context.Context, subnet uint64, sMsg *zondpb.SyncCommitteeMessage) error
}

// SetStreamHandler configures p2p to handle streams of a certain topic ID.
type SetStreamHandler interface {
	SetStreamHandler(topic string, handler network.StreamHandler)
}

// PubSubTopicUser provides way to join, use and leave PubSub topics.
type PubSubTopicUser interface {
	JoinTopic(topic string, opts ...pubsub.TopicOpt) (*pubsub.Topic, error)
	LeaveTopic(topic string) error
	PublishToTopic(ctx context.Context, topic string, data []byte, opts ...pubsub.PubOpt) error
	SubscribeToTopic(topic string, opts ...pubsub.SubOpt) (*pubsub.Subscription, error)
}

// ConnectionHandler configures p2p to handle connections with a peer.
type ConnectionHandler interface {
	AddConnectionHandler(f func(ctx context.Context, id peer.ID) error,
		j func(ctx context.Context, id peer.ID) error)
	AddDisconnectionHandler(f func(ctx context.Context, id peer.ID) error)
	connmgr.ConnectionGater
}

// SenderEncoder allows sending functionality from libp2p as well as encoding for requests and responses.
type SenderEncoder interface {
	EncodingProvider
	Sender
}

// EncodingProvider provides p2p network encoding.
type EncodingProvider interface {
	Encoding() encoder.NetworkEncoding
}

// PubSubProvider provides the p2p pubsub protocol.
type PubSubProvider interface {
	PubSub() *pubsub.PubSub
}

// PeerManager abstracts some peer management methods from libp2p.
type PeerManager interface {
	Disconnect(peer.ID) error
	PeerID() peer.ID
	Host() host.Host
	QNR() *qnr.Record
	DiscoveryAddresses() ([]multiaddr.Multiaddr, error)
	RefreshQNR()
	FindPeersWithSubnet(ctx context.Context, topic string, subIndex uint64, threshold int) (bool, error)
	AddPingMethod(reqFunc func(ctx context.Context, id peer.ID) error)
}

// Sender abstracts the sending functionality from libp2p.
type Sender interface {
	Send(context.Context, interface{}, string, peer.ID) (network.Stream, error)
}

// PeersProvider abstracts obtaining our current list of known peers status.
type PeersProvider interface {
	Peers() *peers.Status
}

// MetadataProvider returns the metadata related information for the local peer.
type MetadataProvider interface {
	Metadata() metadata.Metadata
	MetadataSeq() uint64
}
