package cfg

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type (
	TestWithEnv struct {
		suite.Suite
		ConfigFile string
	}

	HeuristicTest struct {
		suite.Suite
		ConfigLocations []string
	}
)

func (suite *TestWithEnv) SetupTest() {
	suite.ConfigFile = "example_config.yaml"
	os.Setenv("LIBRATE_CONFIG", suite.ConfigFile)
	loc, err := lookForExisting([]string{suite.ConfigFile})
	assert.NoErrorf(suite.T(), err, "error looking for existing config file: %s", err.Error())
	assert.NotEmptyf(suite.T(), loc, "configLocation is empty")
}

func (ht *HeuristicTest) SetupTest() {
	ht.ConfigLocations = tryLocations()
	assert.NotEmptyf(ht.T(), ht.ConfigLocations, "configLocations is empty")
	configLocation, err := lookForExisting(ht.ConfigLocations)
	assert.NotEmptyf(ht.T(), configLocation, "configLocation is empty")
	assert.NoErrorf(ht.T(), err, "error looking for existing config file: %s", err.Error())
}

func (suite *TestWithEnv) TearDownTest() {
	os.Unsetenv("LIBRATE_CONFIG")
}

func TestLookForExisting(t *testing.T) {
	suites := []suite.TestingSuite{
		&TestWithEnv{},
		&HeuristicTest{},
	}
	for s := range suites {
		suite.Run(t, suites[s])
	}
}
