package kv

import (
	"context"
	"testing"

	"github.com/theQRL/go-zond/common"
	"github.com/theQRL/go-zond/common/hexutil"
	field_params "github.com/theQRL/qrysm/config/fieldparams"
	"github.com/theQRL/qrysm/config/params"
	validatorServiceConfig "github.com/theQRL/qrysm/config/validator/service"
	"github.com/theQRL/qrysm/consensus-types/validator"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	"github.com/theQRL/qrysm/testing/require"
)

func TestStore_ProposerSettings_ReadAndWrite(t *testing.T) {
	recipient0, err := common.NewAddressFromString("Z50155530FCE8a85ec7055A5F8b2bE214B3DaeFd3")
	require.NoError(t, err)
	recipient1, err := common.NewAddressFromString("Z6e35733c5af9B61374A128e6F85f553aF09ff89A")
	require.NoError(t, err)
	recipient2, err := common.NewAddressFromString("Z9995733c5af9B61374A128e6F85f553aF09ff89B")
	require.NoError(t, err)
	t.Run("save to db in full", func(t *testing.T) {
		ctx := context.Background()
		db := setupDB(t, [][field_params.DilithiumPubkeyLength]byte{})
		key1, err := hexutil.Decode("0xa057816155ad77931185101128655c0191bd0214c201ca48ed887f6c4c6adf334070efcd75140eada5ac83a92506dd7a")
		require.NoError(t, err)
		settings := &validatorServiceConfig.ProposerSettings{
			ProposeConfig: map[[field_params.DilithiumPubkeyLength]byte]*validatorServiceConfig.ProposerOption{
				bytesutil.ToBytes2592(key1): {
					FeeRecipientConfig: &validatorServiceConfig.FeeRecipientConfig{
						FeeRecipient: recipient0,
					},
					BuilderConfig: &validatorServiceConfig.BuilderConfig{
						Enabled:  true,
						GasLimit: validator.Uint64(40000000),
					},
				},
			},
			DefaultConfig: &validatorServiceConfig.ProposerOption{
				FeeRecipientConfig: &validatorServiceConfig.FeeRecipientConfig{
					FeeRecipient: recipient1,
				},
				BuilderConfig: &validatorServiceConfig.BuilderConfig{
					Enabled:  false,
					GasLimit: validator.Uint64(params.BeaconConfig().DefaultBuilderGasLimit),
				},
			},
		}
		err = db.SaveProposerSettings(ctx, settings)
		require.NoError(t, err)

		dbSettings, err := db.ProposerSettings(ctx)
		require.NoError(t, err)
		require.DeepEqual(t, settings, dbSettings)
	})
	t.Run("update default settings then update at specific key", func(t *testing.T) {
		ctx := context.Background()
		db := setupDB(t, [][field_params.DilithiumPubkeyLength]byte{})
		key1, err := hexutil.Decode("0xa057816155ad77931185101128655c0191bd0214c201ca48ed887f6c4c6adf334070efcd75140eada5ac83a92506dd7a")
		require.NoError(t, err)
		settings := &validatorServiceConfig.ProposerSettings{
			DefaultConfig: &validatorServiceConfig.ProposerOption{
				FeeRecipientConfig: &validatorServiceConfig.FeeRecipientConfig{
					FeeRecipient: recipient1,
				},
				BuilderConfig: &validatorServiceConfig.BuilderConfig{
					Enabled:  false,
					GasLimit: validator.Uint64(params.BeaconConfig().DefaultBuilderGasLimit),
				},
			},
		}
		err = db.SaveProposerSettings(ctx, settings)
		require.NoError(t, err)
		upatedDefault := &validatorServiceConfig.ProposerOption{
			FeeRecipientConfig: &validatorServiceConfig.FeeRecipientConfig{
				FeeRecipient: recipient2,
			},
			BuilderConfig: &validatorServiceConfig.BuilderConfig{
				Enabled:  true,
				GasLimit: validator.Uint64(params.BeaconConfig().DefaultBuilderGasLimit),
			},
		}
		err = db.UpdateProposerSettingsDefault(ctx, upatedDefault)
		require.NoError(t, err)

		dbSettings, err := db.ProposerSettings(ctx)
		require.NoError(t, err)
		require.NotNil(t, dbSettings)
		require.DeepEqual(t, dbSettings.DefaultConfig, upatedDefault)
		option := &validatorServiceConfig.ProposerOption{
			FeeRecipientConfig: &validatorServiceConfig.FeeRecipientConfig{
				FeeRecipient: recipient0,
			},
			BuilderConfig: &validatorServiceConfig.BuilderConfig{
				Enabled:  true,
				GasLimit: validator.Uint64(40000000),
			},
		}
		err = db.UpdateProposerSettingsForPubkey(ctx, bytesutil.ToBytes2592(key1), option)
		require.NoError(t, err)

		newSettings, err := db.ProposerSettings(ctx)
		require.NoError(t, err)
		require.NotNil(t, newSettings)
		require.DeepEqual(t, newSettings.DefaultConfig, upatedDefault)
		op, ok := newSettings.ProposeConfig[bytesutil.ToBytes2592(key1)]
		require.Equal(t, ok, true)
		require.DeepEqual(t, op, option)
	})
}
