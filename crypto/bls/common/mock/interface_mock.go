// Code generated by MockGen. DO NOT EDIT.
// Source: crypto/bls/common/interface.go

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	common "github.com/theQRL/qrysm/v4/crypto/bls/common"
)

// MockSecretKey is a mock of SecretKey interface.
type MockSecretKey struct {
	ctrl     *gomock.Controller
	recorder *MockSecretKeyMockRecorder
}

// MockSecretKeyMockRecorder is the mock recorder for MockSecretKey.
type MockSecretKeyMockRecorder struct {
	mock *MockSecretKey
}

// NewMockSecretKey creates a new mock instance.
func NewMockSecretKey(ctrl *gomock.Controller) *MockSecretKey {
	mock := &MockSecretKey{ctrl: ctrl}
	mock.recorder = &MockSecretKeyMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSecretKey) EXPECT() *MockSecretKeyMockRecorder {
	return m.recorder
}

// Marshal mocks base method.
func (m *MockSecretKey) Marshal() []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Marshal")
	ret0, _ := ret[0].([]byte)
	return ret0
}

// Marshal indicates an expected call of Marshal.
func (mr *MockSecretKeyMockRecorder) Marshal() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Marshal", reflect.TypeOf((*MockSecretKey)(nil).Marshal))
}

// PublicKey mocks base method.
func (m *MockSecretKey) PublicKey() common.PublicKey {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PublicKey")
	ret0, _ := ret[0].(common.PublicKey)
	return ret0
}

// PublicKey indicates an expected call of PublicKey.
func (mr *MockSecretKeyMockRecorder) PublicKey() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PublicKey", reflect.TypeOf((*MockSecretKey)(nil).PublicKey))
}

// Sign mocks base method.
func (m *MockSecretKey) Sign(msg []byte) common.Signature {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Sign", msg)
	ret0, _ := ret[0].(common.Signature)
	return ret0
}

// Sign indicates an expected call of Sign.
func (mr *MockSecretKeyMockRecorder) Sign(msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sign", reflect.TypeOf((*MockSecretKey)(nil).Sign), msg)
}

// MockPublicKey is a mock of PublicKey interface.
type MockPublicKey struct {
	ctrl     *gomock.Controller
	recorder *MockPublicKeyMockRecorder
}

// MockPublicKeyMockRecorder is the mock recorder for MockPublicKey.
type MockPublicKeyMockRecorder struct {
	mock *MockPublicKey
}

// NewMockPublicKey creates a new mock instance.
func NewMockPublicKey(ctrl *gomock.Controller) *MockPublicKey {
	mock := &MockPublicKey{ctrl: ctrl}
	mock.recorder = &MockPublicKeyMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPublicKey) EXPECT() *MockPublicKeyMockRecorder {
	return m.recorder
}

// Aggregate mocks base method.
func (m *MockPublicKey) Aggregate(p2 common.PublicKey) common.PublicKey {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Aggregate", p2)
	ret0, _ := ret[0].(common.PublicKey)
	return ret0
}

// Aggregate indicates an expected call of Aggregate.
func (mr *MockPublicKeyMockRecorder) Aggregate(p2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Aggregate", reflect.TypeOf((*MockPublicKey)(nil).Aggregate), p2)
}

// Copy mocks base method.
func (m *MockPublicKey) Copy() common.PublicKey {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Copy")
	ret0, _ := ret[0].(common.PublicKey)
	return ret0
}

// Copy indicates an expected call of Copy.
func (mr *MockPublicKeyMockRecorder) Copy() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Copy", reflect.TypeOf((*MockPublicKey)(nil).Copy))
}

// Equals mocks base method.
func (m *MockPublicKey) Equals(p2 common.PublicKey) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Equals", p2)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Equals indicates an expected call of Equals.
func (mr *MockPublicKeyMockRecorder) Equals(p2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Equals", reflect.TypeOf((*MockPublicKey)(nil).Equals), p2)
}

// IsInfinite mocks base method.
func (m *MockPublicKey) IsInfinite() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsInfinite")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsInfinite indicates an expected call of IsInfinite.
func (mr *MockPublicKeyMockRecorder) IsInfinite() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsInfinite", reflect.TypeOf((*MockPublicKey)(nil).IsInfinite))
}

// Marshal mocks base method.
func (m *MockPublicKey) Marshal() []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Marshal")
	ret0, _ := ret[0].([]byte)
	return ret0
}

// Marshal indicates an expected call of Marshal.
func (mr *MockPublicKeyMockRecorder) Marshal() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Marshal", reflect.TypeOf((*MockPublicKey)(nil).Marshal))
}

// MockSignature is a mock of Signature interface.
type MockSignature struct {
	ctrl     *gomock.Controller
	recorder *MockSignatureMockRecorder
}

// MockSignatureMockRecorder is the mock recorder for MockSignature.
type MockSignatureMockRecorder struct {
	mock *MockSignature
}

// NewMockSignature creates a new mock instance.
func NewMockSignature(ctrl *gomock.Controller) *MockSignature {
	mock := &MockSignature{ctrl: ctrl}
	mock.recorder = &MockSignatureMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSignature) EXPECT() *MockSignatureMockRecorder {
	return m.recorder
}

// AggregateVerify mocks base method.
func (m *MockSignature) AggregateVerify(pubKeys []common.PublicKey, msgs [][32]byte) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AggregateVerify", pubKeys, msgs)
	ret0, _ := ret[0].(bool)
	return ret0
}

// AggregateVerify indicates an expected call of AggregateVerify.
func (mr *MockSignatureMockRecorder) AggregateVerify(pubKeys, msgs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AggregateVerify", reflect.TypeOf((*MockSignature)(nil).AggregateVerify), pubKeys, msgs)
}

// Copy mocks base method.
func (m *MockSignature) Copy() common.Signature {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Copy")
	ret0, _ := ret[0].(common.Signature)
	return ret0
}

// Copy indicates an expected call of Copy.
func (mr *MockSignatureMockRecorder) Copy() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Copy", reflect.TypeOf((*MockSignature)(nil).Copy))
}

// Eth2FastAggregateVerify mocks base method.
func (m *MockSignature) Eth2FastAggregateVerify(pubKeys []common.PublicKey, msg [32]byte) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Eth2FastAggregateVerify", pubKeys, msg)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Eth2FastAggregateVerify indicates an expected call of Eth2FastAggregateVerify.
func (mr *MockSignatureMockRecorder) Eth2FastAggregateVerify(pubKeys, msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Eth2FastAggregateVerify", reflect.TypeOf((*MockSignature)(nil).Eth2FastAggregateVerify), pubKeys, msg)
}

// FastAggregateVerify mocks base method.
func (m *MockSignature) FastAggregateVerify(pubKeys []common.PublicKey, msg [32]byte) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FastAggregateVerify", pubKeys, msg)
	ret0, _ := ret[0].(bool)
	return ret0
}

// FastAggregateVerify indicates an expected call of FastAggregateVerify.
func (mr *MockSignatureMockRecorder) FastAggregateVerify(pubKeys, msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FastAggregateVerify", reflect.TypeOf((*MockSignature)(nil).FastAggregateVerify), pubKeys, msg)
}

// Marshal mocks base method.
func (m *MockSignature) Marshal() []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Marshal")
	ret0, _ := ret[0].([]byte)
	return ret0
}

// Marshal indicates an expected call of Marshal.
func (mr *MockSignatureMockRecorder) Marshal() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Marshal", reflect.TypeOf((*MockSignature)(nil).Marshal))
}

// Verify mocks base method.
func (m *MockSignature) Verify(pubKey common.PublicKey, msg []byte) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Verify", pubKey, msg)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Verify indicates an expected call of Verify.
func (mr *MockSignatureMockRecorder) Verify(pubKey, msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Verify", reflect.TypeOf((*MockSignature)(nil).Verify), pubKey, msg)
}