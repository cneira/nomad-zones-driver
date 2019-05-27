job "test-zone-props" {
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
		Brand = "sparse",
		CpuShares = "8000",
		Memory = "2G",
		Networks = [
			 {
			    Physical = "vnic0",
			    AllowedAddress = "192.168.1.120/24",
			    Defrouter = "192.168.1.1"
	             }
	    ]
	}
    }
 }
}
