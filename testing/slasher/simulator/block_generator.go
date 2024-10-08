package simulator

import (
	"context"

	"github.com/theQRL/qrysm/beacon-chain/core/signing"
	"github.com/theQRL/qrysm/beacon-chain/state"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/crypto/dilithium"
	"github.com/theQRL/qrysm/crypto/rand"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
)

func (s *Simulator) generateBlockHeadersForSlot(
	ctx context.Context, slot primitives.Slot,
) ([]*zondpb.SignedBeaconBlockHeader, []*zondpb.ProposerSlashing, error) {
	blocks := make([]*zondpb.SignedBeaconBlockHeader, 0)
	slashings := make([]*zondpb.ProposerSlashing, 0)
	proposer := rand.NewGenerator().Uint64() % s.srvConfig.Params.NumValidators

	var parentRoot [32]byte
	beaconState, err := s.srvConfig.StateGen.StateByRoot(ctx, parentRoot)
	if err != nil {
		return nil, nil, err
	}
	block := &zondpb.SignedBeaconBlockHeader{
		Header: &zondpb.BeaconBlockHeader{
			Slot:          slot,
			ProposerIndex: primitives.ValidatorIndex(proposer),
			ParentRoot:    bytesutil.PadTo([]byte{}, 32),
			StateRoot:     bytesutil.PadTo([]byte{}, 32),
			BodyRoot:      bytesutil.PadTo([]byte("good block"), 32),
		},
	}
	sig, err := s.signBlockHeader(beaconState, block)
	if err != nil {
		return nil, nil, err
	}
	block.Signature = sig.Marshal()

	blocks = append(blocks, block)
	if rand.NewGenerator().Float64() < s.srvConfig.Params.ProposerSlashingProbab {
		log.WithField("proposerIndex", proposer).Infof("Slashable block made")
		slashableBlock := &zondpb.SignedBeaconBlockHeader{
			Header: &zondpb.BeaconBlockHeader{
				Slot:          slot,
				ProposerIndex: primitives.ValidatorIndex(proposer),
				ParentRoot:    bytesutil.PadTo([]byte{}, 32),
				StateRoot:     bytesutil.PadTo([]byte{}, 32),
				BodyRoot:      bytesutil.PadTo([]byte("bad block"), 32),
			},
			Signature: sig.Marshal(),
		}
		sig, err = s.signBlockHeader(beaconState, slashableBlock)
		if err != nil {
			return nil, nil, err
		}
		slashableBlock.Signature = sig.Marshal()

		blocks = append(blocks, slashableBlock)
		slashings = append(slashings, &zondpb.ProposerSlashing{
			Header_1: block,
			Header_2: slashableBlock,
		})
	}
	return blocks, slashings, nil
}

func (s *Simulator) signBlockHeader(
	beaconState state.BeaconState,
	header *zondpb.SignedBeaconBlockHeader,
) (dilithium.Signature, error) {
	domain, err := signing.Domain(
		beaconState.Fork(),
		0,
		params.BeaconConfig().DomainBeaconProposer,
		beaconState.GenesisValidatorsRoot(),
	)
	if err != nil {
		return nil, err
	}
	htr, err := header.Header.HashTreeRoot()
	if err != nil {
		return nil, err
	}
	container := &zondpb.SigningData{
		ObjectRoot: htr[:],
		Domain:     domain,
	}
	signingRoot, err := container.HashTreeRoot()
	if err != nil {
		return nil, err
	}
	validatorPrivKey := s.srvConfig.PrivateKeysByValidatorIndex[header.Header.ProposerIndex]
	return validatorPrivKey.Sign(signingRoot[:]), nil
}
