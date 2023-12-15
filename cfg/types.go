package cfg

import (
	"codeberg.org/mjh/LibRate/internal/logging"
)

// Config is the struct that holds all the configuration for the application
// unfortunately, camel case must be used, instead the yaml parser will not work
type Config struct {
	DBConfig `json:"database,omitempty" yaml:"database" mapstructure:"database"`
	Fiber    FiberConfig `json:"fiber,omitempty" yaml:"fiber" mapstructure:"fiber"`
	// used to encrypt sessions database
	Secret string `json:"secret,omitempty" yaml:"secret" mapstructure:"secret" env:"LIBRATE_SECRET"`
	// default to production for security reasons
	LibrateEnv string         `json:"librateEnv,omitempty" yaml:"librateEnv" default:"production" mapstructure:"librateEnv" env:"LIBRATE_ENV"`
	Redis      RedisConfig    `json:"redis,omitempty" yaml:"redis" mapstructure:"redis"`
	Logging    logging.Config `json:"logging,omitempty" yaml:"logging" mapstructure:"logging"`
	Keys       KeysConfig     `json:"keys,omitempty" yaml:"keys" mapstructure:"keys"`
	JWTSecret  string         `json:"jwtSecret,omitempty" yaml:"jwtSecret" mapstructure:"jwtSecret" env:"LIBRATE_JWT_SECRET"`
	GRPC       GrpcConfig     `json:"grpc,omitempty" yaml:"grpc" mapstructure:"grpc"`
}

// nolint: musttag // tagged in the struct above
type DBConfig struct {
	Engine             string `yaml:"engine" default:"postgres" env:"LIBRATE_DB_ENGINE"`
	Host               string `yaml:"host" default:"localhost" env:"LIBRATE_DB_HOST"`
	Port               uint16 `yaml:"port" default:"5432" env:"LIBRATE_DB_PORT"`
	Database           string `yaml:"database" default:"librate" env:"LIBRATE_DB_NAME"`
	User               string `yaml:"user" default:"postgres" env:"LIBRATE_DB_USER"`
	Password           string `yaml:"password,omitempty" default:"postgres" env:"LIBRATE_DB_PASSWORD"`
	SSL                string `yaml:"SSL" default:"unknown" env:"LIBRATE_DB_SSL"`
	PGConfig           string `yaml:"pgConfig,omitempty" default:"/usr/bin/pg_config" env:"PG_CONFIG_PATH"`
	AutoMigrate        bool   `yaml:"autoMigrate,omitempty" default:"true" env:"LIBRATE_AUTO_MIGRATE"`
	ExitAfterMigration bool   `yaml:"exitAfterMigration,omitempty" default:"false" env:"LIBRATE_EXIT_AFTER_MIGRATION"`
}

type RedisConfig struct {
	Host     string `yaml:"host,omitempty" default:"localhost" env:"LIBRATE_REDIS_HOST"`
	Port     int    `yaml:"port,omitempty" default:"6379" env:"LIBRATE_REDIS_PORT"`
	Username string `yaml:"username,omitempty" default:"" env:"LIBRATE_REDIS_USERNAME"`
	Password string `yaml:"password,omitempty" default:"" env:"LIBRATE_REDIS_PASSWORD"`
	CacheDB  int    `yaml:"cacheDb,omitempty" default:"0" env:"LIBRATE_CACHE_DB"`
	CsrfDB   int    `yaml:"csrfDb,omitempty" default:"2" env:"LIBRATE_CSRF_DB"`
	PowDB    int    `yaml:"powDb,omitempty" default:"3" env:"LIBRATE_POW_DB"`
}

// refer to https://docs.gofiber.io/api/fiber#config
type FiberConfig struct {
	Host           string `yaml:"host" default:"localhost" env:"LIBRATE_HOST"`
	Domain         string `yaml:"domain" default:"lr.localhost" env:"DOMAIN"`
	Port           int    `yaml:"port" default:"3000" env:"LIBRATE_PORT"`
	Prefork        bool   `yaml:"prefork" default:"false" env:"LIBRATE_PREFORK"`
	ReduceMemUsage bool   `yaml:"reduceMemUsage" default:"false" env:"LIBRATE_REDUCE_MEM"`
	StaticDir      string `yaml:"staticDir" default:"./static" env:"LIBRATE_ASSETS"`
	PowInterval    int    `yaml:"powInterval" default:"300" env:"POW_INTERVAL"`
	PowDifficulty  int    `yaml:"powDifficulty" default:"30000" env:"POW_DIFFICULTY"`
	RequestTimeout int    `yaml:"requestTimeout" default:"10" env:"LIBRATE_REQUEST_TIMEOUT"`
	TLS            bool   `yaml:"tls" default:"false" env:"LIBRATE_TLS"`
	MaxUploadSize  int64  `yaml:"maxUploadSize" default:"4194304" env:"LIBRATE_MAX_UPLOAD_SIZE"`
}

// KeysConfig defines the location of keys used for TLS
type KeysConfig struct {
	Private string `yaml:"private" default:"./keys/private.pem" env:"LIBRATE_PRIVATE_KEY"`
	Public  string `yaml:"public" default:"./keys/public.pem" env:"LIBRATE_PUBLIC_KEY"`
}

type GrpcConfig struct {
	Host            string `yaml:"host" default:"localhost" env:"LIBRATE_GRPC_HOST"`
	Port            int    `yaml:"port" default:"3030" env:"LIBRATE_GRPC_PORT"`
	ShutdownTimeout int    `yaml:"shutdownTimeout" default:"10" env:"LIBRATE_SHUTDOWN_TIMEOUT"`
}
