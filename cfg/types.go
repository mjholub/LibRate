package cfg

// unfortunately, camel case must be used, instead the yaml parser will not work

type Config struct {
	DBConfig `json:"database,omitempty" yaml:"database" mapstructure:"database"`
	Fiber    FiberConfig `json:"fiber,omitempty" yaml:"fiber" mapstructure:"fiber"`
	Secret   string      `json:"secret,omitempty" yaml:"secret" mapstructure:"secret"`
	// default to production for security reasons
	LibrateEnv string      `json:"librateEnv,omitempty" yaml:"librateEnv" default:"production" mapstructure:"librate_env"`
	Redis      RedisConfig `json:"redis,omitempty" yaml:"redis" mapstructure:"redis"`
}

type DBConfig struct {
	Engine             string `yaml:"engine" default:"postgres"`
	Host               string `yaml:"host" default:"localhost"`
	Port               uint16 `yaml:"port" default:"5432"`
	Database           string `yaml:"database" default:"librate"`
	User               string `yaml:"user" default:"postgres"`
	Password           string `yaml:"password,omitempty" default:"postgres"`
	SSL                string `yaml:"SSL" default:"unknown"`
	PGConfig           string `yaml:"pgConfig,omitempty" default:"/usr/bin/pg_config"`
	StartCmd           string `yaml:"startCmd,omitempty" default:"sudo service postgresql start"`
	AutoMigrate        bool   `yaml:"autoMigrate,omitempty" default:"true"`
	ExitAfterMigration bool   `yaml:"exitAfterMigration,omitempty" default:"false"`
}

type RedisConfig struct {
	Host     string `yaml:"host,omitempty" default:"localhost"`
	Port     int    `yaml:"port,omitempty" default:"6379"`
	Username string `yaml:"username,omitempty" default:""`
	Password string `yaml:"password,omitempty" default:""`
	Database int    `yaml:"database,omitempty" default:"0"`
}

// refer to https://docs.gofiber.io/api/fiber#config
type FiberConfig struct {
	Host           string `yaml:"host" default:"localhost"`
	Port           int    `yaml:"port" default:"3000"`
	Prefork        bool   `yaml:"prefork" default:"false"`
	ReduceMemUsage bool   `yaml:"reduceMemUsage" default:"false"`
	StaticDir      string `yaml:"staticDir" default:"./static"`
	PowInterval    int    `yaml:"powInterval" default:"300"`
	PowDifficulty  int    `yaml:"powDifficulty" default:"30000"`
}
