# Changelog

## master

- Fix docker user defined network.
- Add redis reconnect after unhealthy ping. 

## v0.3.0

- Add Redis Authentication Support
- Bootstrap: Add update call to update the bootstrap server.
- Add update api call to update the bootstrap server.
- Fix hostname constraint method. The server/agent should never run
  on a other worker as defined in the contrains variable.
- Fix mesos-cli to determine framework uuid by name.
- Move statefile to Redis DB
- Add TLS Server Support (env variable SSL_CRT_BASE64, SSL_KEY_BASE64)
- Change DB items framework and framework_config to be saved with the
  frameworkName as prefix.
- Fix M3SStatus and scalet
- Optimize framework suppress
- Check if port is already in use
- Add version as flag to bootstrap server and m3s
- IMPORTANT!!! Change all API Calls URL's.
- Add Docker Network (`docker network create`) support. The configuration variable ist "DOCKER_CNI".
- Seperate K3S Server and Agent Memory and CPU resource definition.
- Change K3S to use docker engine
- Add DOCKER_SHM_SIZE variable to configure shm-size.
- Increase etcd election timeout

## v0.2.0

- Add status information about the M3s services.
- Mesos-CLI: Get out the M3s status information.
- Add Mesos Label support to (as example) control traefik
- Add possibility to run the K8 Server and Agent on a specified hostname
- Fix status information of all m3s services
- Mesos-CLI: Resolve framework ID also by name
- Add configurable FrameworkHostname
- Add FrameworkName to Port Label
- Add Kubernetes Status to Framework, bootstrap Server and Mesos-CLI
- Add the possibility to set custom environment variables for the bootstrap.sh
- Remove unneeded API calls
- Add API call to get the current count of running and expected Kubernetes Server/Agents.

## v0.1.0

- Update Mesos Protobuf to the current Mesos Version
- Add custom MesosCNI via env
- Fix snapshooter, flannel and cgroups error
- Fix server/agent mismatch during shutdown
- Removes via api killed tasks from config. Same with scale down agents/servers.
- Add support of external etcd service
- Add configurable docker socket
- Add scale up/down from etcd
- Change container engine to to containerd
- Enable persistent storage for K3S server
- Add wildcard certificate for K3S to prevent certificate error during K3S server restart
- Enable Traefik and ServiceLB
- Add startup order
- Fix communication between pods via different mesos agents
- Add fetch custom bootstrap script to init kubernetes
- Change Mesos TASK ID's
- Change the K3S Image to plain ubuntu. The bootstrap will install K3S.
- Add Dashbaord configuration
- Add K3S Framework API Server into the Container
- Add API to get Kubeconfig
- Add Kubernetes server check to scheduler the agent only if the server is running
- Add automatic Kubernetes dashboard deploy to the API server
- Add seperated memory and CPU resources for ETCD and K3S
- Add Heatbeat for ETCD and K3S services
- Add automatic Traefik dashboard deploy
- Add mesos-cli plugin
- Mesos-CLI: Show all running M3s frameworks
- Mesos-CLI: Get kubeconfig from a selected running m3s
- Add supress offers if every services are running

## v0.0.4

- Add volume driver support

## v0.0.3

- Add persistent storage Support
- Change to go mod

## v0.0.2

- Add start mesos container
- Add start command
- Add persist framework info therefore the framework know what to do after a crash
- Add save task state
- Add start k3s
- Add persist task state
- Add multinode support
- Add container monitor to restart if a container failed
- Add RestCall to reflate missing k3s processes, for the case the monitoring can not find the problem
- Add RestCall scale up and down
- Add RestCall to kill a task
- Add Authentication
- Add Support to configure (non)SSL Support to Mesos
- Add Custom Domain to match external DNS (like Consul)
- Add Service Name ENV Variable to match external DNS (like Consule)
- Add Call_Suppress to tell mesos it does not send us offers until we ask
- Add default values for some env
- Add custom image name via env
- Add schedule start order
