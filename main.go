package main

import (
	"github.com/gorilla/mux"
	"github.com/lengdanran/sbl/proxy"
	"log"
	"net/http"
	"strconv"
)

const CONF_FILE = "./conf/config.yaml" // filename of configuration.

// readConf read the configuration from CONF_FILE
func readConf() *Config {
	conf, err := ReadConfig(CONF_FILE)
	if err != nil {
		log.Fatalf("read config error: %s", err)
		return nil
	}
	err = conf.Validation()
	if err != nil {
		log.Fatalf("verify config error: %s", err)
		return nil
	}
	conf.Print()
	return conf
}

func maxAllowedMiddleware(n uint) mux.MiddlewareFunc {
	sem := make(chan struct{}, n)
	acquire := func() { sem <- struct{}{} }
	release := func() { <-sem }

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			acquire()
			defer release()
			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	// 1. read configuration
	config := readConf()
	if config == nil {
		log.Printf("Read configuration from %s failed, exit....", CONF_FILE)
		return
	}
	// 2. make routers for locations
	router := mux.NewRouter()
	for _, l := range config.Location {
		httpProxy, err := proxy.NewHTTPProxy(l.ProxyPass, l.BalanceMode)
		if err != nil {
			log.Fatalf("create proxy error: %s", err)
		}
		// start health check
		if config.HealthCheck {
			httpProxy.HealthCheck(config.HealthCheckInterval)
		}
		router.Handle(l.Pattern, httpProxy)
	}
	if config.MaxAllowed > 0 {
		router.Use(maxAllowedMiddleware(config.MaxAllowed))
	}
	svr := http.Server{
		Addr:    ":" + strconv.Itoa(config.Port),
		Handler: router,
	}

	// 3. listen and serve
	log.Printf("Serve Schema = %s\n", config.Schema)
	if config.Schema == "http" {
		err := svr.ListenAndServe()
		if err != nil {
			log.Fatalf("listen and serve error: %s", err)
		}
	} else if config.Schema == "https" {
		err := svr.ListenAndServeTLS(config.SSLCertificate, config.SSLCertificateKey)
		if err != nil {
			log.Fatalf("listen and serve error: %s", err)
		}
	}

}
