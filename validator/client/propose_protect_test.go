package client

import (
	"context"
	"testing"

	"github.com/cyyber/qrysm/v4/config/features"
	"github.com/cyyber/qrysm/v4/config/params"
	"github.com/cyyber/qrysm/v4/consensus-types/blocks"
	"github.com/cyyber/qrysm/v4/consensus-types/interfaces"
	"github.com/cyyber/qrysm/v4/consensus-types/primitives"
	ethpb "github.com/cyyber/qrysm/v4/proto/prysm/v1alpha1"
	"github.com/cyyber/qrysm/v4/testing/require"
	"github.com/cyyber/qrysm/v4/testing/util"
	"github.com/golang/mock/gomock"
	dilithium2 "github.com/theQRL/go-qrllib/dilithium"
)

func Test_slashableProposalCheck_PreventsLowerThanMinProposal(t *testing.T) {
	ctx := context.Background()
	validator, _, validatorKey, finish := setup(t)
	defer finish()
	lowestSignedSlot := primitives.Slot(10)
	var pubKeyBytes [dilithium2.CryptoPublicKeyBytes]byte
	copy(pubKeyBytes[:], validatorKey.PublicKey().Marshal())

	// We save a proposal at the lowest signed slot in the DB.
	err := validator.db.SaveProposalHistoryForSlot(ctx, pubKeyBytes, lowestSignedSlot, []byte{1})
	require.NoError(t, err)
	require.NoError(t, err)

	// We expect the same block with a slot lower than the lowest
	// signed slot to fail validation.
	blk := &ethpb.SignedBeaconBlock{
		Block: &ethpb.BeaconBlock{
			Slot:          lowestSignedSlot - 1,
			ProposerIndex: 0,
			Body:          &ethpb.BeaconBlockBody{},
		},
		Signature: params.BeaconConfig().EmptySignature[:],
	}
	wsb, err := blocks.NewSignedBeaconBlock(blk)
	require.NoError(t, err)
	err = validator.slashableProposalCheck(context.Background(), pubKeyBytes, wsb, [32]byte{4})
	require.ErrorContains(t, "could not sign block with slot <= lowest signed", err)

	// We expect the same block with a slot equal to the lowest
	// signed slot to pass validation if signing roots are equal.
	blk = &ethpb.SignedBeaconBlock{
		Block: &ethpb.BeaconBlock{
			Slot:          lowestSignedSlot,
			ProposerIndex: 0,
			Body:          &ethpb.BeaconBlockBody{},
		},
		Signature: params.BeaconConfig().EmptySignature[:],
	}
	wsb, err = blocks.NewSignedBeaconBlock(blk)
	require.NoError(t, err)
	err = validator.slashableProposalCheck(context.Background(), pubKeyBytes, wsb, [32]byte{1})
	require.NoError(t, err)

	// We expect the same block with a slot equal to the lowest
	// signed slot to fail validation if signing roots are different.
	wsb, err = blocks.NewSignedBeaconBlock(blk)
	require.NoError(t, err)
	err = validator.slashableProposalCheck(context.Background(), pubKeyBytes, wsb, [32]byte{4})
	require.ErrorContains(t, failedBlockSignLocalErr, err)

	// We expect the same block with a slot > than the lowest
	// signed slot to pass validation.
	blk = &ethpb.SignedBeaconBlock{
		Block: &ethpb.BeaconBlock{
			Slot:          lowestSignedSlot + 1,
			ProposerIndex: 0,
			Body:          &ethpb.BeaconBlockBody{},
		},
		Signature: params.BeaconConfig().EmptySignature[:],
	}

	wsb, err = blocks.NewSignedBeaconBlock(blk)
	require.NoError(t, err)
	err = validator.slashableProposalCheck(context.Background(), pubKeyBytes, wsb, [32]byte{3})
	require.NoError(t, err)
}

func Test_slashableProposalCheck(t *testing.T) {
	ctx := context.Background()
	config := &features.Flags{
		RemoteSlasherProtection: true,
	}
	reset := features.InitWithReset(config)
	defer reset()
	validator, mocks, validatorKey, finish := setup(t)
	defer finish()

	blk := util.HydrateSignedBeaconBlock(&ethpb.SignedBeaconBlock{
		Block: &ethpb.BeaconBlock{
			Slot:          10,
			ProposerIndex: 0,
			Body:          &ethpb.BeaconBlockBody{},
		},
		Signature: params.BeaconConfig().EmptySignature[:],
	})

	var pubKeyBytes [dilithium2.CryptoPublicKeyBytes]byte
	copy(pubKeyBytes[:], validatorKey.PublicKey().Marshal())

	// We save a proposal at slot 1 as our lowest proposal.
	err := validator.db.SaveProposalHistoryForSlot(ctx, pubKeyBytes, 1, []byte{1})
	require.NoError(t, err)

	// We save a proposal at slot 10 with a dummy signing root.
	dummySigningRoot := [32]byte{1}
	err = validator.db.SaveProposalHistoryForSlot(ctx, pubKeyBytes, 10, dummySigningRoot[:])
	require.NoError(t, err)
	var pubKey [dilithium2.CryptoPublicKeyBytes]byte
	copy(pubKey[:], validatorKey.PublicKey().Marshal())
	sBlock, err := blocks.NewSignedBeaconBlock(blk)
	require.NoError(t, err)
	blockHdr, err := interfaces.SignedBeaconBlockHeaderFromBlockInterface(sBlock)
	require.NoError(t, err)

	mocks.slasherClient.EXPECT().IsSlashableBlock(
		gomock.Any(), // ctx
		blockHdr,
	).Return(&ethpb.ProposerSlashingResponse{}, nil /*err*/)

	// We expect the same block sent out with the same root should not be slasahble.
	err = validator.slashableProposalCheck(context.Background(), pubKey, sBlock, dummySigningRoot)
	require.NoError(t, err)

	// We expect the same block sent out with a different signing root should be slasahble.
	err = validator.slashableProposalCheck(context.Background(), pubKey, sBlock, [32]byte{2})
	require.ErrorContains(t, failedBlockSignLocalErr, err)

	// We save a proposal at slot 11 with a nil signing root.
	blk.Block.Slot = 11
	sBlock, err = blocks.NewSignedBeaconBlock(blk)
	require.NoError(t, err)
	err = validator.db.SaveProposalHistoryForSlot(ctx, pubKeyBytes, blk.Block.Slot, nil)
	require.NoError(t, err)

	// We expect the same block sent out should return slashable error even
	// if we had a nil signing root stored in the database.
	err = validator.slashableProposalCheck(context.Background(), pubKey, sBlock, [32]byte{2})
	require.ErrorContains(t, failedBlockSignLocalErr, err)

	// A block with a different slot for which we do not have a proposing history
	// should not be failing validation.
	blk.Block.Slot = 9
	sBlock, err = blocks.NewSignedBeaconBlock(blk)
	require.NoError(t, err)
	blockHdr, err = interfaces.SignedBeaconBlockHeaderFromBlockInterface(sBlock)
	require.NoError(t, err)
	mocks.slasherClient.EXPECT().IsSlashableBlock(
		gomock.Any(), // ctx
		blockHdr,
	).Return(&ethpb.ProposerSlashingResponse{}, nil /*err*/)
	err = validator.slashableProposalCheck(context.Background(), pubKey, sBlock, [32]byte{3})
	require.NoError(t, err, "Expected allowed block not to throw error")
}

func Test_slashableProposalCheck_RemoteProtection(t *testing.T) {
	config := &features.Flags{
		RemoteSlasherProtection: true,
	}
	reset := features.InitWithReset(config)
	defer reset()
	validator, m, validatorKey, finish := setup(t)
	defer finish()
	var pubKey [dilithium2.CryptoPublicKeyBytes]byte
	copy(pubKey[:], validatorKey.PublicKey().Marshal())

	blk := util.NewBeaconBlock()
	blk.Block.Slot = 10
	sBlock, err := blocks.NewSignedBeaconBlock(blk)
	require.NoError(t, err)
	blockHdr, err := interfaces.SignedBeaconBlockHeaderFromBlockInterface(sBlock)
	require.NoError(t, err)
	m.slasherClient.EXPECT().IsSlashableBlock(
		gomock.Any(), // ctx
		blockHdr,
	).Return(&ethpb.ProposerSlashingResponse{ProposerSlashings: []*ethpb.ProposerSlashing{{}}}, nil /*err*/)

	err = validator.slashableProposalCheck(context.Background(), pubKey, sBlock, [32]byte{2})
	require.ErrorContains(t, failedBlockSignExternalErr, err)

	m.slasherClient.EXPECT().IsSlashableBlock(
		gomock.Any(), // ctx
		blockHdr,
	).Return(&ethpb.ProposerSlashingResponse{}, nil /*err*/)

	err = validator.slashableProposalCheck(context.Background(), pubKey, sBlock, [32]byte{2})
	require.NoError(t, err, "Expected allowed block not to throw error")
}