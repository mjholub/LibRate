package cfg

type Config struct {
	ArangoDB    ArangoDBConfig
	Dgraph      DgraphConfig
	Fiber       FiberConfig
	SiginingKey string
}

type ArangoDBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

type FiberConfig struct {
	Host string
	Port string
}

type DgraphConfig struct {
	Host           string
	GRPCPort       string
	HTTPPort       string
	AlphaBadger    string
	AlphaBlockRate string
	AlphaTrace     string
	AlphaTLS       string
	AlphaSecurity  string
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}
