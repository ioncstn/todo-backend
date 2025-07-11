// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/todos/repository.go

// Package todos is a generated GoMock package.
package todos

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockRepository is a mock of Repository interface.
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance.
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// CountAll mocks base method.
func (m *MockRepository) CountAll(ctx context.Context) int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CountAll", ctx)
	ret0, _ := ret[0].(int)
	return ret0
}

// CountAll indicates an expected call of CountAll.
func (mr *MockRepositoryMockRecorder) CountAll(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CountAll", reflect.TypeOf((*MockRepository)(nil).CountAll), ctx)
}

// Create mocks base method.
func (m *MockRepository) Create(ctx context.Context, item *ToDoItem) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, item)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockRepositoryMockRecorder) Create(ctx, item interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockRepository)(nil).Create), ctx, item)
}

// Delete mocks base method.
func (m *MockRepository) Delete(ctx context.Context, id uint) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockRepositoryMockRecorder) Delete(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockRepository)(nil).Delete), ctx, id)
}

// GetAll mocks base method.
func (m *MockRepository) GetAll(ctx context.Context, details PaginationDetails) ([]ToDoItem, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll", ctx, details)
	ret0, _ := ret[0].([]ToDoItem)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAll indicates an expected call of GetAll.
func (mr *MockRepositoryMockRecorder) GetAll(ctx, details interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockRepository)(nil).GetAll), ctx, details)
}

// GetById mocks base method.
func (m *MockRepository) GetById(ctx context.Context, id uint) (ToDoItem, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetById", ctx, id)
	ret0, _ := ret[0].(ToDoItem)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetById indicates an expected call of GetById.
func (mr *MockRepositoryMockRecorder) GetById(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetById", reflect.TypeOf((*MockRepository)(nil).GetById), ctx, id)
}

// Update mocks base method.
func (m *MockRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, id, updates)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockRepositoryMockRecorder) Update(ctx, id, updates interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockRepository)(nil).Update), ctx, id, updates)
}
