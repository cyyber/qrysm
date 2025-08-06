// Package validator defines a gRPC validator service implementation, providing
// critical endpoints for validator clients to submit blocks/attestations to the
// beacon node, receive assignments, and more.
package validator

import (
	"context"
	"time"

	"github.com/theQRL/qrysm/beacon-chain/blockchain"
	"github.com/theQRL/qrysm/beacon-chain/builder"
	"github.com/theQRL/qrysm/beacon-chain/cache"
	"github.com/theQRL/qrysm/beacon-chain/cache/depositcache"
	blockfeed "github.com/theQRL/qrysm/beacon-chain/core/feed/block"
	opfeed "github.com/theQRL/qrysm/beacon-chain/core/feed/operation"
	statefeed "github.com/theQRL/qrysm/beacon-chain/core/feed/state"
	"github.com/theQRL/qrysm/beacon-chain/core/signing"
	"github.com/theQRL/qrysm/beacon-chain/db"
	"github.com/theQRL/qrysm/beacon-chain/execution"
	"github.com/theQRL/qrysm/beacon-chain/operations/attestations"
	"github.com/theQRL/qrysm/beacon-chain/operations/dilithiumtoexec"
	"github.com/theQRL/qrysm/beacon-chain/operations/slashings"
	"github.com/theQRL/qrysm/beacon-chain/operations/synccommittee"
	"github.com/theQRL/qrysm/beacon-chain/operations/voluntaryexits"
	"github.com/theQRL/qrysm/beacon-chain/p2p"
	"github.com/theQRL/qrysm/beacon-chain/rpc/core"
	"github.com/theQRL/qrysm/beacon-chain/startup"
	"github.com/theQRL/qrysm/beacon-chain/state/stategen"
	"github.com/theQRL/qrysm/beacon-chain/sync"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	"github.com/theQRL/qrysm/network/forks"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server defines a server implementation of the gRPC Validator service,
// providing RPC endpoints for obtaining validator assignments per epoch, the slots
// and committees in which particular validators need to perform their responsibilities,
// and more.
type Server struct {
	Ctx                    context.Context
	ProposerSlotIndexCache *cache.ProposerPayloadIDsCache
	HeadFetcher            blockchain.HeadFetcher
	ForkFetcher            blockchain.ForkFetcher
	ForkchoiceFetcher      blockchain.ForkchoiceFetcher
	GenesisFetcher         blockchain.GenesisFetcher
	FinalizationFetcher    blockchain.FinalizationFetcher
	TimeFetcher            blockchain.TimeFetcher
	BlockFetcher           execution.ExecutionBlockFetcher
	DepositFetcher         cache.DepositFetcher
	ChainStartFetcher      execution.ChainStartFetcher
	ExecutionInfoFetcher   execution.ChainInfoFetcher
	OptimisticModeFetcher  blockchain.OptimisticModeFetcher
	SyncChecker            sync.Checker
	StateNotifier          statefeed.Notifier
	BlockNotifier          blockfeed.Notifier
	P2P                    p2p.Broadcaster
	AttPool                attestations.Pool
	SlashingsPool          slashings.PoolManager
	ExitPool               voluntaryexits.PoolManager
	SyncCommitteePool      synccommittee.Pool
	BlockReceiver          blockchain.BlockReceiver
	MockExecutionVotes     bool
	ExecutionBlockFetcher  execution.ExecutionBlockFetcher
	PendingDepositsFetcher depositcache.PendingDepositsFetcher
	OperationNotifier      opfeed.Notifier
	StateGen               stategen.StateManager
	ReplayerBuilder        stategen.ReplayerBuilder
	BeaconDB               db.HeadAccessDatabase
	ExecutionEngineCaller  execution.EngineCaller
	BlockBuilder           builder.BlockBuilder
	DilithiumChangesPool   dilithiumtoexec.PoolManager
	ClockWaiter            startup.ClockWaiter
	CoreService            *core.Service
}

// WaitForActivation checks if a validator public key exists in the active validator registry of the current
// beacon state, if not, then it creates a stream which listens for canonical states which contain
// the validator with the public key as an active validator record.
func (vs *Server) WaitForActivation(req *qrysmpb.ValidatorActivationRequest, stream qrysmpb.BeaconNodeValidator_WaitForActivationServer) error {
	activeValidatorExists, validatorStatuses, err := vs.activationStatus(stream.Context(), req.PublicKeys)
	if err != nil {
		return status.Errorf(codes.Internal, "Could not fetch validator status: %v", err)
	}
	res := &qrysmpb.ValidatorActivationResponse{
		Statuses: validatorStatuses,
	}
	if activeValidatorExists {
		return stream.Send(res)
	}
	if err := stream.Send(res); err != nil {
		return status.Errorf(codes.Internal, "Could not send response over stream: %v", err)
	}

	for {
		select {
		// Pinging every slot for activation.
		case <-time.After(time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second):
			activeValidatorExists, validatorStatuses, err := vs.activationStatus(stream.Context(), req.PublicKeys)
			if err != nil {
				return status.Errorf(codes.Internal, "Could not fetch validator status: %v", err)
			}
			res := &qrysmpb.ValidatorActivationResponse{
				Statuses: validatorStatuses,
			}
			if activeValidatorExists {
				return stream.Send(res)
			}
			if err := stream.Send(res); err != nil {
				return status.Errorf(codes.Internal, "Could not send response over stream: %v", err)
			}
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "Stream context canceled")
		case <-vs.Ctx.Done():
			return status.Error(codes.Canceled, "RPC context canceled")
		}
	}
}

// ValidatorIndex is called by a validator to get its index location in the beacon state.
func (vs *Server) ValidatorIndex(ctx context.Context, req *qrysmpb.ValidatorIndexRequest) (*qrysmpb.ValidatorIndexResponse, error) {
	st, err := vs.HeadFetcher.HeadStateReadOnly(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not determine head state: %v", err)
	}
	if st == nil || st.IsNil() {
		return nil, status.Errorf(codes.Internal, "head state is empty")
	}
	index, ok := st.ValidatorIndexByPubkey(bytesutil.ToBytes2592(req.PublicKey))
	if !ok {
		return nil, status.Errorf(codes.NotFound, "Could not find validator index for public key %#x", req.PublicKey)
	}

	return &qrysmpb.ValidatorIndexResponse{Index: index}, nil
}

// DomainData fetches the current domain version information from the beacon state.
func (vs *Server) DomainData(ctx context.Context, request *qrysmpb.DomainRequest) (*qrysmpb.DomainResponse, error) {
	fork, err := forks.Fork(request.Epoch)
	if err != nil {
		return nil, err
	}
	headGenesisValidatorsRoot := vs.HeadFetcher.HeadGenesisValidatorsRoot()
	dv, err := signing.Domain(fork, request.Epoch, bytesutil.ToBytes4(request.Domain), headGenesisValidatorsRoot[:])
	if err != nil {
		return nil, err
	}
	return &qrysmpb.DomainResponse{
		SignatureDomain: dv,
	}, nil
}

// WaitForChainStart queries the logs of the Deposit Contract in order to verify the beacon chain
// has started its runtime and validators begin their responsibilities. If it has not, it then
// subscribes to an event stream triggered by the execution chain service whenever the ChainStart log does
// occur in the Deposit Contract on QRL execution layer.
func (vs *Server) WaitForChainStart(_ *emptypb.Empty, stream qrysmpb.BeaconNodeValidator_WaitForChainStartServer) error {
	head, err := vs.HeadFetcher.HeadStateReadOnly(stream.Context())
	if err != nil {
		return status.Errorf(codes.Internal, "Could not retrieve head state: %v", err)
	}
	if head != nil && !head.IsNil() {
		res := &qrysmpb.ChainStartResponse{
			Started:               true,
			GenesisTime:           head.GenesisTime(),
			GenesisValidatorsRoot: head.GenesisValidatorsRoot(),
		}
		return stream.Send(res)
	}

	clock, err := vs.ClockWaiter.WaitForClock(vs.Ctx)
	if err != nil {
		return status.Error(codes.Canceled, "Context canceled")
	}
	log.WithField("starttime", clock.GenesisTime()).Debug("Received chain started event")
	log.Debug("Sending genesis time notification to connected validator clients")
	gvr := clock.GenesisValidatorsRoot()
	res := &qrysmpb.ChainStartResponse{
		Started:               true,
		GenesisTime:           uint64(clock.GenesisTime().Unix()),
		GenesisValidatorsRoot: gvr[:],
	}
	return stream.Send(res)
}
