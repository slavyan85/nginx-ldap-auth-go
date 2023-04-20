package main

import (
	"flag"
	"net/http"

	"github.com/vkryuchenko/nginx-ldap-auth-go/app/clients"
	"github.com/vkryuchenko/nginx-ldap-auth-go/app/handlers"
	"github.com/vkryuchenko/nginx-ldap-auth-go/config"
	"go.uber.org/zap"
)

func main() {
	configFile := "config.yml"
	flag.StringVar(&configFile, "config", configFile, "Path to configuration file")
	flag.Parse()
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	cfg, err := config.NewAppConfig(configFile)
	if err != nil {
		logger.Fatal(err.Error())
	}
	if cfg.Debug {
		logger, err = zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		logger.Debug("debug mode")
	}
	defer logger.Sync()
	ldapClient, err := clients.NewLdapClient(cfg.Ldap)
	if err != nil {
		logger.Fatal(err.Error())
	}
	routeHandler := handlers.HttpHandler{
		CookieName: cfg.CookieName,
		LdapClient: ldapClient,
		Logger:     logger,
	}

	http.HandleFunc(cfg.Url, routeHandler.AuthRoute)
	http.HandleFunc("/", routeHandler.DefaultRoute)
	logger.Fatal(http.ListenAndServe(cfg.Bind, nil).Error())
}
