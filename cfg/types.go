package cfg

// unfortunately, camel case must be used, instead the yaml parser will not work

type Config struct {
	DBConfig `json:"database,omitempty" yaml:"database" mapstructure:"database"`
	Fiber    FiberConfig `json:"fiber,omitempty" yaml:"fiber" mapstructure:"fiber"`
	Secret   string      `json:"secret,omitempty" yaml:"secret" mapstructure:"secret"`
	// default to production for security reasons
	LibrateEnv string `json:"librateEnv,omitempty" yaml:"librateEnv" default:"production" mapstructure:"librate_env"`
}

type DBConfig struct {
	Engine   string `yaml:"engine" default:"postgres"`
	Host     string `yaml:"host" default:"localhost"`
	Port     uint16 `yaml:"port" default:"5432"`
	Database string `yaml:"database" default:"librate"`
	TestDB   string `yaml:"testDatabase" default:"librate_test"`
	User     string `yaml:"user" default:"postgres"`
	Password string `yaml:"password,omitempty" default:"postgres"`
	SSL      string `yaml:"SSL" default:"unknown"`
	PGConfig string `yaml:"pgConfig,omitempty" default:"/usr/bin/pg_config"`
	StartCmd string `yaml:"startCmd,omitempty" default:"sudo service postgresql start"`
}

// refer to https://docs.gofiber.io/api/fiber#config
type FiberConfig struct {
	Host           string `yaml:"host" default:"localhost"`
	Port           int    `yaml:"port" default:"3000"`
	Prefork        bool   `yaml:"prefork" default:"false"`
	ReduceMemUsage bool   `yaml:"reduceMemUsage" default:"false"`
	StaticDir      string `yaml:"staticDir" default:"./static"`
}
