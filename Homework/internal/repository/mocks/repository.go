// Code generated by MockGen. DO NOT EDIT.
// Source: ./repository.go

// Package mock_repository is a generated GoMock package.
package mock_repository

import (
	repository "Homework/internal/repository"
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockPickupPointRepo is a mock of PickupPointRepo interface.
type MockPickupPointRepo struct {
	ctrl     *gomock.Controller
	recorder *MockPickupPointRepoMockRecorder
}

// MockPickupPointRepoMockRecorder is the mock recorder for MockPickupPointRepo.
type MockPickupPointRepoMockRecorder struct {
	mock *MockPickupPointRepo
}

// NewMockPickupPointRepo creates a new mock instance.
func NewMockPickupPointRepo(ctrl *gomock.Controller) *MockPickupPointRepo {
	mock := &MockPickupPointRepo{ctrl: ctrl}
	mock.recorder = &MockPickupPointRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPickupPointRepo) EXPECT() *MockPickupPointRepoMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockPickupPointRepo) Add(ctx context.Context, pickupPoint repository.PickupPoint) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", ctx, pickupPoint)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Add indicates an expected call of Add.
func (mr *MockPickupPointRepoMockRecorder) Add(ctx, pickupPoint interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockPickupPointRepo)(nil).Add), ctx, pickupPoint)
}

// Delete mocks base method.
func (m *MockPickupPointRepo) Delete(ctx context.Context, id int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockPickupPointRepoMockRecorder) Delete(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockPickupPointRepo)(nil).Delete), ctx, id)
}

// GetByID mocks base method.
func (m *MockPickupPointRepo) GetByID(ctx context.Context, id int64) (repository.PickupPoint, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(repository.PickupPoint)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MockPickupPointRepoMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockPickupPointRepo)(nil).GetByID), ctx, id)
}

// List mocks base method.
func (m *MockPickupPointRepo) List(ctx context.Context) ([]repository.PickupPoint, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx)
	ret0, _ := ret[0].([]repository.PickupPoint)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockPickupPointRepoMockRecorder) List(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockPickupPointRepo)(nil).List), ctx)
}

// Update mocks base method.
func (m *MockPickupPointRepo) Update(ctx context.Context, pickupPoint repository.PickupPoint) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, pickupPoint)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockPickupPointRepoMockRecorder) Update(ctx, pickupPoint interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockPickupPointRepo)(nil).Update), ctx, pickupPoint)
}
