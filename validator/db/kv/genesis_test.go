package kv

import (
	"context"
	"testing"

	"github.com/cyyber/qrysm/v4/config/params"
	"github.com/cyyber/qrysm/v4/testing/require"
	dilithium2 "github.com/theQRL/go-qrllib/dilithium"
)

func TestStore_GenesisValidatorsRoot_ReadAndWrite(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t, [][dilithium2.CryptoPublicKeyBytes]byte{})
	tests := []struct {
		name    string
		want    []byte
		write   []byte
		wantErr bool
	}{
		{
			name:  "empty then write",
			want:  nil,
			write: params.BeaconConfig().ZeroHash[:],
		},
		{
			name:    "zero then overwrite rejected",
			want:    params.BeaconConfig().ZeroHash[:],
			write:   []byte{5},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.GenesisValidatorsRoot(ctx)
			require.NoError(t, err)
			require.DeepEqual(t, tt.want, got)
			err = db.SaveGenesisValidatorsRoot(ctx, tt.write)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenesisValidatorsRoot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
