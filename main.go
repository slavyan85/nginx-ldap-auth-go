package main

import (
	"flag"
	"go.uber.org/zap"
	"net/http"
	"nginx-ldap-auth-go/appconfig"
	"nginx-ldap-auth-go/httphandler"
	"nginx-ldap-auth-go/ldaphandler"
)

func main() {
	configFile := "config.yml"
	flag.StringVar(&configFile, "config", configFile, "Path to configuration file")
	flag.Parse()
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	appConfig := appconfig.AppConfig{}
	err = appConfig.Apply(configFile)
	if err != nil {
		logger.Fatal(err.Error())
	}
	if appConfig.Debug {
		logger, err = zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		logger.Debug("debug mode")
	}
	defer logger.Sync()
	ldapClient, err := ldaphandler.NewClient(appConfig.Ldap)
	if err != nil {
		logger.Fatal(err.Error())
	}
	routeHandler := httphandler.Handler{
		CookieName: appConfig.CookieName,
		LdapClient: ldapClient,
		Logger:     logger,
	}

	http.HandleFunc(appConfig.Url, routeHandler.AuthRoute)
	http.HandleFunc("/", routeHandler.DefaultRoute)
	logger.Fatal(http.ListenAndServe(appConfig.Bind, nil).Error())
}
