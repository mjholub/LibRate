package logging

type Config struct {
	Level     string          `yaml:"level"`
	Target    string          `yaml:"target"`
	Format    string          `yaml:"format"`
	Caller    bool            `yaml:"caller"`
	Timestamp TimestampConfig `yaml:"timestamp"`
}

type TimestampConfig struct {
	Enabled bool   `yaml:"enabled"`
	Format  string `yaml:"format"`
}
