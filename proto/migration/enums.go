package migration

import (
	"github.com/pkg/errors"
	zond "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
	zondpb "github.com/theQRL/qrysm/v4/proto/zond/v1"
)

func V1Alpha1ToV1ConnectionState(connState zond.ConnectionState) zondpb.ConnectionState {
	alphaString := connState.String()
	v1Value := zondpb.ConnectionState_value[alphaString]
	return zondpb.ConnectionState(v1Value)
}

func V1Alpha1ToV1PeerDirection(peerDirection zond.PeerDirection) (zondpb.PeerDirection, error) {
	alphaString := peerDirection.String()
	if alphaString == zond.PeerDirection_UNKNOWN.String() {
		return 0, errors.New("peer direction unknown")
	}
	v1Value := zondpb.PeerDirection_value[alphaString]
	return zondpb.PeerDirection(v1Value), nil
}
