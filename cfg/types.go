package cfg

type Config struct {
	DBConfig   `json:"database,omitempty" yaml:"database"`
	Fiber      FiberConfig `json:"fiber,omitempty" yaml:"fiber"`
	SigningKey string      `json:"sigining_key,omitempty" yaml:"sigining_key"`
	Secret     string      `json:"secret,omitempty" yaml:"secret"`
	DBPass     string      `json:"db_pass,omitempty" yaml:"db_pass"`
	// default to production for security reasons
	LibrateEnv string `json:"librate_env,omitempty" yaml:"librate_env,default:production"`
}

type DBConfig struct {
	Engine    string `yaml:"engine,default:postgres"`
	Host      string `yaml:"host,default:localhost"`
	Port      uint16 `yaml:"port,default:5432"`
	Database  string `yaml:"database,default:librate"`
	TestDB    string `yaml:"test_db,default:librate_test"`
	User      string `yaml:"user,default:postgres"`
	Password  string `yaml:"password,default:postgres"`
	SSL       string `yaml:"ssl,default:unknown"`
	PG_Config string `yaml:"pg_config,default:/usr/bin/pg_config"`
}

type FiberConfig struct {
	Host string `yaml:"host,default:localhost"`
	Port string `yaml:"port,default:3000"`
}
