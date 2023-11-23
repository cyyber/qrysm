package beacon_api

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/theQRL/go-zond/common/hexutil"
	"github.com/theQRL/qrysm/v4/beacon-chain/rpc/apimiddleware"
	"github.com/theQRL/qrysm/v4/consensus-types/primitives"
	enginev1 "github.com/theQRL/qrysm/v4/proto/engine/v1"
	zondpb "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
)

func convertProposerSlashingsToProto(jsonProposerSlashings []*apimiddleware.ProposerSlashingJson) ([]*zondpb.ProposerSlashing, error) {
	proposerSlashings := make([]*zondpb.ProposerSlashing, len(jsonProposerSlashings))

	for index, jsonProposerSlashing := range jsonProposerSlashings {
		if jsonProposerSlashing == nil {
			return nil, errors.Errorf("proposer slashing at index `%d` is nil", index)
		}

		header1, err := convertProposerSlashingSignedHeaderToProto(jsonProposerSlashing.Header_1)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get proposer header 1")
		}

		header2, err := convertProposerSlashingSignedHeaderToProto(jsonProposerSlashing.Header_2)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get proposer header 2")
		}

		proposerSlashings[index] = &zondpb.ProposerSlashing{
			Header_1: header1,
			Header_2: header2,
		}
	}

	return proposerSlashings, nil
}

func convertProposerSlashingSignedHeaderToProto(signedHeader *apimiddleware.SignedBeaconBlockHeaderJson) (*zondpb.SignedBeaconBlockHeader, error) {
	if signedHeader == nil {
		return nil, errors.New("signed header is nil")
	}

	if signedHeader.Header == nil {
		return nil, errors.New("header is nil")
	}

	slot, err := strconv.ParseUint(signedHeader.Header.Slot, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse header slot `%s`", signedHeader.Header.Slot)
	}

	proposerIndex, err := strconv.ParseUint(signedHeader.Header.ProposerIndex, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse header proposer index `%s`", signedHeader.Header.ProposerIndex)
	}

	parentRoot, err := hexutil.Decode(signedHeader.Header.ParentRoot)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode header parent root `%s`", signedHeader.Header.ParentRoot)
	}

	stateRoot, err := hexutil.Decode(signedHeader.Header.StateRoot)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode header state root `%s`", signedHeader.Header.StateRoot)
	}

	bodyRoot, err := hexutil.Decode(signedHeader.Header.BodyRoot)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode header body root `%s`", signedHeader.Header.BodyRoot)
	}

	signature, err := hexutil.Decode(signedHeader.Signature)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode signature `%s`", signedHeader.Signature)
	}

	return &zondpb.SignedBeaconBlockHeader{
		Header: &zondpb.BeaconBlockHeader{
			Slot:          primitives.Slot(slot),
			ProposerIndex: primitives.ValidatorIndex(proposerIndex),
			ParentRoot:    parentRoot,
			StateRoot:     stateRoot,
			BodyRoot:      bodyRoot,
		},
		Signature: signature,
	}, nil
}

func convertAttesterSlashingsToProto(jsonAttesterSlashings []*apimiddleware.AttesterSlashingJson) ([]*zondpb.AttesterSlashing, error) {
	attesterSlashings := make([]*zondpb.AttesterSlashing, len(jsonAttesterSlashings))

	for index, jsonAttesterSlashing := range jsonAttesterSlashings {
		if jsonAttesterSlashing == nil {
			return nil, errors.Errorf("attester slashing at index `%d` is nil", index)
		}

		attestation1, err := convertIndexedAttestationToProto(jsonAttesterSlashing.Attestation_1)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get attestation 1")
		}

		attestation2, err := convertIndexedAttestationToProto(jsonAttesterSlashing.Attestation_2)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get attestation 2")
		}

		attesterSlashings[index] = &zondpb.AttesterSlashing{
			Attestation_1: attestation1,
			Attestation_2: attestation2,
		}
	}

	return attesterSlashings, nil
}

func convertIndexedAttestationToProto(jsonAttestation *apimiddleware.IndexedAttestationJson) (*zondpb.IndexedAttestation, error) {
	if jsonAttestation == nil {
		return nil, errors.New("indexed attestation is nil")
	}

	attestingIndices := make([]uint64, len(jsonAttestation.AttestingIndices))

	for index, jsonAttestingIndex := range jsonAttestation.AttestingIndices {
		attestingIndex, err := strconv.ParseUint(jsonAttestingIndex, 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse attesting index `%s`", jsonAttestingIndex)
		}

		attestingIndices[index] = attestingIndex
	}

	signatures := make([][]byte, len(jsonAttestation.Signatures))
	var err error
	for i, sig := range jsonAttestation.Signatures {
		signatures[i], err = hexutil.Decode(sig)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode attestation signature `%s`", jsonAttestation.Signatures[i])
		}
	}

	attestationData, err := convertAttestationDataToProto(jsonAttestation.Data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get attestation data")
	}

	return &zondpb.IndexedAttestation{
		AttestingIndices: attestingIndices,
		Data:             attestationData,
		Signatures:       signatures,
	}, nil
}

func convertCheckpointToProto(jsonCheckpoint *apimiddleware.CheckpointJson) (*zondpb.Checkpoint, error) {
	if jsonCheckpoint == nil {
		return nil, errors.New("checkpoint is nil")
	}

	epoch, err := strconv.ParseUint(jsonCheckpoint.Epoch, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse checkpoint epoch `%s`", jsonCheckpoint.Epoch)
	}

	root, err := hexutil.Decode(jsonCheckpoint.Root)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode checkpoint root `%s`", jsonCheckpoint.Root)
	}

	return &zondpb.Checkpoint{
		Epoch: primitives.Epoch(epoch),
		Root:  root,
	}, nil
}

func convertAttestationToProto(jsonAttestation *apimiddleware.AttestationJson) (*zondpb.Attestation, error) {
	if jsonAttestation == nil {
		return nil, errors.New("json attestation is nil")
	}

	participationBits, err := hexutil.Decode(jsonAttestation.ParticipationBits)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode aggregation bits `%s`", jsonAttestation.ParticipationBits)
	}

	attestationData, err := convertAttestationDataToProto(jsonAttestation.Data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get attestation data")
	}

	signatures := make([][]byte, len(jsonAttestation.Signatures))
	for i, sig := range jsonAttestation.Signatures {
		signatures[i], err = hexutil.Decode(sig)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode attestation signature `%s`", sig)
		}
	}

	signaturesIdxToParticipationIdx := make([]uint64, len(jsonAttestation.SignaturesIdxToParticipationIdx))
	for i, participationIdx := range jsonAttestation.SignaturesIdxToParticipationIdx {
		signaturesIdxToParticipationIdx[i], err = strconv.ParseUint(participationIdx, 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse attestation signature to participation index `%s`", participationIdx)
		}
	}

	return &zondpb.Attestation{
		ParticipationBits:               participationBits,
		Data:                            attestationData,
		Signatures:                      signatures,
		SignaturesIdxToParticipationIdx: signaturesIdxToParticipationIdx,
	}, nil
}

func convertAttestationsToProto(jsonAttestations []*apimiddleware.AttestationJson) ([]*zondpb.Attestation, error) {
	var attestations []*zondpb.Attestation
	for index, jsonAttestation := range jsonAttestations {
		if jsonAttestation == nil {
			return nil, errors.Errorf("attestation at index `%d` is nil", index)
		}

		attestation, err := convertAttestationToProto(jsonAttestation)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to convert json attestation to proto at index %d", index)
		}

		attestations = append(attestations, attestation)
	}

	return attestations, nil
}

func convertAttestationDataToProto(jsonAttestationData *apimiddleware.AttestationDataJson) (*zondpb.AttestationData, error) {
	if jsonAttestationData == nil {
		return nil, errors.New("attestation data is nil")
	}

	slot, err := strconv.ParseUint(jsonAttestationData.Slot, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse attestation slot `%s`", jsonAttestationData.Slot)
	}

	committeeIndex, err := strconv.ParseUint(jsonAttestationData.CommitteeIndex, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse attestation committee index `%s`", jsonAttestationData.CommitteeIndex)
	}

	beaconBlockRoot, err := hexutil.Decode(jsonAttestationData.BeaconBlockRoot)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode attestation beacon block root `%s`", jsonAttestationData.BeaconBlockRoot)
	}

	sourceCheckpoint, err := convertCheckpointToProto(jsonAttestationData.Source)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get attestation source checkpoint")
	}

	targetCheckpoint, err := convertCheckpointToProto(jsonAttestationData.Target)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get attestation target checkpoint")
	}

	return &zondpb.AttestationData{
		Slot:            primitives.Slot(slot),
		CommitteeIndex:  primitives.CommitteeIndex(committeeIndex),
		BeaconBlockRoot: beaconBlockRoot,
		Source:          sourceCheckpoint,
		Target:          targetCheckpoint,
	}, nil
}

func convertDepositsToProto(jsonDeposits []*apimiddleware.DepositJson) ([]*zondpb.Deposit, error) {
	deposits := make([]*zondpb.Deposit, len(jsonDeposits))

	for depositIndex, jsonDeposit := range jsonDeposits {
		if jsonDeposit == nil {
			return nil, errors.Errorf("deposit at index `%d` is nil", depositIndex)
		}

		proofs := make([][]byte, len(jsonDeposit.Proof))
		for proofIndex, jsonProof := range jsonDeposit.Proof {
			proof, err := hexutil.Decode(jsonProof)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to decode deposit proof `%s`", jsonProof)
			}

			proofs[proofIndex] = proof
		}

		if jsonDeposit.Data == nil {
			return nil, errors.Errorf("deposit data at index `%d` is nil", depositIndex)
		}

		pubkey, err := hexutil.Decode(jsonDeposit.Data.PublicKey)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode deposit public key `%s`", jsonDeposit.Data.PublicKey)
		}

		withdrawalCredentials, err := hexutil.Decode(jsonDeposit.Data.WithdrawalCredentials)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode deposit withdrawal credentials `%s`", jsonDeposit.Data.WithdrawalCredentials)
		}

		amount, err := strconv.ParseUint(jsonDeposit.Data.Amount, 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse deposit amount `%s`", jsonDeposit.Data.Amount)
		}

		signature, err := hexutil.Decode(jsonDeposit.Data.Signature)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode signature `%s`", jsonDeposit.Data.Signature)
		}

		deposits[depositIndex] = &zondpb.Deposit{
			Proof: proofs,
			Data: &zondpb.Deposit_Data{
				PublicKey:             pubkey,
				WithdrawalCredentials: withdrawalCredentials,
				Amount:                amount,
				Signature:             signature,
			},
		}
	}

	return deposits, nil
}

func convertVoluntaryExitsToProto(jsonVoluntaryExits []*apimiddleware.SignedVoluntaryExitJson) ([]*zondpb.SignedVoluntaryExit, error) {
	attestingIndices := make([]*zondpb.SignedVoluntaryExit, len(jsonVoluntaryExits))

	for index, jsonVoluntaryExit := range jsonVoluntaryExits {
		if jsonVoluntaryExit == nil {
			return nil, errors.Errorf("signed voluntary exit at index `%d` is nil", index)
		}

		if jsonVoluntaryExit.Exit == nil {
			return nil, errors.Errorf("voluntary exit at index `%d` is nil", index)
		}

		epoch, err := strconv.ParseUint(jsonVoluntaryExit.Exit.Epoch, 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse voluntary exit epoch `%s`", jsonVoluntaryExit.Exit.Epoch)
		}

		validatorIndex, err := strconv.ParseUint(jsonVoluntaryExit.Exit.ValidatorIndex, 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse voluntary exit validator index `%s`", jsonVoluntaryExit.Exit.ValidatorIndex)
		}

		signature, err := hexutil.Decode(jsonVoluntaryExit.Signature)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode signature `%s`", jsonVoluntaryExit.Signature)
		}

		attestingIndices[index] = &zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				Epoch:          primitives.Epoch(epoch),
				ValidatorIndex: primitives.ValidatorIndex(validatorIndex),
			},
			Signature: signature,
		}
	}

	return attestingIndices, nil
}

func convertTransactionsToProto(jsonTransactions []string) ([][]byte, error) {
	transactions := make([][]byte, len(jsonTransactions))

	for index, jsonTransaction := range jsonTransactions {
		transaction, err := hexutil.Decode(jsonTransaction)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode transaction `%s`", jsonTransaction)
		}

		transactions[index] = transaction
	}

	return transactions, nil
}

func convertWithdrawalsToProto(jsonWithdrawals []*apimiddleware.WithdrawalJson) ([]*enginev1.Withdrawal, error) {
	withdrawals := make([]*enginev1.Withdrawal, len(jsonWithdrawals))

	for index, jsonWithdrawal := range jsonWithdrawals {
		if jsonWithdrawal == nil {
			return nil, errors.Errorf("withdrawal at index `%d` is nil", index)
		}

		withdrawalIndex, err := strconv.ParseUint(jsonWithdrawal.WithdrawalIndex, 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse withdrawal index `%s`", jsonWithdrawal.WithdrawalIndex)
		}

		validatorIndex, err := strconv.ParseUint(jsonWithdrawal.ValidatorIndex, 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse validator index `%s`", jsonWithdrawal.ValidatorIndex)
		}

		executionAddress, err := hexutil.Decode(jsonWithdrawal.ExecutionAddress)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode execution address `%s`", jsonWithdrawal.ExecutionAddress)
		}

		amount, err := strconv.ParseUint(jsonWithdrawal.Amount, 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse withdrawal amount `%s`", jsonWithdrawal.Amount)
		}

		withdrawals[index] = &enginev1.Withdrawal{
			Index:          withdrawalIndex,
			ValidatorIndex: primitives.ValidatorIndex(validatorIndex),
			Address:        executionAddress,
			Amount:         amount,
		}
	}

	return withdrawals, nil
}

func convertDilithiumToExecutionChangesToProto(jsonSignedDilithiumToExecutionChanges []*apimiddleware.SignedDilithiumToExecutionChangeJson) ([]*zondpb.SignedDilithiumToExecutionChange, error) {
	signedBlsToExecutionChanges := make([]*zondpb.SignedDilithiumToExecutionChange, len(jsonSignedDilithiumToExecutionChanges))

	for index, jsonDilithiumToExecutionChange := range jsonSignedDilithiumToExecutionChanges {
		if jsonDilithiumToExecutionChange == nil {
			return nil, errors.Errorf("dilithium to execution change at index `%d` is nil", index)
		}

		if jsonDilithiumToExecutionChange.Message == nil {
			return nil, errors.Errorf("dilithium to execution change message at index `%d` is nil", index)
		}

		validatorIndex, err := strconv.ParseUint(jsonDilithiumToExecutionChange.Message.ValidatorIndex, 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode validator index `%s`", jsonDilithiumToExecutionChange.Message.ValidatorIndex)
		}

		fromDilithiumPubkey, err := hexutil.Decode(jsonDilithiumToExecutionChange.Message.FromDilithiumPubkey)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode bls pubkey `%s`", jsonDilithiumToExecutionChange.Message.FromDilithiumPubkey)
		}

		toExecutionAddress, err := hexutil.Decode(jsonDilithiumToExecutionChange.Message.ToExecutionAddress)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode execution address `%s`", jsonDilithiumToExecutionChange.Message.ToExecutionAddress)
		}

		signature, err := hexutil.Decode(jsonDilithiumToExecutionChange.Signature)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode signature `%s`", jsonDilithiumToExecutionChange.Signature)
		}

		signedBlsToExecutionChanges[index] = &zondpb.SignedDilithiumToExecutionChange{
			Message: &zondpb.DilithiumToExecutionChange{
				ValidatorIndex:      primitives.ValidatorIndex(validatorIndex),
				FromDilithiumPubkey: fromDilithiumPubkey,
				ToExecutionAddress:  toExecutionAddress,
			},
			Signature: signature,
		}
	}

	return signedBlsToExecutionChanges, nil
}
