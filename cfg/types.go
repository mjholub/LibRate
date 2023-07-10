package cfg

type Config struct {
	DBConfig    `json:"database,omitempty" yaml:"database"`
	Fiber       FiberConfig `json:"fiber,omitempty" yaml:"fiber"`
	SiginingKey string      `json:"sigining_key,omitempty" yaml:"sigining_key"`
	DBPass      string      `json:"db_pass,omitempty" yaml:"db_pass"`
	// default to production for security reasons
	LibrateEnv string `json:"librate_env,omitempty" yaml:"librate_env,default:production"`
}

type DBConfig struct {
	Engine    string `yaml:"engine,default:postgres"`
	Host      string `yaml:"host,default:localhost"`
	Port      uint16 `yaml:"port,default:5432"`
	Database  string `yaml:"database,default:librerym"`
	User      string `yaml:"user,default:postgres"`
	Password  string `yaml:"password,default:postgres"`
	SSL       string `yaml:"ssl,default:unknown"`
	PG_Config string `yaml:"pg_config,default:/usr/bin/pg_config"`
}

type FiberConfig struct {
	Host string `yaml:"host,default:localhost"`
	Port string `yaml:"port,default:3000"`
}
