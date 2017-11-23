package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/aybabtme/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	DefaultDomain     = "localhost"
	DefaultListenAddr = "0.0.0.0:8080"
)

func main() {
	app := kingpin.New(
		"subdomainr",
		"test experiment around subdomains",
	).DefaultEnvars()
	serveCmd(app.Command("serve", "serve up the server"))
	_, err := app.Parse(os.Args[1:])
	if err != nil {
		log.Err(err).Fatal("failed to run")
	}
}

func serveCmd(cmd *kingpin.CmdClause) {
	var cfg = struct {
		Domain     string
		ListenAddr string
	}{}

	cmd.Flag("listen.addr", "address where to listen.").Default(DefaultListenAddr).StringVar(&cfg.ListenAddr)
	cmd.Flag("domain.url", "domain of the application").Default(DefaultDomain).StringVar(&cfg.Domain)

	cmd.Action(func(*kingpin.ParseContext) error {
		wildcardDomain := "{subdomain:[\\w-]+}." + cfg.Domain
		r := mux.NewRouter()
		log.KV("domain", wildcardDomain).Info("debugging wildcard domain")
		s := r.Host(wildcardDomain).Subrouter()
		s.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello from the current host: %s", r.Host)
		})

		r.Host(cfg.Domain).PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello from the base host: %s", r.Host)
		})

		l, err := net.Listen("tcp", cfg.ListenAddr)
		if err != nil {
			return errors.Wrap(err, "creating listener")
		}

		actualAddr := l.Addr().String()

		srv := &http.Server{
			Addr:         actualAddr,
			Handler:      r,
			WriteTimeout: 60 * time.Second,
			ReadTimeout:  60 * time.Second,
		}

		ll := log.KV("listen.addr", actualAddr)
		ll.Info("listening")

		return srv.Serve(l)
	})
}
