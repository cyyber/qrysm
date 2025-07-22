package execution

import (
	"context"

	"github.com/pkg/errors"
	"github.com/theQRL/go-zond/common"
	"github.com/theQRL/qrysm/beacon-chain/core/blocks"
	"github.com/theQRL/qrysm/config/params"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
)

// DepositContractAddress returns the deposit contract address for the given chain.
func DepositContractAddress() (string, error) {
	address := params.BeaconConfig().DepositContractAddress
	if address == "" {
		return "", errors.New("valid deposit contract is required")
	}

	if !common.IsAddress(address) {
		return "", errors.New("invalid deposit contract address given: " + address)
	}
	return address, nil
}

func (s *Service) processDeposit(ctx context.Context, executionNodeData *qrysmpb.ExecutionNodeData, deposit *qrysmpb.Deposit) error {
	var err error
	if err := s.preGenesisState.SetExecutionNodeData(executionNodeData); err != nil {
		return err
	}
	beaconState, err := blocks.ProcessPreGenesisDeposits(ctx, s.preGenesisState, []*qrysmpb.Deposit{deposit})
	if err != nil {
		return errors.Wrap(err, "could not process pre-genesis deposits")
	}
	if beaconState != nil && !beaconState.IsNil() {
		s.preGenesisState = beaconState
	}
	return nil
}
