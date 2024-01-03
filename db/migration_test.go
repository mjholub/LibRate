package db

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/zerolog"
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
	log := zerolog.Nop()
	for _, tc := range tcs {
		switch tc.inputs {
		case nil:
			err = Migrate(&log, conf)
		default:
			if _, ok := tc.inputs.(string); ok {
				err = Migrate(&log, conf, tc.inputs.(string))
			} else if _, ok := tc.inputs.([]string); ok {
				err = Migrate(&log, conf, tc.inputs.([]string)...)
			}
		}
		if tc.wantErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
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

	dirInfo, err := getDir("", "./migrations")
	assert.Nil(t, err)
	fmt.Printf("%+v\n", dirInfo)

	// test parsing of directory used in Migrate()
	for dir, files := range dirInfo {
		dirPath := dir.Name()
		for i := range files {
			filePath := files[i].Name()
			joined := filepath.Join("./migrations/", dirPath, filePath)
			assert.Equal(t, len(strings.Split(joined, "/")), 3)
		}
	}

	for _, tc := range tcs {
		if tc.wantErr {
			dirInfo, err := getDir(tc.inputs.(string), "./migrations")
			assert.NotNil(t, err)
			assert.Empty(t, dirInfo)
		} else {
			dirInfo, err := getDir(tc.inputs.(string), "./migrations")
			assert.Nil(t, err)
			assert.NotZero(t, dirInfo)
		}
	}
}
