// Code generated by MockGen. DO NOT EDIT.
// Source: controllers/review.go

// Package mock_controllers is a generated GoMock package.
package mock_controllers

import (
	reflect "reflect"

	v2 "github.com/gofiber/fiber/v2"
	gomock "github.com/golang/mock/gomock"
)

// MockIReviewController is a mock of IReviewController interface.
type MockIReviewController struct {
	ctrl     *gomock.Controller
	recorder *MockIReviewControllerMockRecorder
}

// MockIReviewControllerMockRecorder is the mock recorder for MockIReviewController.
type MockIReviewControllerMockRecorder struct {
	mock *MockIReviewController
}

// NewMockIReviewController creates a new mock instance.
func NewMockIReviewController(ctrl *gomock.Controller) *MockIReviewController {
	mock := &MockIReviewController{ctrl: ctrl}
	mock.recorder = &MockIReviewControllerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIReviewController) EXPECT() *MockIReviewControllerMockRecorder {
	return m.recorder
}

// GetAverageRatings mocks base method.
func (m *MockIReviewController) GetAverageRatings(c *v2.Ctx) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAverageRatings", c)
	ret0, _ := ret[0].(error)
	return ret0
}

// GetAverageRatings indicates an expected call of GetAverageRatings.
func (mr *MockIReviewControllerMockRecorder) GetAverageRatings(c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAverageRatings", reflect.TypeOf((*MockIReviewController)(nil).GetAverageRatings), c)
}

// GetLatestRatings mocks base method.
func (m *MockIReviewController) GetLatestRatings(c *v2.Ctx) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLatestRatings", c)
	ret0, _ := ret[0].(error)
	return ret0
}

// GetLatestRatings indicates an expected call of GetLatestRatings.
func (mr *MockIReviewControllerMockRecorder) GetLatestRatings(c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLatestRatings", reflect.TypeOf((*MockIReviewController)(nil).GetLatestRatings), c)
}

// GetRatings mocks base method.
func (m *MockIReviewController) GetRatings(c *v2.Ctx) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRatings", c)
	ret0, _ := ret[0].(error)
	return ret0
}

// GetRatings indicates an expected call of GetRatings.
func (mr *MockIReviewControllerMockRecorder) GetRatings(c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRatings", reflect.TypeOf((*MockIReviewController)(nil).GetRatings), c)
}

// PostRating mocks base method.
func (m *MockIReviewController) PostRating(c *v2.Ctx) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PostRating", c)
	ret0, _ := ret[0].(error)
	return ret0
}

// PostRating indicates an expected call of PostRating.
func (mr *MockIReviewControllerMockRecorder) PostRating(c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PostRating", reflect.TypeOf((*MockIReviewController)(nil).PostRating), c)
}
