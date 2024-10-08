package evaluators

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	mathutil "github.com/theQRL/qrysm/math"
	"github.com/theQRL/qrysm/proto/zond/service"
	v1 "github.com/theQRL/qrysm/proto/zond/v1"
	"github.com/theQRL/qrysm/testing/endtoend/policies"
	"github.com/theQRL/qrysm/testing/endtoend/types"
	"github.com/theQRL/qrysm/time/slots"
	"google.golang.org/grpc"
)

// OptimisticSyncEnabled checks that the node is in an optimistic state.
var OptimisticSyncEnabled = types.Evaluator{
	Name:       "optimistic_sync_at_epoch_%d",
	Policy:     policies.AllEpochs,
	Evaluation: optimisticSyncEnabled,
}

func optimisticSyncEnabled(_ *types.EvaluationContext, conns ...*grpc.ClientConn) error {
	for _, conn := range conns {
		client := service.NewBeaconChainClient(conn)
		head, err := client.GetBlindedBlock(context.Background(), &v1.BlockRequest{BlockId: []byte("head")})
		if err != nil {
			return err
		}
		headSlot := uint64(0)
		switch hb := head.Data.Message.(type) {
		case *v1.SignedBlindedBeaconBlockContainer_CapellaBlock:
			headSlot = uint64(hb.CapellaBlock.Slot)
		default:
			return errors.New("no valid block type retrieved")
		}
		currEpoch := slots.ToEpoch(primitives.Slot(headSlot))
		startSlot, err := slots.EpochStart(currEpoch)
		if err != nil {
			return err
		}
		isOptimistic := false
		for i := startSlot; i <= primitives.Slot(headSlot); i++ {
			castI, err := mathutil.Int(uint64(i))
			if err != nil {
				return err
			}
			block, err := client.GetBlindedBlock(context.Background(), &v1.BlockRequest{BlockId: []byte(strconv.Itoa(castI))})
			if err != nil {
				// Continue in the event of non-existent blocks.
				continue
			}
			if !block.ExecutionOptimistic {
				return errors.New("expected block to be optimistic, but it is not")
			}
			isOptimistic = true
		}
		if !isOptimistic {
			return errors.New("expected block to be optimistic, but it is not")
		}
	}
	return nil
}
