package config

type Config struct {
	Name   string `yaml:"name"`
	Port   int    `yaml:"port"`
	MySQL  struct {
		Host            string `yaml:"host"`
		Port            int    `yaml:"port"`
		User            string `yaml:"user"`
		Password        string `yaml:"password"`
		Database        string `yaml:"database"`
		MaxOpenConns    int    `yaml:"maxOpenConns"`
		MaxIdleConns    int    `yaml:"maxIdleConns"`
		ConnMaxLifetime int    `yaml:"connMaxLifetime"`
	} `yaml:"mysql"`
	JWT struct {
		Secret   string `yaml:"secret"`
		SignKey  string `yaml:"signKey"`
		Expiry   int    `yaml:"expiry"`
	} `yaml:"jwt"`
}
