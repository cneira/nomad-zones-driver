Nomad Illumos Zones Driver
===========================

At this point this is WIP, only sparse and pkgsrc zone work without resource control nor network access.

- Website: https://www.nomadproject.io

Requirements
------------

- [Nomad](https://www.nomadproject.io/downloads.html) 0.9+
- [Go](https://golang.org/doc/install) 1.11 (to build the provider plugin)
- Illumos Omnios Host with pkgsrc zone brand packages installed.

Job spec
---------

job "test-zone" {
  datacenters = ["dc1"]
  type        = "service"
  group "test" {
  restart {
  	  attempts = 0
	  mode = "fail"
	}
  task "test01" {
      driver = "zone"
	config {
		Zonepath = "/zcage/vms",
		Autoboot = false,
		Brand = "sparse"
	}
    }
 }
}

## Zonepath : a valid dataset where zones will be created.
## Autoboot : the zone will be restarted at boot
## Brand :  zone type at this moment only  sparse, pkgsrc and lipkg work.

 TODO:
-------

* Add Network, cpu-shares and memory resource control.