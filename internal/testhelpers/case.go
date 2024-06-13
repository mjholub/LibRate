package testhelpers

type TestCase[T any, R any] struct {
	Name   string
	Input  func() T
	Output R
}
