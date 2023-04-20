package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Bind       string     `yaml:"bind"`
	CookieName string     `yaml:"cookieName"`
	Debug      bool       `yaml:"debug"`
	Url        string     `yaml:"url"`
	Ldap       LdapConfig `yaml:"ldap"`
}

func NewAppConfig(config string) (*AppConfig, error) {
	cf, err := os.Open(config)
	if err != nil {
		return nil, err
	}
	defer cf.Close()
	ac := &AppConfig{}
	err = yaml.NewDecoder(cf).Decode(ac)
	if err != nil {
		return nil, err
	}
	return ac, nil
}
