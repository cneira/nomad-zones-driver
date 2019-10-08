package main

import (
	log "github.com/hashicorp/go-hclog"
	"github.com/cneira/nomad-zones-driver/zones"
	"github.com/hashicorp/nomad/plugins"
)

func main() {
	// Serve the plugin
	plugins.Serve(factory)
}

func factory(log log.Logger) interface{} {
	return zone.NewZoneDriver(log)
}
