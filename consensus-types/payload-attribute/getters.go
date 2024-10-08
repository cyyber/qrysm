package payloadattribute

import (
	enginev1 "github.com/theQRL/qrysm/proto/engine/v1"
)

// Version returns the version of the payload attribute.
func (a *data) Version() int {
	return a.version
}

// PrevRandao returns the previous randao value of the payload attribute.
func (a *data) PrevRandao() []byte {
	return a.prevRandao
}

// Timestamps returns the timestamp of the payload attribute.
func (a *data) Timestamps() uint64 {
	return a.timeStamp
}

// SuggestedFeeRecipient returns the suggested fee recipient of the payload attribute.
func (a *data) SuggestedFeeRecipient() []byte {
	return a.suggestedFeeRecipient
}

// Withdrawals returns the withdrawals of the payload attribute.
// Support for withdrawals was added in version 2 of the payload attribute.
func (a *data) Withdrawals() ([]*enginev1.Withdrawal, error) {
	if a == nil {
		return nil, errNilPayloadAttribute
	}
	return a.withdrawals, nil
}

// PbV2 returns the payload attribute in version 2.
func (a *data) PbV2() (*enginev1.PayloadAttributesV2, error) {
	if a == nil {
		return nil, errNilPayloadAttribute
	}
	if a.timeStamp == 0 && len(a.prevRandao) == 0 {
		return nil, nil
	}
	return &enginev1.PayloadAttributesV2{
		Timestamp:             a.timeStamp,
		PrevRandao:            a.prevRandao,
		SuggestedFeeRecipient: a.suggestedFeeRecipient,
		Withdrawals:           a.withdrawals,
	}, nil
}
