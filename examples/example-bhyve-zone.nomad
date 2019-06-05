job "bhyve-test" {
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
        Brand     = "bhyve"
        CpuShares = "8000"
        Lwps      = "3000"
	CappedMemory = "4G"
	LockedMemory = "2G"
	SwapMemory = "4G"

        Attributes = [
          {
            Name  = "bootdisk"
            Type  = "string"
            Value = "rpool/b0"
          },
          {
            Name  = "cdrom"
            Type  = "string"
            Value = "/home/cneira/test.iso"
          },
        ]

        FileSystems = [
          {
            Dir     = "/home/cneira/test.iso"
            Special = "/home/cneira/test.iso"
            Type    = "lofs"

            Fsoption = [
              {
                Name = "ro"
              },
              {
                Name = "nodevices"
              },
            ]
          },
        ]

        Devices = [
          {
            Match = "/dev/zvol/rdsk/rpool/b0"
          },
        ]

        Networks = [
          {
            Physical       = "vnic5"
            AllowedAddress = "192.168.1.254/24"
            Defrouter      = "192.168.1.1"
          },
        ]
      }
    }
  }
}
