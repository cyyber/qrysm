package signing

import (
	"github.com/pkg/errors"
	"github.com/theQRL/qrysm/config/params"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
)

var ErrNilRegistration = errors.New("nil signed registration")

// VerifyRegistrationSignature verifies the signature of a validator's registration.
func VerifyRegistrationSignature(
	sr *qrysmpb.SignedValidatorRegistrationV1,
) error {
	if sr == nil || sr.Message == nil {
		return ErrNilRegistration
	}

	d := params.BeaconConfig().DomainApplicationBuilder
	// Per spec, we want the fork version and genesis validator to be nil.
	// Which is genesis value and zero by default.
	sd, err := ComputeDomain(
		d,
		nil, /* fork version */
		nil /* genesis val root */)
	if err != nil {
		return err
	}

	if err := VerifySigningRoot(sr.Message, sr.Message.Pubkey, sr.Signature, sd); err != nil {
		return ErrSigFailedToVerify
	}
	return nil
}
