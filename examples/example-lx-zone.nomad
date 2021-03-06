job "lx-test" {
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
        Brand     = "lx"
        CpuShares = "8000"
	CappedMemory = "4G"
	LockedMemory = "2G"
	SwapMemory = "4G"
        Lwps      = "3000"

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
          {
            Name  = "img"
            Type  = "string"
            Value = "/zcage/images/19aa3328-0025-11e7-a19a-c39077bfd4cf.zss.gz"
          },
          {
            Name  = "kernel-version"
            Type  = "string"
            Value = "3.16.0"
          },
       ]

        Networks = [
          {
            Physical       = "vnic0"
            AllowedAddress = "192.168.1.120/24"
            Defrouter      = "192.168.1.1"
          },
        ]
      }
    }
  }
}
