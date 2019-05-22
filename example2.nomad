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
