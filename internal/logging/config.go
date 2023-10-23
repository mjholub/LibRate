package logging

type Config struct {
	Level     string          `yaml:"level" default:"info"`
	Target    string          `yaml:"target" default:"stdout"`
	Format    string          `yaml:"format" default:"json"`
	Caller    bool            `yaml:"caller" default:"true"`
	Timestamp TimestampConfig `yaml:"timestamp" mapstructure:"timestamp"`
}

type TimestampConfig struct {
	Enabled bool   `yaml:"enabled" default:"true"`
	Format  string `yaml:"format" default:"2006-01-0215:04:05.000Z07:00"`
}
