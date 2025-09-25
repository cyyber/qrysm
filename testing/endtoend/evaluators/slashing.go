package evaluators

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/theQRL/go-bitfield"
	"github.com/theQRL/qrysm/beacon-chain/core/signing"
	field_params "github.com/theQRL/qrysm/config/fieldparams"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/blocks"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/container/slice"
	"github.com/theQRL/qrysm/crypto/ml_dsa_87"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	e2e "github.com/theQRL/qrysm/testing/endtoend/params"
	"github.com/theQRL/qrysm/testing/endtoend/policies"
	e2eTypes "github.com/theQRL/qrysm/testing/endtoend/types"
	"github.com/theQRL/qrysm/testing/util"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// InjectDoubleVoteOnEpoch broadcasts a double vote into the beacon node pool for the slasher to detect.
var InjectDoubleVoteOnEpoch = func(n primitives.Epoch) e2eTypes.Evaluator {
	return e2eTypes.Evaluator{
		Name:       "inject_double_vote_%d",
		Policy:     policies.OnEpoch(n),
		Evaluation: insertDoubleAttestationIntoPool,
	}
}

// InjectDoubleBlockOnEpoch proposes a double block to the beacon node for the slasher to detect.
var InjectDoubleBlockOnEpoch = func(n primitives.Epoch) e2eTypes.Evaluator {
	return e2eTypes.Evaluator{
		Name:       "inject_double_block_%d",
		Policy:     policies.OnEpoch(n),
		Evaluation: proposeDoubleBlock,
	}
}

// ValidatorsSlashedAfterEpoch ensures the expected amount of validators are slashed.
var ValidatorsSlashedAfterEpoch = func(n primitives.Epoch) e2eTypes.Evaluator {
	return e2eTypes.Evaluator{
		Name:       "validators_slashed_epoch_%d",
		Policy:     policies.AfterNthEpoch(n),
		Evaluation: validatorsSlashed,
	}
}

// SlashedValidatorsLoseBalanceAfterEpoch checks if the validators slashed lose the right balance.
var SlashedValidatorsLoseBalanceAfterEpoch = func(n primitives.Epoch) e2eTypes.Evaluator {
	return e2eTypes.Evaluator{
		Name:       "slashed_validators_lose_valance_epoch_%d",
		Policy:     policies.AfterNthEpoch(n),
		Evaluation: validatorsLoseBalance,
	}
}

var slashedIndices []uint64

func validatorsSlashed(_ *e2eTypes.EvaluationContext, conns ...*grpc.ClientConn) error {
	conn := conns[0]
	ctx := context.Background()
	client := qrysmpb.NewBeaconChainClient(conn)

	actualSlashedIndices := 0

	for _, slashedIndex := range slashedIndices {
		req := &qrysmpb.GetValidatorRequest{
			QueryFilter: &qrysmpb.GetValidatorRequest_Index{
				Index: primitives.ValidatorIndex(slashedIndex),
			},
		}
		valResp, err := client.GetValidator(ctx, req)
		if err != nil {
			return err
		}

		if valResp.Slashed {
			actualSlashedIndices++
		}
	}

	if actualSlashedIndices != len(slashedIndices) {
		return fmt.Errorf("expected %d indices to be slashed, received %d", len(slashedIndices), actualSlashedIndices)
	}
	return nil
}

func validatorsLoseBalance(_ *e2eTypes.EvaluationContext, conns ...*grpc.ClientConn) error {
	conn := conns[0]
	ctx := context.Background()
	client := qrysmpb.NewBeaconChainClient(conn)

	for i, slashedIndex := range slashedIndices {
		req := &qrysmpb.GetValidatorRequest{
			QueryFilter: &qrysmpb.GetValidatorRequest_Index{
				Index: primitives.ValidatorIndex(slashedIndex),
			},
		}
		valResp, err := client.GetValidator(ctx, req)
		if err != nil {
			return err
		}

		slashedPenalty := params.BeaconConfig().MaxEffectiveBalance / params.BeaconConfig().MinSlashingPenaltyQuotient
		slashedBal := params.BeaconConfig().MaxEffectiveBalance - slashedPenalty + params.BeaconConfig().EffectiveBalanceIncrement/10
		if valResp.EffectiveBalance >= slashedBal {
			return fmt.Errorf(
				"expected slashed validator %d to balance less than %d, received %d",
				i,
				slashedBal,
				valResp.EffectiveBalance,
			)
		}
	}
	return nil
}

func insertDoubleAttestationIntoPool(_ *e2eTypes.EvaluationContext, conns ...*grpc.ClientConn) error {
	conn := conns[0]
	valClient := qrysmpb.NewBeaconNodeValidatorClient(conn)
	beaconClient := qrysmpb.NewBeaconChainClient(conn)

	ctx := context.Background()
	chainHead, err := beaconClient.GetChainHead(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.Wrap(err, "could not get chain head")
	}

	_, privKeys, err := util.DeterministicDepositsAndKeys(params.BeaconConfig().MinGenesisActiveValidatorCount)
	if err != nil {
		return err
	}
	pubKeys := make([][]byte, len(privKeys))
	for i, priv := range privKeys {
		pubKeys[i] = priv.PublicKey().Marshal()
	}
	duties, err := valClient.GetDuties(ctx, &qrysmpb.DutiesRequest{
		Epoch:      chainHead.HeadEpoch,
		PublicKeys: pubKeys,
	})
	if err != nil {
		return errors.Wrap(err, "could not get duties")
	}

	var committeeIndex primitives.CommitteeIndex
	var committee []primitives.ValidatorIndex
	for _, duty := range duties.CurrentEpochDuties {
		if duty.AttesterSlot == chainHead.HeadSlot-1 {
			committeeIndex = duty.CommitteeIndex
			committee = duty.Committee
			break
		}
	}

	attDataReq := &qrysmpb.AttestationDataRequest{
		CommitteeIndex: committeeIndex,
		Slot:           chainHead.HeadSlot - 1,
	}

	attData, err := valClient.GetAttestationData(ctx, attDataReq)
	if err != nil {
		return err
	}
	blockRoot := bytesutil.ToBytes32([]byte("muahahahaha I'm an evil validator"))
	attData.BeaconBlockRoot = blockRoot[:]

	req := &qrysmpb.DomainRequest{
		Epoch:  chainHead.HeadEpoch,
		Domain: params.BeaconConfig().DomainBeaconAttester[:],
	}
	resp, err := valClient.DomainData(ctx, req)
	if err != nil {
		return errors.Wrap(err, "could not get domain data")
	}
	signingRoot, err := signing.ComputeSigningRoot(attData, resp.SignatureDomain)
	if err != nil {
		return errors.Wrap(err, "could not compute signing root")
	}

	valsToSlash := uint64(2)
	for i := uint64(0); i < valsToSlash && i < uint64(len(committee)); i++ {
		if len(slice.IntersectionUint64(slashedIndices, []uint64{uint64(committee[i])})) > 0 {
			valsToSlash++
			continue
		}
		// Set the bits of half the committee to be slashed.
		attBitfield := bitfield.NewBitlist(uint64(len(committee)))
		attBitfield.SetBitAt(i, true)

		att := &qrysmpb.Attestation{
			AggregationBits: attBitfield,
			Data:            attData,
			Signatures:      [][]byte{privKeys[committee[i]].Sign(signingRoot[:]).Marshal()},
		}
		// We only broadcast to conns[0] here since we can trust that at least 1 node will be online.
		// Only broadcasting the attestation to one node also helps test slashing propagation.
		client := qrysmpb.NewBeaconNodeValidatorClient(conns[0])
		if _, err = client.ProposeAttestation(ctx, att); err != nil {
			return errors.Wrap(err, "could not propose attestation")
		}
		slashedIndices = append(slashedIndices, uint64(committee[i]))
	}
	return nil
}

func proposeDoubleBlock(_ *e2eTypes.EvaluationContext, conns ...*grpc.ClientConn) error {
	conn := conns[0]
	valClient := qrysmpb.NewBeaconNodeValidatorClient(conn)
	beaconClient := qrysmpb.NewBeaconChainClient(conn)

	ctx := context.Background()
	chainHead, err := beaconClient.GetChainHead(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.Wrap(err, "could not get chain head")
	}
	_, privKeys, err := util.DeterministicDepositsAndKeys(params.BeaconConfig().MinGenesisActiveValidatorCount)
	if err != nil {
		return err
	}
	pubKeys := make([][]byte, len(privKeys))
	for i, priv := range privKeys {
		pubKeys[i] = priv.PublicKey().Marshal()
	}
	duties, err := valClient.GetDuties(ctx, &qrysmpb.DutiesRequest{
		Epoch:      chainHead.HeadEpoch,
		PublicKeys: pubKeys,
	})
	if err != nil {
		return errors.Wrap(err, "could not get duties")
	}

	var proposerIndex primitives.ValidatorIndex
	for i, duty := range duties.CurrentEpochDuties {
		if slice.IsInSlots(chainHead.HeadSlot-1, duty.ProposerSlots) {
			proposerIndex = primitives.ValidatorIndex(i)
			break
		}
	}

	validatorNum := int(params.BeaconConfig().MinGenesisActiveValidatorCount)
	beaconNodeNum := e2e.TestParams.BeaconNodeCount
	if validatorNum%beaconNodeNum != 0 {
		return errors.New("validator count is not easily divisible by beacon node count")
	}
	validatorsPerNode := validatorNum / beaconNodeNum

	// If the proposer index is in the second validator client, we connect to
	// the corresponding beacon node instead.
	if proposerIndex >= primitives.ValidatorIndex(uint64(validatorsPerNode)) {
		valClient = qrysmpb.NewBeaconNodeValidatorClient(conns[1])
	}

	b, err := generateSignedBeaconBlock(chainHead, proposerIndex, valClient, privKeys, "bad state root")
	if err != nil {
		return err
	}
	if _, err = valClient.ProposeBeaconBlock(ctx, b); err == nil {
		return errors.New("expected block to fail processing")
	}

	b, err = generateSignedBeaconBlock(chainHead, proposerIndex, valClient, privKeys, "bad state root 2")
	if err != nil {
		return err
	}
	if _, err = valClient.ProposeBeaconBlock(ctx, b); err == nil {
		return errors.New("expected block to fail processing")
	}

	slashedIndices = append(slashedIndices, uint64(proposerIndex))
	return nil
}

func generateSignedBeaconBlock(
	chainHead *qrysmpb.ChainHead,
	proposerIndex primitives.ValidatorIndex,
	valClient qrysmpb.BeaconNodeValidatorClient,
	privKeys []ml_dsa_87.MLDSA87Key,
	stateRoot string,
) (*qrysmpb.GenericSignedBeaconBlock, error) {
	ctx := context.Background()

	hashLen := 32
	blk := &qrysmpb.BeaconBlockCapella{
		Slot:          chainHead.HeadSlot - 1,
		ParentRoot:    chainHead.HeadBlockRoot,
		StateRoot:     bytesutil.PadTo([]byte(stateRoot), hashLen),
		ProposerIndex: proposerIndex,
		Body: &qrysmpb.BeaconBlockBodyCapella{
			ExecutionData: &qrysmpb.ExecutionData{
				BlockHash:    bytesutil.PadTo([]byte("bad block hash"), hashLen),
				DepositRoot:  bytesutil.PadTo([]byte("bad deposit root"), hashLen),
				DepositCount: 1,
			},
			RandaoReveal:      bytesutil.PadTo([]byte("bad randao"), field_params.MLDSA87SignatureLength),
			Graffiti:          bytesutil.PadTo([]byte("teehee"), hashLen),
			ProposerSlashings: []*qrysmpb.ProposerSlashing{},
			AttesterSlashings: []*qrysmpb.AttesterSlashing{},
			Attestations:      []*qrysmpb.Attestation{},
			Deposits:          []*qrysmpb.Deposit{},
			VoluntaryExits:    []*qrysmpb.SignedVoluntaryExit{},
		},
	}

	req := &qrysmpb.DomainRequest{
		Epoch:  chainHead.HeadEpoch,
		Domain: params.BeaconConfig().DomainBeaconProposer[:],
	}
	resp, err := valClient.DomainData(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "could not get domain data")
	}
	signingRoot, err := signing.ComputeSigningRoot(blk, resp.SignatureDomain)
	if err != nil {
		return nil, errors.Wrap(err, "could not compute signing root")
	}
	sig := privKeys[proposerIndex].Sign(signingRoot[:]).Marshal()
	signedBlk := &qrysmpb.SignedBeaconBlockCapella{
		Block:     blk,
		Signature: sig,
	}

	// We only broadcast to conns[0] here since we can trust that at least 1 node will be online.
	// Only broadcasting the attestation to one node also helps test slashing propagation.
	wb, err := blocks.NewSignedBeaconBlock(signedBlk)
	if err != nil {
		return nil, err
	}
	return wb.PbGenericBlock()
}
