// Code generated by MockGen. DO NOT EDIT.
// Source: ./chains/evm/executor/executor.go
//
// Generated by this command:
//
//	mockgen -source=./chains/evm/executor/executor.go -destination=./mock/executor.go -package mock
//
// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	common "github.com/ethereum/go-ethereum/common"
	message "github.com/sygmaprotocol/spectre-node/chains/evm/message"
	transactor "github.com/sygmaprotocol/sygma-core/chains/evm/transactor"
	gomock "go.uber.org/mock/gomock"
)

// MockProofSubmitter is a mock of ProofSubmitter interface.
type MockProofSubmitter struct {
	ctrl     *gomock.Controller
	recorder *MockProofSubmitterMockRecorder
}

// MockProofSubmitterMockRecorder is the mock recorder for MockProofSubmitter.
type MockProofSubmitterMockRecorder struct {
	mock *MockProofSubmitter
}

// NewMockProofSubmitter creates a new mock instance.
func NewMockProofSubmitter(ctrl *gomock.Controller) *MockProofSubmitter {
	mock := &MockProofSubmitter{ctrl: ctrl}
	mock.recorder = &MockProofSubmitterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProofSubmitter) EXPECT() *MockProofSubmitterMockRecorder {
	return m.recorder
}

// Rotate mocks base method.
func (m *MockProofSubmitter) Rotate(domainID uint8, rotateProof []byte, stepInput message.SyncStepInput, stepProof []byte, opts transactor.TransactOptions) (*common.Hash, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Rotate", domainID, rotateProof, stepInput, stepProof, opts)
	ret0, _ := ret[0].(*common.Hash)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Rotate indicates an expected call of Rotate.
func (mr *MockProofSubmitterMockRecorder) Rotate(domainID, rotateProof, stepInput, stepProof, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rotate", reflect.TypeOf((*MockProofSubmitter)(nil).Rotate), domainID, rotateProof, stepInput, stepProof, opts)
}

// Step mocks base method.
func (m *MockProofSubmitter) Step(domainID uint8, input message.SyncStepInput, stepProof []byte, stateRoot [32]byte, stateRootProof [][]byte, opts transactor.TransactOptions) (*common.Hash, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Step", domainID, input, stepProof, stateRoot, stateRootProof, opts)
	ret0, _ := ret[0].(*common.Hash)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Step indicates an expected call of Step.
func (mr *MockProofSubmitterMockRecorder) Step(domainID, input, stepProof, stateRoot, stateRootProof, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Step", reflect.TypeOf((*MockProofSubmitter)(nil).Step), domainID, input, stepProof, stateRoot, stateRootProof, opts)
}
