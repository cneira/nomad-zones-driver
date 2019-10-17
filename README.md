Zones Task Driver
===========================

Task driver for [Illumos](https://illumos.org/) zones. 


- Website: https://www.nomadproject.io

Requirements
------------

- [Nomad](https://www.nomadproject.io/downloads.html) 0.9+
- [Go](https://golang.org/doc/install) 1.11 (to build the provider plugin)
- [OmniOS](https://omniosce.org/)
- [Consul](https://releases.hashicorp.com/consul/1.5.1/consul_1.5.1_solaris_amd64.zip)   

Examples 
---------

[Omnios Operations: Simple Zone](https://omniosce.org/setup/firstzone.html)

```hcl
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
            Physical       = "vnic0"
            AllowedAddress = "192.168.1.120/24"
            Defrouter      = "192.168.1.1"
          },
        ]
      }
    }
  }
}
```
 
[Omnios Operations: LX branded Zone](https://omniosce.org/info/lxzones.html)
  
```hcl
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
```
  
[Omnios Operations: BHYVE/KVM branded zone](https://omniosce.org/info/bhyve_kvm_brand.html)

```hcl
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
```
[Using a docker image from V2 registry ](https://hub.docker.com/r/beamdog/nwserver)
```hcl
job "docker-test" {
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
	Docker = "beamdog/nwserver latest"
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
```

Task driver zone config match [ZONECFG(1M)](https://illumos.org/man/1m/zonecfg) options, except for "img" custom attribute that is needed
for a lx branded zone, the "img" attribute should be a .zss.gz or tar.gz file.
Docker: Specify image from the docker registry v2 from which the zone will be created. 
Check information on settings in [ZONECFG(1M)](https://illumos.org/man/1m/zonecfg) man page.


USAGE:
--------

* Start [Consul](https://releases.hashicorp.com/consul/1.5.1/consul_1.5.1_solaris_amd64.zip)
```
cneira@Trixie:$  screen consul agent -dev -bind 0.0.0.0 -client 0.0.0.0  
```

* Now start the nomad agent 

```
cneira@Trixie:$ pfexec nomad agent -dev -config=config.hcl -data-dir=$GOPATH/src/github.com/hashicorp/nomad-zone-driver -plugin-dir=$GOPATH/src/github.com/hashicorp/nomad-zones-driver/plugin -bind=0.0.0.0 
```

* Finally submit a job 
``` 
cneira@Trixie:..m/hashicorp/nomad-zones-driver$ nomad run example-zone-task.nomad
```

* Check the status of the allocations of your job and grab the allocation Id.
```
cneira@Trixie:..m/hashicorp/nomad-zones-driver$ nomad job status example-zone-task

cneira@Trixie:…m/hashicorp/nomad-zones-driver$ nomad alloc status 676f5c1d
ID                  = 676f5c1d
Eval ID             = d1067dd1
Name                = test-zone-props-3.test[0]
Node ID             = c00ecba7
Node Name           = Trixie
Job ID              = test-zone-props-3
Job Version         = 2
Client Status       = running
Client Description  = Tasks are running
Desired Status      = run
Desired Description = <none>
Created             = 2m44s ago
Modified            = 2s ago

Task "test01" is "running"
Task Resources
CPU      Memory   Disk     Addresses
100 MHz  300 MiB  300 MiB

Task Events:
Started At     = 2019-05-27T15:17:56Z
Finished At    = N/A
Total Restarts = 0
Last Restart   = N/A

Recent Events:
Time                       Type        Description
2019-05-27T11:17:56-04:00  Started     Task started by client
2019-05-27T11:15:15-04:00  Task Setup  Building Task Directory
2019-05-27T11:15:15-04:00  Received    Task received by client
```

## Support

[![ko-fi](https://www.ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/J3J4YM9U)

It's also possible to support the project on [Patreon](https://www.patreon.com/neirac)

 TODO:
-------

* Implement exec interface
* Test all zone properties.
* Match naming convention between [ZONECFG(1M)](https://illumos.org/man/1m/zonecfg) and nomad zone driver. 
* Specify a registry option from where to pull images.
