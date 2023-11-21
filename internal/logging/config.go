package logging

type Config struct {
	Level     string          `yaml:"level" default:"info" validate:"oneof=trace debug info error fatal panic"`
	Target    string          `yaml:"target" default:"stdout" validate:"required,oneof=stdout stderr"`
	Format    string          `yaml:"format" default:"console" validate:"oneof=json console"`
	Caller    bool            `yaml:"caller" default:"true"`
	Timestamp TimestampConfig `yaml:"timestamp" mapstructure:"timestamp"`
}

type TimestampConfig struct {
	Enabled bool   `yaml:"enabled" default:"true"`
	Format  string `yaml:"format" default:"2006-01-0215:04:05.000Z07:00"`
}
