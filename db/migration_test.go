package db

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"codeberg.org/mjh/LibRate/cfg"
)

type testCase struct {
	name    string
	inputs  interface{}
	wantErr bool
}

func TestMigrate(t *testing.T) {
	conf := &cfg.TestConfig
	tcs := []testCase{
		{
			name:    "all migrations",
			inputs:  nil,
			wantErr: false,
		},
		{
			name:    "single migration",
			inputs:  "000001-fix-missing-timestamps",
			wantErr: false,
		},
		{
			name:    "multiple paths",
			inputs:  []string{"000001-fix-missing-timestamps", "000002-reduce-uuid-usage"},
			wantErr: false,
		},
		{
			name:    "non-existent migration",
			inputs:  "obsaiwrbiweqb93928",
			wantErr: true,
		},
	}
	var err error
	for _, tc := range tcs {
		switch tc.inputs {
		case nil:
			err = Migrate(conf)
		default:
			if _, ok := tc.inputs.(string); ok {
				err = Migrate(conf, tc.inputs.(string))
			} else if _, ok := tc.inputs.([]string); ok {
				err = Migrate(conf, tc.inputs.([]string)...)
			}
		}
		if tc.wantErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestCountFiles(t *testing.T) {
	testDir := "000011-film-images"
	count, err := countFiles(testDir)
	assert.Equal(t, uint8(6), count)
	assert.NoError(t, err)
}

func TestGetDir(t *testing.T) {
	tcs := []testCase{
		{
			name:    "empty base path",
			inputs:  "",
			wantErr: false,
		},
		{
			name:    "non-existent base path",
			inputs:  "bajasehyfsayy2347",
			wantErr: true,
		},
		{
			name:    "an existing migration base path",
			inputs:  "000001-fix-missing-timestamps",
			wantErr: false,
		},
	}
	for _, tc := range tcs {
		if tc.wantErr {
			dirInfo, err := getDir(tc.inputs.(string))
			assert.NotNil(t, err)
			assert.Empty(t, dirInfo)
		} else {
			dirInfo, err := getDir(tc.inputs.(string))
			assert.Nil(t, err)
			assert.NotEmpty(t, dirInfo)
		}
	}
}
