// Package blocks contains block processing libraries according to
// the Ethereum beacon chain spec.
package blocks

import (
	"context"

	"github.com/pkg/errors"
	"github.com/theQRL/qrysm/beacon-chain/state"
	fieldparams "github.com/theQRL/qrysm/config/fieldparams"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/blocks"
	"github.com/theQRL/qrysm/consensus-types/interfaces"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	enginev1 "github.com/theQRL/qrysm/proto/engine/v1"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
)

// NewGenesisBlock returns the canonical, genesis block for the beacon chain protocol.
func NewGenesisBlock(stateRoot []byte) *qrysmpb.SignedBeaconBlockCapella {
	zeroHash := params.BeaconConfig().ZeroHash[:]
	block := &qrysmpb.SignedBeaconBlockCapella{
		Block: &qrysmpb.BeaconBlockCapella{
			ParentRoot: zeroHash,
			StateRoot:  bytesutil.PadTo(stateRoot, 32),
			Body: &qrysmpb.BeaconBlockBodyCapella{
				RandaoReveal: make([]byte, fieldparams.DilithiumSignatureLength),
				ExecutionData: &qrysmpb.ExecutionData{
					DepositRoot: make([]byte, 32),
					BlockHash:   make([]byte, 32),
				},
				Graffiti: make([]byte, 32),
				SyncAggregate: &qrysmpb.SyncAggregate{
					SyncCommitteeBits: make([]byte, fieldparams.SyncCommitteeLength/8),
				},
				ExecutionPayload: &enginev1.ExecutionPayloadCapella{
					ParentHash:    make([]byte, 32),
					FeeRecipient:  make([]byte, 20),
					StateRoot:     make([]byte, 32),
					ReceiptsRoot:  make([]byte, 32),
					LogsBloom:     make([]byte, 256),
					PrevRandao:    make([]byte, 32),
					BaseFeePerGas: make([]byte, 32),
					BlockHash:     make([]byte, 32),
				},
			},
		},
		Signature: params.BeaconConfig().EmptyDilithiumSignature[:],
	}
	return block
}

var ErrUnrecognizedState = errors.New("unknown underlying type for state.BeaconState value")

func NewGenesisBlockForState(ctx context.Context, st state.BeaconState) (interfaces.ReadOnlySignedBeaconBlock, error) {
	root, err := st.HashTreeRoot(ctx)
	if err != nil {
		return nil, err
	}
	ps := st.ToProto()
	switch ps.(type) {
	case *qrysmpb.BeaconStateCapella:
		return blocks.NewSignedBeaconBlock(&qrysmpb.SignedBeaconBlockCapella{
			Block: &qrysmpb.BeaconBlockCapella{
				ParentRoot: params.BeaconConfig().ZeroHash[:],
				StateRoot:  root[:],
				Body: &qrysmpb.BeaconBlockBodyCapella{
					RandaoReveal: make([]byte, fieldparams.DilithiumSignatureLength),
					ExecutionData: &qrysmpb.ExecutionData{
						DepositRoot: make([]byte, 32),
						BlockHash:   make([]byte, 32),
					},
					Graffiti: make([]byte, 32),
					SyncAggregate: &qrysmpb.SyncAggregate{
						SyncCommitteeBits:       make([]byte, fieldparams.SyncCommitteeLength/8),
						SyncCommitteeSignatures: [][]byte{},
					},
					ExecutionPayload: &enginev1.ExecutionPayloadCapella{
						ParentHash:    make([]byte, 32),
						FeeRecipient:  make([]byte, 20),
						StateRoot:     make([]byte, 32),
						ReceiptsRoot:  make([]byte, 32),
						LogsBloom:     make([]byte, 256),
						PrevRandao:    make([]byte, 32),
						BaseFeePerGas: make([]byte, 32),
						BlockHash:     make([]byte, 32),
						Transactions:  make([][]byte, 0),
						Withdrawals:   make([]*enginev1.Withdrawal, 0),
					},
				},
			},
			Signature: params.BeaconConfig().EmptyDilithiumSignature[:],
		})
	default:
		return nil, ErrUnrecognizedState
	}
}
