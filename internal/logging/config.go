package logging

type Config struct {
	Level  string `yaml:"level"`
	Target string `yaml:"target"`
	Format string `yaml:"format"`
}
