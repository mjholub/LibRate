// this file contaisn an exported global file with defalut config as a fallback if loading the
// proper config fails.

package cfg

var DefaultConfig = Config{
	DBConfig: DBConfig{
		Engine:   "postgres",
		Host:     "localhost",
		Port:     uint16(5432),
		Database: "librate",
		TestDB:   "librate_test",
		User:     "postgres",
		Password: "postgres",
		SSL:      "unknown",
		PGConfig: "/usr/bin/pg_config",
		StartCmd: "sudo service postgresql start",
	},
	Fiber: FiberConfig{
		Host:    "localhost",
		Port:    3000,
		Prefork: false,
	},
	Secret:     "secret",
	LibrateEnv: "production",
}
