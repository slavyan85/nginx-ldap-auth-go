package appconfig

import (
	"gopkg.in/yaml.v2"
	"nginx-ldap-auth-go/ldaphandler"
	"os"
)

type AppConfig struct {
	Bind       string                 `yaml:"bind"`
	CookieName string                 `yaml:"cookieName"`
	Debug      bool                   `yaml:"debug"`
	Url        string                 `yaml:"url"`
	Ldap       ldaphandler.LdapConfig `yaml:"ldap"`
}

func (ac *AppConfig) Apply(config string) error {
	cf, err := os.Open(config)
	if err != nil {
		return err
	}
	defer cf.Close()
	return yaml.NewDecoder(cf).Decode(ac)
}
