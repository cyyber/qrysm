package blocks

import (
	"github.com/pkg/errors"
	field_params "github.com/theQRL/qrysm/config/fieldparams"
	"github.com/theQRL/qrysm/consensus-types/interfaces"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
)

var (
	_ = interfaces.ReadOnlySignedBeaconBlock(&SignedBeaconBlock{})
	_ = interfaces.ReadOnlyBeaconBlock(&BeaconBlock{})
	_ = interfaces.ReadOnlyBeaconBlockBody(&BeaconBlockBody{})
)

var (
	errPayloadWrongType       = errors.New("execution payload has wrong type")
	errPayloadHeaderWrongType = errors.New("execution payload header has wrong type")
)

const (
	incorrectBlockVersion = "incorrect beacon block version"
	incorrectBodyVersion  = "incorrect beacon block body version"
)

var (
	// ErrUnsupportedVersion for beacon block methods.
	ErrUnsupportedVersion    = errors.New("unsupported beacon block version")
	errNilBlock              = errors.New("received nil beacon block")
	errNilBlockBody          = errors.New("received nil beacon block body")
	errIncorrectBlockVersion = errors.New(incorrectBlockVersion)
	errIncorrectBodyVersion  = errors.New(incorrectBodyVersion)
)

// BeaconBlockBody is the main beacon block body structure. It can represent any block type.
type BeaconBlockBody struct {
	version                   int
	isBlinded                 bool
	randaoReveal              [field_params.MLDSA87SignatureLength]byte
	executionData             *qrysmpb.ExecutionData
	graffiti                  [field_params.RootLength]byte
	proposerSlashings         []*qrysmpb.ProposerSlashing
	attesterSlashings         []*qrysmpb.AttesterSlashing
	attestations              []*qrysmpb.Attestation
	deposits                  []*qrysmpb.Deposit
	voluntaryExits            []*qrysmpb.SignedVoluntaryExit
	syncAggregate             *qrysmpb.SyncAggregate
	executionPayload          interfaces.ExecutionData
	executionPayloadHeader    interfaces.ExecutionData
	mlDSA87ToExecutionChanges []*qrysmpb.SignedMLDSA87ToExecutionChange
}

// BeaconBlock is the main beacon block structure. It can represent any block type.
type BeaconBlock struct {
	version       int
	slot          primitives.Slot
	proposerIndex primitives.ValidatorIndex
	parentRoot    [field_params.RootLength]byte
	stateRoot     [field_params.RootLength]byte
	body          *BeaconBlockBody
}

// SignedBeaconBlock is the main signed beacon block structure. It can represent any block type.
type SignedBeaconBlock struct {
	version   int
	block     *BeaconBlock
	signature [field_params.MLDSA87SignatureLength]byte
}
