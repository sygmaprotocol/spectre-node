// Code generated by MockGen. DO NOT EDIT.
// Source: ./chains/evm/listener/listener.go
//
// Generated by this command:
//
//	mockgen -source=./chains/evm/listener/listener.go -destination=./mock/listener.go -package mock
//
// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	big "math/big"
	reflect "reflect"

	api "github.com/attestantio/go-eth2-client/api"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	gomock "go.uber.org/mock/gomock"
)

// MockEventHandler is a mock of EventHandler interface.
type MockEventHandler struct {
	ctrl     *gomock.Controller
	recorder *MockEventHandlerMockRecorder
}

// MockEventHandlerMockRecorder is the mock recorder for MockEventHandler.
type MockEventHandlerMockRecorder struct {
	mock *MockEventHandler
}

// NewMockEventHandler creates a new mock instance.
func NewMockEventHandler(ctrl *gomock.Controller) *MockEventHandler {
	mock := &MockEventHandler{ctrl: ctrl}
	mock.recorder = &MockEventHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEventHandler) EXPECT() *MockEventHandlerMockRecorder {
	return m.recorder
}

// HandleEvents mocks base method.
func (m *MockEventHandler) HandleEvents(checkpoint *v1.Finality) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HandleEvents", checkpoint)
	ret0, _ := ret[0].(error)
	return ret0
}

// HandleEvents indicates an expected call of HandleEvents.
func (mr *MockEventHandlerMockRecorder) HandleEvents(checkpoint any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleEvents", reflect.TypeOf((*MockEventHandler)(nil).HandleEvents), checkpoint)
}

// MockBeaconProvider is a mock of BeaconProvider interface.
type MockBeaconProvider struct {
	ctrl     *gomock.Controller
	recorder *MockBeaconProviderMockRecorder
}

// MockBeaconProviderMockRecorder is the mock recorder for MockBeaconProvider.
type MockBeaconProviderMockRecorder struct {
	mock *MockBeaconProvider
}

// NewMockBeaconProvider creates a new mock instance.
func NewMockBeaconProvider(ctrl *gomock.Controller) *MockBeaconProvider {
	mock := &MockBeaconProvider{ctrl: ctrl}
	mock.recorder = &MockBeaconProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBeaconProvider) EXPECT() *MockBeaconProviderMockRecorder {
	return m.recorder
}

// Finality mocks base method.
func (m *MockBeaconProvider) Finality(ctx context.Context, opts *api.FinalityOpts) (*api.Response[*v1.Finality], error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Finality", ctx, opts)
	ret0, _ := ret[0].(*api.Response[*v1.Finality])
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Finality indicates an expected call of Finality.
func (mr *MockBeaconProviderMockRecorder) Finality(ctx, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Finality", reflect.TypeOf((*MockBeaconProvider)(nil).Finality), ctx, opts)
}

// MockBlockStorer is a mock of BlockStorer interface.
type MockBlockStorer struct {
	ctrl     *gomock.Controller
	recorder *MockBlockStorerMockRecorder
}

// MockBlockStorerMockRecorder is the mock recorder for MockBlockStorer.
type MockBlockStorerMockRecorder struct {
	mock *MockBlockStorer
}

// NewMockBlockStorer creates a new mock instance.
func NewMockBlockStorer(ctrl *gomock.Controller) *MockBlockStorer {
	mock := &MockBlockStorer{ctrl: ctrl}
	mock.recorder = &MockBlockStorerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBlockStorer) EXPECT() *MockBlockStorerMockRecorder {
	return m.recorder
}

// StoreBlock mocks base method.
func (m *MockBlockStorer) StoreBlock(epoch *big.Int, domainID uint8) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StoreBlock", epoch, domainID)
	ret0, _ := ret[0].(error)
	return ret0
}

// StoreBlock indicates an expected call of StoreBlock.
func (mr *MockBlockStorerMockRecorder) StoreBlock(epoch, domainID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreBlock", reflect.TypeOf((*MockBlockStorer)(nil).StoreBlock), epoch, domainID)
}
