package config

type Config struct {
	Default DefaultConfig
	MYSQL   MysqlConfig
}

type DefaultConfig struct {
	Listen int
}

type MysqlConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}
