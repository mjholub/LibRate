package logging

type Config struct {
	Level     string          `yaml:"level" default:"info" validate:"oneof=trace debug info error fatal panic" env:"LIBRATE_LOG_LEVEL"`
	Target    string          `yaml:"target" default:"stdout" validate:"required,oneof=stdout stderr" env:"LIBRATE_LOG_TARGET"`
	Format    string          `yaml:"format" default:"console" validate:"oneof=json console" env:"LIBRATE_LOG_FMT"`
	Caller    bool            `yaml:"caller" default:"true" env:"LIBRATE_LOG_CALLER"`
	Timestamp TimestampConfig `yaml:"timestamp" mapstructure:"timestamp"`
}

type TimestampConfig struct {
	Enabled bool   `yaml:"enabled" default:"true" env:"LIBRATE_LOG_TS_ENABLED"`
	Format  string `yaml:"format" default:"2006-01-0215:04:05.000Z07:00" env:"LIBRATE_LOG_TS_FORMAT"`
}
