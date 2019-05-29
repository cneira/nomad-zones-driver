Nomad Illumos Zones Driver
===========================
Task driver for managing Illumos zones


- Website: https://www.nomadproject.io

Requirements
------------

- [Nomad](https://www.nomadproject.io/downloads.html) 0.9+
- [Go](https://golang.org/doc/install) 1.11 (to build the provider plugin)
- Illumos Omnios Host with pkgsrc and sparse zone brand packages installed.
- [Consul](https://releases.hashicorp.com/consul/1.5.1/consul_1.5.1_solaris_amd64.zip)   

Examples 
---------

* sparse zone  

```
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
        Memory    = "2G"
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
* LX zone 
```
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
        Memory    = "2G"
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
* Zonepath : a valid dataset where zones will be created.
* Autoboot : the zone will be restarted at boot
* Brand :  zone type at this moment only  sparse, pkgsrc and lipkg work.
* Networks: Configure network for zone, if omitted no nic will be associated to the zone.
* Memory : Maximum memory that the zone is allowed to use (in GB).
* Lwps :   Maximum amount of lwps allowed.
* Attributes : custom attributes that will be added to the zone.
* LX Attributes: 
- img : path to the zss file that will be used to create the zone. 
- kernel-version : will be used by programs that check kernel version.  

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

 TODO:
-------

* Add brands bhyve and kvm.
* Test all zone properties.
