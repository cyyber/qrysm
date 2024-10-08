package accounts

import (
	"context"
	"time"

	"github.com/pkg/errors"
	grpcutil "github.com/theQRL/qrysm/api/grpc"
	"github.com/theQRL/qrysm/crypto/dilithium"
	"github.com/theQRL/qrysm/validator/accounts/wallet"
	iface "github.com/theQRL/qrysm/validator/client/iface"
	nodeClientFactory "github.com/theQRL/qrysm/validator/client/node-client-factory"
	validatorClientFactory "github.com/theQRL/qrysm/validator/client/validator-client-factory"
	validatorHelpers "github.com/theQRL/qrysm/validator/helpers"
	"github.com/theQRL/qrysm/validator/keymanager"
	"google.golang.org/grpc"
)

// NewCLIManager allows for managing validator accounts via CLI commands.
func NewCLIManager(opts ...Option) (*AccountsCLIManager, error) {
	acc := &AccountsCLIManager{
		// mnemonicLanguage: derived.DefaultMnemonicLanguage,
	}
	for _, opt := range opts {
		if err := opt(acc); err != nil {
			return nil, err
		}
	}
	return acc, nil
}

// AccountsCLIManager defines a struct capable of performing various validator
// wallet & account operations via the command line.
type AccountsCLIManager struct {
	wallet               *wallet.Wallet
	keymanager           keymanager.IKeymanager
	keymanagerKind       keymanager.Kind
	showDepositData      bool
	showPrivateKeys      bool
	listValidatorIndices bool
	deletePublicKeys     bool
	importPrivateKeys    bool
	readPasswordFile     bool
	// skipMnemonicConfirm  bool
	dialOpts          []grpc.DialOption
	grpcHeaders       []string
	beaconRPCProvider string
	walletKeyCount    int
	privateKeyFile    string
	passwordFilePath  string
	keysDir           string
	// mnemonicLanguage     string
	backupsDir         string
	backupsPassword    string
	filteredPubKeys    []dilithium.PublicKey
	rawPubKeys         [][]byte
	formattedPubKeys   []string
	exitJSONOutputPath string
	walletDir          string
	walletPassword     string
	// mnemonic             string
	// numAccounts          int
	// mnemonic25thWord     string
	beaconApiEndpoint string
	beaconApiTimeout  time.Duration
}

func (acm *AccountsCLIManager) prepareBeaconClients(ctx context.Context) (*iface.ValidatorClient, *iface.NodeClient, error) {
	if acm.dialOpts == nil {
		return nil, nil, errors.New("failed to construct dial options for beacon clients")
	}

	ctx = grpcutil.AppendHeaders(ctx, acm.grpcHeaders)
	grpcConn, err := grpc.DialContext(ctx, acm.beaconRPCProvider, acm.dialOpts...)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not dial endpoint %s", acm.beaconRPCProvider)
	}

	conn := validatorHelpers.NewNodeConnection(
		grpcConn,
		acm.beaconApiEndpoint,
		acm.beaconApiTimeout,
	)

	validatorClient := validatorClientFactory.NewValidatorClient(conn)
	nodeClient := nodeClientFactory.NewNodeClient(conn)
	return &validatorClient, &nodeClient, nil
}
