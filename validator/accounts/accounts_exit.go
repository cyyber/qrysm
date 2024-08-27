package accounts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/theQRL/go-zond/common/hexutil"
	"github.com/theQRL/qrysm/v4/beacon-chain/core/blocks"
	field_params "github.com/theQRL/qrysm/v4/config/fieldparams"
	"github.com/theQRL/qrysm/v4/encoding/bytesutil"
	"github.com/theQRL/qrysm/v4/io/file"
	zond "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/v4/validator/client"
	beacon_api "github.com/theQRL/qrysm/v4/validator/client/beacon-api"
	"github.com/theQRL/qrysm/v4/validator/client/iface"
	"github.com/theQRL/qrysm/v4/validator/keymanager"
	"google.golang.org/protobuf/types/known/emptypb"
)

// PerformExitCfg for account voluntary exits.
type PerformExitCfg struct {
	ValidatorClient  iface.ValidatorClient
	NodeClient       iface.NodeClient
	Keymanager       keymanager.IKeymanager
	RawPubKeys       [][]byte
	FormattedPubKeys []string
	OutputDirectory  string
}

// ExitPassphrase exported for use in test.
const ExitPassphrase = "Exit my validator"

// Exit performs a voluntary exit on one or more accounts.
func (acm *AccountsCLIManager) Exit(ctx context.Context) error {
	// User decided to cancel the voluntary exit.
	if acm.rawPubKeys == nil && acm.formattedPubKeys == nil {
		return nil
	}

	validatorClient, nodeClient, err := acm.prepareBeaconClients(ctx)
	if err != nil {
		return err
	}
	if nodeClient == nil {
		return errors.New("could not prepare beacon node client")
	}
	syncStatus, err := (*nodeClient).GetSyncStatus(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}
	if syncStatus == nil {
		return errors.New("could not get sync status")
	}

	if syncStatus.Syncing {
		return errors.New("could not perform exit: beacon node is syncing.")
	}

	cfg := PerformExitCfg{
		*validatorClient,
		*nodeClient,
		acm.keymanager,
		acm.rawPubKeys,
		acm.formattedPubKeys,
		acm.exitJSONOutputPath,
	}
	rawExitedKeys, trimmedExitedKeys, err := PerformVoluntaryExit(ctx, cfg)
	if err != nil {
		return err
	}
	displayExitInfo(rawExitedKeys, trimmedExitedKeys)

	return nil
}

// PerformVoluntaryExit uses gRPC clients to submit a voluntary exit message to a beacon node.
func PerformVoluntaryExit(
	ctx context.Context, cfg PerformExitCfg,
) (rawExitedKeys [][]byte, formattedExitedKeys []string, err error) {
	var rawNotExitedKeys [][]byte
	genesisResponse, err := cfg.NodeClient.GetGenesis(ctx, &emptypb.Empty{})
	if err != nil {
		log.WithError(err).Errorf("voluntary exit failed: %v", err)
	}
	for i, key := range cfg.RawPubKeys {
		// When output directory is present, only create the signed exit, but do not propose it.
		// Otherwise, propose the exit immediately.
		epoch, err := client.CurrentEpoch(genesisResponse.GenesisTime)
		if err != nil {
			log.WithError(err).Errorf("voluntary exit failed: %v", err)
		}
		if len(cfg.OutputDirectory) > 0 {
			sve, err := client.CreateSignedVoluntaryExit(ctx, cfg.ValidatorClient, cfg.Keymanager.Sign, key, epoch)
			if err != nil {
				rawNotExitedKeys = append(rawNotExitedKeys, key)
				msg := err.Error()
				if strings.Contains(msg, blocks.ValidatorAlreadyExitedMsg) ||
					strings.Contains(msg, blocks.ValidatorCannotExitYetMsg) {
					log.Warningf("Could not create voluntary exit for account %s: %s", cfg.FormattedPubKeys[i], msg)
				} else {
					log.WithError(err).Errorf("voluntary exit failed for account %s", cfg.FormattedPubKeys[i])
				}
			} else if err := writeSignedVoluntaryExitJSON(ctx, sve, cfg.OutputDirectory); err != nil {
				log.WithError(err).Error("failed to write voluntary exit")
			}
		} else if err := client.ProposeExit(ctx, cfg.ValidatorClient, cfg.Keymanager.Sign, key, epoch); err != nil {
			rawNotExitedKeys = append(rawNotExitedKeys, key)

			msg := err.Error()
			if strings.Contains(msg, blocks.ValidatorAlreadyExitedMsg) ||
				strings.Contains(msg, blocks.ValidatorCannotExitYetMsg) {
				log.Warningf("Could not perform voluntary exit for account %s: %s", cfg.FormattedPubKeys[i], msg)
			} else {
				log.WithError(err).Errorf("voluntary exit failed for account %s", cfg.FormattedPubKeys[i])
			}
		}
	}

	rawExitedKeys = make([][]byte, 0)
	formattedExitedKeys = make([]string, 0)
	for i, key := range cfg.RawPubKeys {
		found := false
		for _, notExited := range rawNotExitedKeys {
			if bytes.Equal(notExited, key) {
				found = true
				break
			}
		}
		if !found {
			rawExitedKeys = append(rawExitedKeys, key)
			formattedExitedKeys = append(formattedExitedKeys, cfg.FormattedPubKeys[i])
		}
	}

	return rawExitedKeys, formattedExitedKeys, nil
}

func prepareAllKeys(validatingKeys [][field_params.DilithiumPubkeyLength]byte) (raw [][]byte, formatted []string) {
	raw = make([][]byte, len(validatingKeys))
	formatted = make([]string, len(validatingKeys))
	for i, pk := range validatingKeys {
		raw[i] = make([]byte, len(pk))
		copy(raw[i], pk[:])
		formatted[i] = fmt.Sprintf("%#x", bytesutil.Trunc(pk[:]))
	}
	return
}

func displayExitInfo(rawExitedKeys [][]byte, trimmedExitedKeys []string) {
	if len(rawExitedKeys) > 0 {
		// TODO(now.youtrack.cloud/issue/TQ-1): include our explorer url when available
		// urlFormattedPubKeys := make([]string, len(rawExitedKeys))
		// for i, key := range rawExitedKeys {
		// 	var baseUrl string
		// 	if params.BeaconConfig().ConfigName == params.PraterName || params.BeaconConfig().ConfigName == params.GoerliName {
		// 		baseUrl = "https://goerli.beaconcha.in/validator/"
		// 	} else {
		// 		baseUrl = "https://beaconcha.in/validator/"
		// 	}
		// 	// Remove '0x' prefix
		// 	urlFormattedPubKeys[i] = baseUrl + hexutil.Encode(key)[2:]
		// }

		// ifaceKeys := make([]interface{}, len(urlFormattedPubKeys))
		// for i, k := range urlFormattedPubKeys {
		// 	ifaceKeys[i] = k
		// }

		ifaceKeys := make([]interface{}, len(rawExitedKeys))
		for i, k := range rawExitedKeys {
			ifaceKeys[i] = hexutil.Encode(k)
		}

		info := fmt.Sprintf("Voluntary exit was successful for the accounts listed:\n"+
			strings.Repeat("%s\n", len(ifaceKeys)), ifaceKeys...)

		log.WithField("publicKeys", strings.Join(trimmedExitedKeys, ", ")).Info(info)
	} else {
		log.Info("No successful voluntary exits")
	}
}

func writeSignedVoluntaryExitJSON(ctx context.Context, sve *zond.SignedVoluntaryExit, outputDirectory string) error {
	if err := file.MkdirAll(outputDirectory); err != nil {
		return err
	}

	jsve := beacon_api.JsonifySignedVoluntaryExits([]*zond.SignedVoluntaryExit{sve})[0]
	b, err := json.Marshal(jsve)
	if err != nil {
		return errors.Wrap(err, "failed to marshal JSON signed voluntary exit")
	}

	filepath := path.Join(outputDirectory, fmt.Sprintf("validator-exit-%s.json", jsve.Exit.ValidatorIndex))
	if err := file.WriteFile(filepath, b); err != nil {
		return errors.Wrap(err, "failed to write validator exist json")
	}

	log.Infof("Wrote signed validator exit JSON to %s", filepath)

	return nil
}
