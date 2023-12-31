// Code generated by MockGen. DO NOT EDIT.
// Source: controllers/search.go

// Package mock_controllers is a generated GoMock package.
package mock_controllers

import (
	reflect "reflect"

	v2 "github.com/gofiber/fiber/v2"
	gomock "github.com/golang/mock/gomock"
)

// MockISearchController is a mock of ISearchController interface.
type MockISearchController struct {
	ctrl     *gomock.Controller
	recorder *MockISearchControllerMockRecorder
}

// MockISearchControllerMockRecorder is the mock recorder for MockISearchController.
type MockISearchControllerMockRecorder struct {
	mock *MockISearchController
}

// NewMockISearchController creates a new mock instance.
func NewMockISearchController(ctrl *gomock.Controller) *MockISearchController {
	mock := &MockISearchController{ctrl: ctrl}
	mock.recorder = &MockISearchControllerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockISearchController) EXPECT() *MockISearchControllerMockRecorder {
	return m.recorder
}

// Search mocks base method.
func (m *MockISearchController) Search(c *v2.Ctx) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Search", c)
	ret0, _ := ret[0].(error)
	return ret0
}

// Search indicates an expected call of Search.
func (mr *MockISearchControllerMockRecorder) Search(c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Search", reflect.TypeOf((*MockISearchController)(nil).Search), c)
}
