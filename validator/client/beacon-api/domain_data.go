package beacon_api

import (
	"context"

	"github.com/pkg/errors"
	"github.com/theQRL/go-zond/common/hexutil"
	"github.com/theQRL/qrysm/beacon-chain/core/signing"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/network/forks"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
)

func (c beaconApiValidatorClient) getDomainData(ctx context.Context, epoch primitives.Epoch, domainType [4]byte) (*zondpb.DomainResponse, error) {
	// Get the fork version from the given epoch
	fork, err := forks.Fork(epoch)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get fork version for epoch %d", epoch)
	}

	// Get the genesis validator root
	genesis, _, err := c.genesisProvider.GetGenesis(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get genesis info")
	}

	if !validRoot(genesis.GenesisValidatorsRoot) {
		return nil, errors.Errorf("invalid genesis validators root: %s", genesis.GenesisValidatorsRoot)
	}

	genesisValidatorRoot, err := hexutil.Decode(genesis.GenesisValidatorsRoot)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode genesis validators root")
	}

	signatureDomain, err := signing.Domain(fork, epoch, domainType, genesisValidatorRoot)
	if err != nil {
		return nil, errors.Wrap(err, "failed to compute signature domain")
	}

	return &zondpb.DomainResponse{SignatureDomain: signatureDomain}, nil
}
