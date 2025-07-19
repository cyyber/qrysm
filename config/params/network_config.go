package params

import (
	"time"

	"github.com/mohae/deepcopy"
	"github.com/theQRL/qrysm/consensus-types/primitives"
)

// NetworkConfig defines the spec based network parameters.
type NetworkConfig struct {
	GossipMaxSize                   uint64          `yaml:"GOSSIP_MAX_SIZE"`                    // GossipMaxSize is the maximum allowed size of uncompressed gossip messages.
	MaxChunkSize                    uint64          `yaml:"MAX_CHUNK_SIZE"`                     // MaxChunkSize is the maximum allowed size of uncompressed req/resp chunked responses.
	AttestationSubnetCount          uint64          `yaml:"ATTESTATION_SUBNET_COUNT"`           // AttestationSubnetCount is the number of attestation subnets used in the gossipsub protocol.
	AttestationPropagationSlotRange primitives.Slot `yaml:"ATTESTATION_PROPAGATION_SLOT_RANGE"` // AttestationPropagationSlotRange is the maximum number of slots during which an attestation can be propagated.
	MaxRequestBlocks                uint64          `yaml:"MAX_REQUEST_BLOCKS"`                 // MaxRequestBlocks is the maximum number of blocks in a single request.
	TtfbTimeout                     time.Duration   `yaml:"TTFB_TIMEOUT"`                       // TtfbTimeout is the maximum time to wait for first byte of request response (time-to-first-byte).
	RespTimeout                     time.Duration   `yaml:"RESP_TIMEOUT"`                       // RespTimeout is the maximum time for complete response transfer.
	MaximumGossipClockDisparity     time.Duration   `yaml:"MAXIMUM_GOSSIP_CLOCK_DISPARITY"`     // MaximumGossipClockDisparity is the maximum milliseconds of clock disparity assumed between honest nodes.
	MessageDomainInvalidSnappy      [4]byte         `yaml:"MESSAGE_DOMAIN_INVALID_SNAPPY"`      // MessageDomainInvalidSnappy is the 4-byte domain for gossip message-id isolation of invalid snappy messages.
	MessageDomainValidSnappy        [4]byte         `yaml:"MESSAGE_DOMAIN_VALID_SNAPPY"`        // MessageDomainValidSnappy is the 4-byte domain for gossip message-id isolation of valid snappy messages.

	// DiscoveryV5 Config
	QRL2Key                    string // QRL2Key is the QNR key of the QRL consensus object in an qnr.
	AttSubnetKey               string // AttSubnetKey is the QNR key of the subnet bitfield in the qnr.
	SyncCommsSubnetKey         string // SyncCommsSubnetKey is the QNR key of the sync committee subnet bitfield in the qnr.
	MinimumPeersInSubnetSearch uint64 // PeersInSubnetSearch is the required amount of peers that we need to be able to lookup in a subnet search.

	// Chain Network Config
	ContractDeploymentBlock uint64   // ContractDeploymentBlock is the qrl1 block in which the deposit contract is deployed.
	BootstrapNodes          []string // BootstrapNodes are the addresses of the bootnodes.
}

var networkConfig = mainnetNetworkConfig

// BeaconNetworkConfig returns the current network config for
// the beacon chain.
func BeaconNetworkConfig() *NetworkConfig {
	return networkConfig
}

// OverrideBeaconNetworkConfig will override the network
// config with the added argument.
func OverrideBeaconNetworkConfig(cfg *NetworkConfig) {
	networkConfig = cfg.Copy()
}

// Copy returns Copy of the config object.
func (c *NetworkConfig) Copy() *NetworkConfig {
	config, ok := deepcopy.Copy(*c).(NetworkConfig)
	if !ok {
		config = *networkConfig
	}
	return &config
}
