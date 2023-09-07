package cfg

type Config struct {
	DBConfig `json:"database,omitempty" yaml:"database" mapstructure:"database"`
	Fiber    FiberConfig `json:"fiber,omitempty" yaml:"fiber" mapstructure:"fiber"`
	Secret   string      `json:"secret,omitempty" yaml:"secret" mapstructure:"secret"`
	// default to production for security reasons
	LibrateEnv string `json:"librate_env,omitempty" yaml:"librate_env" default:"production" mapstructure:"librate_env"`
}

type DBConfig struct {
	Engine    string `yaml:"engine" default:"postgres"`
	Host      string `yaml:"host" default:"localhost"`
	Port      uint16 `yaml:"port" default:"5432"`
	Database  string `yaml:"database" default:"librate"`
	TestDB    string `yaml:"test_database" default:"librate_test"`
	User      string `yaml:"user" default:"postgres"`
	Password  string `yaml:"password,omitempty" default:"postgres"`
	SSL       string `yaml:"SSL" default:"unknown"`
	PG_Config string `yaml:"pg_config,omitempty" default:"/usr/bin/pg_config"`
}

type FiberConfig struct {
	Host string `yaml:"host" default:"localhost"`
	Port int    `yaml:"port" default:"3000"`
}
