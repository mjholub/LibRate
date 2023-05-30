package cfg

type Config struct {
	DBConfig    `json:"database,omitempty" yaml:"database"`
	Fiber       FiberConfig `json:"fiber,omitempty" yaml:"fiber"`
	SiginingKey string      `json:"sigining_key,omitempty" yaml:"sigining_key"`
	DBPass      string      `json:"db_pass,omitempty" yaml:"db_pass"`
}

type DBConfig struct {
	Engine   string `yaml:"engine,default:postgres"`
	Host     string `yaml:"host,default:localhost""`
	Port     uint8  `yaml:"port,default:5432""`
	Database string `yaml:"database,default:librerym"`
	User     string `yaml:"user,default:postgres"`
	Password string `yaml:"password,default:postgres"`
}

type FiberConfig struct {
	Host string `yaml:"host,default:localhost"`
	Port string `yaml:"port,default:3000"`
}

type DgraphConfig struct {
	Host           string `yaml:"host"`
	GRPCPort       string `yaml:"grpc_port"`
	HTTPPort       string `yaml:"http_port"`
	AlphaBadger    string `yaml:"alpha_badger"`
	AlphaBlockRate string `yaml:"alpha_block_rate"`
	AlphaTrace     string `yaml:"alpha_trace"`
	AlphaTLS       string `yaml:"alpha_tls"`
	AlphaSecurity  string `yaml:"alpha_security"`
}
