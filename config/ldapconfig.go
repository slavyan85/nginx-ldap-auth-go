package config

type LdapConfig struct {
	Address string `yaml:"address"`
	Base    string `yaml:"base"`
	Bind    struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"bind"`
	Filter struct {
		User  string `yaml:"user"`
		Group string `yaml:"group"`
	} `yaml:"filter"`
	Ssl struct {
		Use        bool   `yaml:"use"`
		SkipTls    bool   `yaml:"skipTls"`
		SkipVerify bool   `yaml:"skipVerify"`
		ServerName string `yaml:"serverName"`
	} `yaml:"ssl"`
}
