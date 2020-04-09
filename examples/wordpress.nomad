job "blog" {
  datacenters = ["dc1"]
  type        = "service"

  group "test" {
    restart {
      attempts = 0
      mode     = "fail"
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

        FileSystems = [
          {
            Dir     = "/var/www/html/wp-content"
            Special = "/home/cneira/docker/volumes/wpress/wp-content"
            Type    = "lofs"
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
