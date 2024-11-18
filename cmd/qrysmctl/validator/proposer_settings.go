package validator

import (
	"encoding/json"
	"errors"
	"io"
	"regexp"

	log "github.com/sirupsen/logrus"
	"github.com/theQRL/go-zond/common"
	"github.com/theQRL/qrysm/api/client"
	"github.com/theQRL/qrysm/api/client/validator"
	"github.com/theQRL/qrysm/cmd/validator/flags"
	"github.com/theQRL/qrysm/config/params"
	validatorType "github.com/theQRL/qrysm/consensus-types/validator"
	"github.com/theQRL/qrysm/io/file"
	"github.com/theQRL/qrysm/io/prompt"
	validatorpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1/validator-client"
	"github.com/urfave/cli/v2"
	"go.opencensus.io/trace"
)

var addressRegex = regexp.MustCompile("^Z[0-9a-fA-F]{40}$")

func getProposerSettings(c *cli.Context, r io.Reader) error {
	ctx, span := trace.StartSpan(c.Context, "qrysmctl.getProposerSettings")
	defer span.End()
	if !c.IsSet(ValidatorHostFlag.Name) {
		return errNoFlag(ValidatorHostFlag.Name)
	}
	if !c.IsSet(TokenFlag.Name) {
		return errNoFlag(TokenFlag.Name)
	}
	defaultFeeRecipient := params.BeaconConfig().DefaultFeeRecipient.Hex()
	if c.IsSet(ProposerSettingsOutputFlag.Name) {
		if c.IsSet(DefaultFeeRecipientFlag.Name) {
			recipient := c.String(DefaultFeeRecipientFlag.Name)
			if err := validateIsExecutionAddress(recipient); err != nil {
				return err
			}
			defaultFeeRecipient = recipient
		} else {
			promptText := "Please enter a default fee recipient address (a zond address in hex format)"
			resp, err := prompt.ValidatePrompt(r, promptText, validateIsExecutionAddress)
			if err != nil {
				return err
			}
			defaultFeeRecipient = resp
		}
	}

	cl, err := validator.NewClient(c.String(ValidatorHostFlag.Name), client.WithAuthenticationToken(c.String(TokenFlag.Name)))
	if err != nil {
		return err
	}
	validators, err := cl.GetValidatorPubKeys(ctx)
	if err != nil {
		return err
	}
	feeRecipients, err := cl.GetFeeRecipientAddresses(ctx, validators)
	if err != nil {
		return err
	}

	log.Infoln("===============DISPLAYING CURRENT PROPOSER SETTINGS===============")

	for index := range validators {
		log.Infof("Validator: %s. Fee-recipient: %s", validators[index], feeRecipients[index])
	}

	if c.IsSet(ProposerSettingsOutputFlag.Name) {
		log.Infof("The default fee recipient is set to %s", defaultFeeRecipient)
		var builderSettings *validatorpb.BuilderConfig
		if c.Bool(WithBuilderFlag.Name) {
			builderSettings = &validatorpb.BuilderConfig{
				Enabled:  true,
				GasLimit: validatorType.Uint64(params.BeaconConfig().DefaultBuilderGasLimit),
			}
		} else {
			log.Infof("Default builder settings can be included with the `--%s` flag", WithBuilderFlag.Name)
		}
		proposerConfig := make(map[string]*validatorpb.ProposerOptionPayload)
		for index, val := range validators {
			proposerConfig[val] = &validatorpb.ProposerOptionPayload{
				FeeRecipient: feeRecipients[index],
				Builder:      builderSettings,
			}
		}
		fileConfig := &validatorpb.ProposerSettingsPayload{
			ProposerConfig: proposerConfig,
			DefaultConfig: &validatorpb.ProposerOptionPayload{
				FeeRecipient: defaultFeeRecipient,
				Builder:      builderSettings,
			},
		}
		b, err := json.Marshal(fileConfig)
		if err != nil {
			return err
		}
		if err := file.WriteFile(c.String(ProposerSettingsOutputFlag.Name), b); err != nil {
			return err
		}
		log.Infof("Successfully created `%s`. Settings can be imported into validator client using --%s flag.", c.String(ProposerSettingsOutputFlag.Name), flags.ProposerSettingsFlag.Name)
	}

	return nil
}

func validateIsExecutionAddress(input string) error {
	if !IsAddress([]byte(input)) || !(len(input) == common.AddressLength*2+1) {
		return errors.New("no default address entered")
	}
	return nil
}

// IsAddress checks whether the byte array is a hex number prefixed with 'Z'.
func IsAddress(b []byte) bool {
	if b == nil {
		return false
	}
	return addressRegex.Match(b)
}
