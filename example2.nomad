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
		Brand = "sparse",
		CpuShares = 8000,
		Memory = 4000,
		Networks = [
			 {
			    Address = "192.168.1.120/24",
			    Physical = "vnic0",
			    Defrouter = "192.168.1.1"
	             }
	    ]
	}
    }
 }
}
