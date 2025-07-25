package kv

import (
	"context"
	"testing"

	v1alpha1 "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
)

func TestStore_SaveExecutionChainData(t *testing.T) {
	type args struct {
		data *v1alpha1.ExecutionChainData
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil data",
			args: args{
				data: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := setupDB(t)
			if err := store.SaveExecutionChainData(context.Background(), tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("SaveExecutionData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
