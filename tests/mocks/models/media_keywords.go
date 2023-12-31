// Code generated by MockGen. DO NOT EDIT.
// Source: models/media_keywords.go

// Package mock_models is a generated GoMock package.
package mock_models

import (
	context "context"
	reflect "reflect"

	models "codeberg.org/mjh/LibRate/models"
	v5 "github.com/gofrs/uuid/v5"
	gomock "github.com/golang/mock/gomock"
)

// MockKeywordStorer is a mock of KeywordStorer interface.
type MockKeywordStorer struct {
	ctrl     *gomock.Controller
	recorder *MockKeywordStorerMockRecorder
}

// MockKeywordStorerMockRecorder is the mock recorder for MockKeywordStorer.
type MockKeywordStorerMockRecorder struct {
	mock *MockKeywordStorer
}

// NewMockKeywordStorer creates a new mock instance.
func NewMockKeywordStorer(ctrl *gomock.Controller) *MockKeywordStorer {
	mock := &MockKeywordStorer{ctrl: ctrl}
	mock.recorder = &MockKeywordStorerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockKeywordStorer) EXPECT() *MockKeywordStorerMockRecorder {
	return m.recorder
}

// AddKeyword mocks base method.
func (m *MockKeywordStorer) AddKeyword(ctx context.Context, k models.Keyword) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddKeyword", ctx, k)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddKeyword indicates an expected call of AddKeyword.
func (mr *MockKeywordStorerMockRecorder) AddKeyword(ctx, k interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddKeyword", reflect.TypeOf((*MockKeywordStorer)(nil).AddKeyword), ctx, k)
}

// CastVote mocks base method.
func (m *MockKeywordStorer) CastVote(ctx context.Context, k models.Keyword) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CastVote", ctx, k)
	ret0, _ := ret[0].(error)
	return ret0
}

// CastVote indicates an expected call of CastVote.
func (mr *MockKeywordStorerMockRecorder) CastVote(ctx, k interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CastVote", reflect.TypeOf((*MockKeywordStorer)(nil).CastVote), ctx, k)
}

// GetKeyword mocks base method.
func (m *MockKeywordStorer) GetKeyword(ctx context.Context, mediaID v5.UUID) (models.Keyword, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetKeyword", ctx, mediaID)
	ret0, _ := ret[0].(models.Keyword)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetKeyword indicates an expected call of GetKeyword.
func (mr *MockKeywordStorerMockRecorder) GetKeyword(ctx, mediaID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetKeyword", reflect.TypeOf((*MockKeywordStorer)(nil).GetKeyword), ctx, mediaID)
}

// GetKeywords mocks base method.
func (m *MockKeywordStorer) GetKeywords(ctx context.Context, mediaID v5.UUID) ([]models.Keyword, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetKeywords", ctx, mediaID)
	ret0, _ := ret[0].([]models.Keyword)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetKeywords indicates an expected call of GetKeywords.
func (mr *MockKeywordStorerMockRecorder) GetKeywords(ctx, mediaID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetKeywords", reflect.TypeOf((*MockKeywordStorer)(nil).GetKeywords), ctx, mediaID)
}

// RemoveVote mocks base method.
func (m *MockKeywordStorer) RemoveVote(ctx context.Context, k models.Keyword) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveVote", ctx, k)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveVote indicates an expected call of RemoveVote.
func (mr *MockKeywordStorerMockRecorder) RemoveVote(ctx, k interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveVote", reflect.TypeOf((*MockKeywordStorer)(nil).RemoveVote), ctx, k)
}
