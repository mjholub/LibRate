package testhelpers

import "testing"

type TestCase[T any, R any] struct {
	Name   string
	Input  func() T
	Output R
}

// TestCaseFunc is a test case with a function specified function for running it
type TestCaseFunc func(t *testing.T, input any, result any, args ...any)
