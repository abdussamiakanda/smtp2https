package main

import "flag"

var (
	flagServerName     = flag.String("name", "smtp2https", "the server name")
	flagListenAddr     = flag.String("listen", ":smtp", "the smtp address to listen on")
	flagRoutesConfig   = flag.String("config", "", "path to JSON file mapping domains to webhook URLs")
	flagRoutesCLI      = make(routeFlags)
	flagMaxMessageSize = flag.Int64("msglimit", 1024*1024*25, "maximum incoming message size")
	flagReadTimeout    = flag.Int("timeout.read", 5, "the read timeout in seconds")
	flagWriteTimeout   = flag.Int("timeout.write", 5, "the write timeout in seconds")
	flagAuthUSER       = flag.String("user", "", "user for smtp client")
	flagAuthPASS       = flag.String("pass", "", "pass for smtp client")
)

func init() {
	flag.Var(&flagRoutesCLI, "route", "domain=webhookURL route (repeatable); overrides -config for the same domain")
	flag.Parse()
}
