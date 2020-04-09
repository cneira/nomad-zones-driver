job "wpblog" {
  datacenters = ["dc1"]
  type        = "service"

  group "test" {
    restart {
      attempts = 0
      mode     = "fail"
    }

    task "mysql" {
      driver = "zone"

      config {
        Zonepath     = "/zcage/vms"
        Autoboot     = false
        Brand        = "lx"
        Envars       = "MYSQL_ROOT_PASSWORD=yourpass  MYSQL_DATABASE=yourdb"
        Docker       = "mysql 5.7"
        CpuShares    = "8000"
        CappedMemory = "4G"
        LockedMemory = "2G"
        SwapMemory   = "4G"
        Lwps         = "3000"

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

         Networks = [
          {
            Physical       = "vnic2"
            AllowedAddress = "192.168.1.120/24"
            Defrouter      = "192.168.1.1"
          },
        ]
      }
    }

    task "wordpress" {
      driver = "zone"

      config {
        Zonepath     = "/zcage/vms"
        Autoboot     = false
        Brand        = "lx"
        Envars       = "WORDPRESS_DB_PASSWORD=yourdb"
        Docker       = "wordpress php7.3-fpm-alpine"
        CpuShares    = "8000"
        CappedMemory = "4G"
        LockedMemory = "2G"
        SwapMemory   = "4G"
        Lwps         = "3000"

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

         Networks = [
          {
            Physical       = "vnic1"
            AllowedAddress = "192.168.1.121/24"
            Defrouter      = "192.168.1.1"
          },
        ]
      }
    }
  }
}
