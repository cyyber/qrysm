package migration

import (
	"testing"

	qrlpb "github.com/theQRL/qrysm/proto/qrl/v1"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
)

func TestV1Alpha1ConnectionStateToV1(t *testing.T) {
	tests := []struct {
		name      string
		connState qrysmpb.ConnectionState
		want      qrlpb.ConnectionState
	}{
		{
			name:      "DISCONNECTED",
			connState: qrysmpb.ConnectionState_DISCONNECTED,
			want:      qrlpb.ConnectionState_DISCONNECTED,
		},
		{
			name:      "CONNECTED",
			connState: qrysmpb.ConnectionState_CONNECTED,
			want:      qrlpb.ConnectionState_CONNECTED,
		},
		{
			name:      "CONNECTING",
			connState: qrysmpb.ConnectionState_CONNECTING,
			want:      qrlpb.ConnectionState_CONNECTING,
		},
		{
			name:      "DISCONNECTING",
			connState: qrysmpb.ConnectionState_DISCONNECTING,
			want:      qrlpb.ConnectionState_DISCONNECTING,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := V1Alpha1ConnectionStateToV1(tt.connState); got != tt.want {
				t.Errorf("V1Alpha1ConnectionStateToV1() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestV1Alpha1PeerDirectionToV1(t *testing.T) {
	tests := []struct {
		name          string
		peerDirection qrysmpb.PeerDirection
		want          qrlpb.PeerDirection
		wantErr       bool
	}{
		{
			name:          "UNKNOWN",
			peerDirection: qrysmpb.PeerDirection_UNKNOWN,
			want:          0,
			wantErr:       true,
		},
		{
			name:          "INBOUND",
			peerDirection: qrysmpb.PeerDirection_INBOUND,
			want:          qrlpb.PeerDirection_INBOUND,
		},
		{
			name:          "OUTBOUND",
			peerDirection: qrysmpb.PeerDirection_OUTBOUND,
			want:          qrlpb.PeerDirection_OUTBOUND,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := V1Alpha1PeerDirectionToV1(tt.peerDirection)
			if (err != nil) != tt.wantErr {
				t.Errorf("V1Alpha1PeerDirectionToV1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("V1Alpha1PeerDirectionToV1() got = %v, want %v", got, tt.want)
			}
		})
	}
}
