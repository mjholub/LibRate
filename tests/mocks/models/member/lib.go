// Code generated by MockGen. DO NOT EDIT.
// Source: models/member/lib.go
//
// Generated by this command:
//
//	mockgen -source models/member/lib.go -destination tests/mocks/models/member/lib.go
//
// Package mock_member is a generated GoMock package.
package mock_member

import (
	context "context"
	reflect "reflect"

	member "codeberg.org/mjh/LibRate/models/member"
	v5 "github.com/gofrs/uuid/v5"
	gomock "go.uber.org/mock/gomock"
)

// MockStorer is a mock of Storer interface.
type MockStorer struct {
	ctrl     *gomock.Controller
	recorder *MockStorerMockRecorder
}

// MockStorerMockRecorder is the mock recorder for MockStorer.
type MockStorerMockRecorder struct {
	mock *MockStorer
}

// NewMockStorer creates a new mock instance.
func NewMockStorer(ctrl *gomock.Controller) *MockStorer {
	mock := &MockStorer{ctrl: ctrl}
	mock.recorder = &MockStorerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorer) EXPECT() *MockStorerMockRecorder {
	return m.recorder
}

// Check mocks base method.
func (m *MockStorer) Check(ctx context.Context, email, nickname string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Check", ctx, email, nickname)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Check indicates an expected call of Check.
func (mr *MockStorerMockRecorder) Check(ctx, email, nickname any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Check", reflect.TypeOf((*MockStorer)(nil).Check), ctx, email, nickname)
}

// CreateSession mocks base method.
func (m *MockStorer) CreateSession(ctx context.Context, member *member.Member) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateSession", ctx, member)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateSession indicates an expected call of CreateSession.
func (mr *MockStorerMockRecorder) CreateSession(ctx, member any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSession", reflect.TypeOf((*MockStorer)(nil).CreateSession), ctx, member)
}

// Delete mocks base method.
func (m *MockStorer) Delete(ctx context.Context, member *member.Member) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, member)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockStorerMockRecorder) Delete(ctx, member any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockStorer)(nil).Delete), ctx, member)
}

// GetID mocks base method.
func (m *MockStorer) GetID(ctx context.Context, key string) (uint32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetID", ctx, key)
	ret0, _ := ret[0].(uint32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetID indicates an expected call of GetID.
func (mr *MockStorerMockRecorder) GetID(ctx, key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetID", reflect.TypeOf((*MockStorer)(nil).GetID), ctx, key)
}

// GetPassHash mocks base method.
func (m *MockStorer) GetPassHash(email, login string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPassHash", email, login)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPassHash indicates an expected call of GetPassHash.
func (mr *MockStorerMockRecorder) GetPassHash(email, login any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPassHash", reflect.TypeOf((*MockStorer)(nil).GetPassHash), email, login)
}

// GetSessionTimeout mocks base method.
func (m *MockStorer) GetSessionTimeout(ctx context.Context, memberID int, deviceID v5.UUID) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSessionTimeout", ctx, memberID, deviceID)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSessionTimeout indicates an expected call of GetSessionTimeout.
func (mr *MockStorerMockRecorder) GetSessionTimeout(ctx, memberID, deviceID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSessionTimeout", reflect.TypeOf((*MockStorer)(nil).GetSessionTimeout), ctx, memberID, deviceID)
}

// LookupDevice mocks base method.
func (m *MockStorer) LookupDevice(ctx context.Context, deviceID v5.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LookupDevice", ctx, deviceID)
	ret0, _ := ret[0].(error)
	return ret0
}

// LookupDevice indicates an expected call of LookupDevice.
func (mr *MockStorerMockRecorder) LookupDevice(ctx, deviceID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LookupDevice", reflect.TypeOf((*MockStorer)(nil).LookupDevice), ctx, deviceID)
}

// Read mocks base method.
func (m *MockStorer) Read(ctx context.Context, key string, keyNames ...string) (*member.Member, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, key}
	for _, a := range keyNames {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Read", varargs...)
	ret0, _ := ret[0].(*member.Member)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *MockStorerMockRecorder) Read(ctx, key any, keyNames ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, key}, keyNames...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockStorer)(nil).Read), varargs...)
}

// RequestFollow mocks base method.
func (m *MockStorer) RequestFollow(ctx context.Context, fr *member.FollowBlockRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RequestFollow", ctx, fr)
	ret0, _ := ret[0].(error)
	return ret0
}

// RequestFollow indicates an expected call of RequestFollow.
func (mr *MockStorerMockRecorder) RequestFollow(ctx, fr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequestFollow", reflect.TypeOf((*MockStorer)(nil).RequestFollow), ctx, fr)
}

// Save mocks base method.
func (m *MockStorer) Save(ctx context.Context, member *member.Member) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", ctx, member)
	ret0, _ := ret[0].(error)
	return ret0
}

// Save indicates an expected call of Save.
func (mr *MockStorerMockRecorder) Save(ctx, member any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockStorer)(nil).Save), ctx, member)
}

// Update mocks base method.
func (m *MockStorer) Update(ctx context.Context, member *member.Member) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, member)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockStorerMockRecorder) Update(ctx, member any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockStorer)(nil).Update), ctx, member)
}
