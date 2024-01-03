package crypt

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	input, name string
	expectFail  bool
	mockFunc    func() error
}

func TestAbsPathCheckExistence(t *testing.T) {
	// in my case, first two paths exist, so there should be no error
	testPaths := []string{
		"/home/user/.ssh/id_rsa.pub",
		"~/.ssh/id_rsa.pub",
		"id_rsa_mock.pub",
		"../id_rsa_mock.pub",
		"id_rsa.pub",
		"../id_rsa.pub",
	}
	// NOTE: as a pre
	testCases := []testCase{
		{
			input:      testPaths[0],
			name:       "standard absolute path",
			expectFail: false,
		},
		{
			input:      testPaths[1],
			name:       "standard relative path",
			expectFail: false,
		},
		{
			input:      testPaths[4],
			name:       "relative path 0, no mock file",
			expectFail: true,
		},
		{
			input:      testPaths[5],
			name:       "relative path 1, no mock file",
			expectFail: true,
		},
		{
			input:      testPaths[2],
			name:       "relative mock path 0. file exists",
			expectFail: false,
			mockFunc: func() error {
				err := createMockFile(testPaths[2])
				if err != nil {
					t.Fatalf("could not create mock file: %v", err)
				}
				return nil
			},
		},
		{
			input:      testPaths[3],
			name:       "relative mock path 1, file exists",
			expectFail: false,
			mockFunc: func() error {
				err := createMockFile(testPaths[3])
				if err != nil {
					t.Fatalf("could not create mock file: %v", err)
				}
				return nil
			},
		},
		{
			input:      testPaths[4],
			name:       "relative path 2, no mock file",
			expectFail: true,
		},
		{
			input:      testPaths[5],
			name:       "relative path 3, no mock file",
			expectFail: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mockFunc != nil {
				err := tc.mockFunc()
				if err != nil {
					t.Fatalf("could not execute mock function: %v", err)
				}
			}
			_, err := getAbsPathCheckExistence(tc.input)
			if tc.expectFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func createMockFile(path string) error {
	_, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create mock file: %w", err)
	}
	return nil
}
