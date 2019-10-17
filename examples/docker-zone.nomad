job "docker-lx-test6" {
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
        Zonepath  = "/vms"
        Autoboot  = false
        Brand     = "lx"
	Envars    = "MYSQL_ROOT_PASSWORD=somepassword  MYSQL_DATABASE=yourdb"
	Docker = "mysql 5.7"
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
            Name  = "kernel-version"
            Type  = "string"
            Value = "3.16.0"
          },
       ]
 FileSystems = [
          {
            Dir     = "/var/lib/mysql"
            Special = "/home/cneira/docker/volumes/mysql"
            Type    = "lofs"
         },
        ]

        Networks = [
          {
            Physical       = "vnic2"
            AllowedAddress = "192.168.1.120/24"
            Defrouter      = "192.168.1.1"
          },
        ]
      }
    }
  }
}
