job "test-nomad-zone-driver" {
  datacenters = ["dc1"]
  type        = "service"

  group "test" {
    restart {
      attempts = 0
      mode     = "fail"
    }

    task "test01" {
      driver = "zone"

      config {
        Zonepath  = "/zcage/vms"
        Autoboot  = false
        Brand     = "sparse"
        CpuShares = "8000"
	CappedMemory = "4G"
	LockedMemory = "2G"
	SwapMemory = "4G"
	DedicatedCpu = "1"
        Lwps      = "3000"
	IpType = "exclusive"
	
        Attributes = [
          {
            Name  = "resolvers"
            Type  = "string"
            Value = "8.8.8.8"
          },
          {
            Name  = "resolvers"
            Type  = "string"
            Value = "8.8.8.4"
          },
       ]

        Networks = [
          {
            Physical       = "net_0"
	    GlobalNic = "auto"
            AllowedAddress = "192.168.1.120/24"
            Defrouter      = "192.168.1.1"
          },
        ]
      }
    }
  }
}
