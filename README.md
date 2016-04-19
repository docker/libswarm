# Swarm: Cluster orchestration for Docker

[![GoDoc](https://godoc.org/github.com/docker/swarm-v2?status.png)](https://godoc.org/github.com/docker/swarm-v2)
[![Circle CI](https://circleci.com/gh/docker/swarm-v2.svg?style=shield&circle-token=a7bf494e28963703a59de71cf19b73ad546058a7)](https://circleci.com/gh/docker/swarm-v2)
[![codecov.io](https://codecov.io/github/docker/swarm-v2/coverage.svg?branch=master&token=LqD1dzTjsN)](https://codecov.io/github/docker/swarm-v2?branch=master)

## Build

Requirements:

- go 1.6
- A [working golang](https://golang.org/doc/code.html) environment


From the project root directory run:

```sh
$ make binaries
```

## Install

```sh
$ sudo -E PATH=$PATH make install
```

This will install `/usr/local/bin/swarmd` (the manager and agent) and `/usr/local/bin/swarmctl` (the command line tool).

## Test

Before running tests for the first time, setup the tooling:

```bash
$ make setup
```

Then run:

```bash
$ make all
```

## Usage Examples

**1 manager + 2 agent cluster on a single host**

These instructions assume that `swarmd` and `swarmctl` are in your PATH.

Start the manager:

```sh
$ swarmd manager --log-level info --state-dir /tmp/manager-state
```

In two additional terminals, start two agents:

```sh
$ swarmd agent --log-level info --hostname node-1
$ swarmd agent --log-level info --hostname node-2
```

In a fourth terminal, use `swarmctl` to explore and control the cluster.  List nodes:

```
$ swarmctl node ls
ID                         Name      Status  Availability
87pn3pug404xs4x86b5nwlwbr  ubuntu15  READY   ACTIVE
by2ihzjyg9m674j3cjdit3reo  ubuntu15  READY   ACTIVE
```

**Create and manage a Service**

Note:  the term "Job" is being gradually replaced by "Service."

The `ping` job in `examples/job/ping.yml` is a place to start:

```
$ cd examples/job/
$ cat ping.yml
name: ping
image: alpine
command: ["sh", "-c", "ping $HOST"]
instances: 2
env:
 - HOST=google.com
```

Let's start it:

```
$ swarmctl job create -f ping.yml
chlkcf9v19kxbccspmiyuttgz
$ swarmctl job ls
ID                         Name  Image   Instances
chlkcf9v19kxbccspmiyuttgz  ping  alpine  2
$ swarmctl task ls
ID                         Job   Status   Node
1y72dcy9us5vvgsltz5dgm2pp  ping  RUNNING  ubuntu15
afhq97lrlw7jx1vh15gnofy59  ping  RUNNING  ubuntu15
```

Now change instance count in the YAML file:

```
$ vi ping.yml
[change instances to 3 and save]
```

Let's look at the delta:

```sh
$ swarmctl job diff ping -f ping.yml
--- remote
+++ local
@@ -6,5 +6,5 @@
 env:
 - HOST=google.com
 name: ping
-instances: 2
+instances: 3
```

Update the job with the modified manifest and see the result:

```sh
$ swarmctl job update ping -f ping.yml
chlkcf9v19kxbccspmiyuttgz
$ swarmctl job ls
ID                         Name  Image   Instances
chlkcf9v19kxbccspmiyuttgz  ping  alpine  3
```

You can also update instance count on the command line with `--instances`:

```sh
$ swarmctl job update ping --instances 4
chlkcf9v19kxbccspmiyuttgz
$ swarmctl job ls
ID                         Name  Image   Instances
chlkcf9v19kxbccspmiyuttgz  ping  alpine  4
$ swarmctl task ls
ID                         Job   Status   Node
1y72dcy9us5vvgsltz5dgm2pp  ping  RUNNING  ubuntu15
703xq3ou3mokfayl2pceu024v  ping  RUNNING  ubuntu15
afhq97lrlw7jx1vh15gnofy59  ping  RUNNING  ubuntu15
b8peuqixb5nd34733ug0njpxo  ping  RUNNING  ubuntu15
```

You can also live edit the state file on the manager:

```
$ EDITOR=nano swarmctl job edit ping
[change instances to 5, Ctrl+o to save, Ctrl+x to exit]
--- old
+++ new
@@ -6,5 +6,5 @@
 env:
 - HOST=google.com
 name: ping
-instances: 4
+instances: 5

Apply changes? [N/y] y
chlkcf9v19kxbccspmiyuttgz
```

Now check the result:

```sh
$ swarmctl job ls
ID                         Name  Image   Instances
chlkcf9v19kxbccspmiyuttgz  ping  alpine  5
$ swarmctl task ls
ID                         Job   Status   Node
1y72dcy9us5vvgsltz5dgm2pp  ping  RUNNING  ubuntu15
703xq3ou3mokfayl2pceu024v  ping  RUNNING  ubuntu15
afhq97lrlw7jx1vh15gnofy59  ping  RUNNING  ubuntu15
b8peuqixb5nd34733ug0njpxo  ping  RUNNING  ubuntu15
```


## Starting a cluster with Compose

You can use the included `docker-compose.yml` to start a cluster as a set of containers for testing purposes:

    $ docker-compose up -d
    [...build output...]
    Creating swarmv2_manager_1
    Creating swarmv2_agent_1

    $ docker-compose scale agent=4
    Creating and starting swarmv2_agent_2 ... done
    Creating and starting swarmv2_agent_3 ... done
    Creating and starting swarmv2_agent_4 ... done

    $ docker-compose ps
          Name                    Command              State              Ports
    ---------------------------------------------------------------------------------------
    swarmv2_agent_1     swarmd agent -m manager:4242   Up
    swarmv2_agent_2     swarmd agent -m manager:4242   Up
    swarmv2_agent_3     swarmd agent -m manager:4242   Up
    swarmv2_agent_4     swarmd agent -m manager:4242   Up
    swarmv2_manager_1   swarmd manager                 Up      192.168.64.24:4242->4242/tcp

    $ alias swarmctl='docker-compose run agent swarmctl -a manager:4242'

    $ swarmctl node ls
    ID                         Name    Status  Availability
    --                         ----    ------  ------------
    3y773r5as4ritj55ckqg12l7l  docker  READY   ACTIVE
    5w2i03cyyhv7hsd2roysnssyu  docker  READY   ACTIVE
    6j9lfh33mruhyrjq58bbv2fyh  docker  READY   ACTIVE
    81hneegzxp7o0ghs5851pxclk  docker  READY   ACTIVE

    $ swarmctl service create -f examples/service/ping.yml
    5an56xpuich7uxqf2y73mpbgq

    $ swarmctl task ls
    ID                         Service  Status   Node
    --                         -------  ------   ----
    7cwxbnaw7voxoepd8ezycuk03  ping     RUNNING  docker
    dmldn80vuosfqaju3dkap2x8p  ping     RUNNING  docker

    $ docker-compose down
    [containers and network are removed]
