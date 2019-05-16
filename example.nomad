job "test-zone" {
  datacenters = ["dc1"]
  type        = "service"

  update {
    canary            = 1
    max_parallel      = 1
    healthy_deadline  = "8m"
    progress_deadline = "10m"
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
