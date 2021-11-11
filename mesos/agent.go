package mesos

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync/atomic"

	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"

	cfg "github.com/AVENTER-UG/mesos-m3s/types"

	"github.com/sirupsen/logrus"
)

// SearchMissingK3SAgent Check if all agents are running. If one is missing, restart it.
func SearchMissingK3SAgent(restart bool) bool {
	status := make([]mesosproto.TaskState, config.K3SAgentMax)
	allRunning := true
	if config.State != nil {
		for i := 0; i < config.K3SAgentMax; i++ {
			state := StatusK3SAgent(i)
			if state != nil {
				status[i] = *state.Status.State
				if *state.Status.State != mesosproto.TASK_RUNNING {
					allRunning = false
					logrus.Debug("Missing K3SAgent: ", i)
					if restart {
						StartK3SAgent(i)
					}
				}
			} else {
				allRunning = false
			}
		}
	} else {
		allRunning = false
	}
	config.M3SStatus.Agent = status
	return allRunning
}

// StatusK3SAgent Get out Status of the given agent
func StatusK3SAgent(id int) *cfg.State {
	if config.State != nil {
		for _, element := range config.State {
			if element.Status != nil {
				if element.Command.InternalID == id && element.Command.IsK3SAgent {
					return &element
				}
			}
		}
	}
	return nil
}

// StartK3SAgent is starting a agent container with the given IDs
func StartK3SAgent(id int) {
	newTaskID := atomic.AddUint64(&config.TaskID, 1)

	var cmd cfg.Command

	// before we will start a new agent, we should be sure its not already running
	status := StatusK3SAgent(id)
	if status != nil {
		if status.Status.State == mesosproto.TASK_STAGING.Enum() {
			logrus.Info("startK3SAgent: kubernetes is staging ", id)
			return
		}
		if status.Status.State == mesosproto.TASK_STARTING.Enum() {
			logrus.Info("startK3SAgent: kubernetes is starting ", id)
			return
		}
		if status.Status.State == mesosproto.TASK_RUNNING.Enum() {
			logrus.Info("startK3SAgent: kubernetes already running ", id)
			return
		}
	}

	var hostport uint32
	hostport = 31859 + uint32(newTaskID)
	protocol := "tcp"

	cmd.TaskID = newTaskID
	taskID := strconv.FormatUint(newTaskID, 10)

	cmd.ContainerType = "DOCKER"
	cmd.ContainerImage = config.ImageK3S
	cmd.NetworkMode = "bridge"
	cmd.NetworkInfo = []mesosproto.NetworkInfo{{
		Name: &config.MesosCNI,
	}}
	cmd.DockerPortMappings = []mesosproto.ContainerInfo_DockerInfo_PortMapping{
		{
			HostPort:      hostport,
			ContainerPort: 80,
			Protocol:      &protocol,
		},
		{
			HostPort:      hostport + 1,
			ContainerPort: 443,
			Protocol:      &protocol,
		},
	}

	cmd.Shell = true
	cmd.Privileged = true
	cmd.InternalID = id
	cmd.Memory = config.K3SMEM
	cmd.CPU = config.K3SCPU
	cmd.TaskName = config.PrefixTaskName + "agent"
	cmd.Hostname = config.PrefixTaskName + "agent" + strconv.Itoa(id) + "." + config.Domain
	cmd.Command = "$MESOS_SANDBOX/bootstrap '" + config.K3SAgentString + " --with-node-id " + taskID + "'"
	cmd.DockerParameter = []mesosproto.Parameter{
		{
			Key:   "cap-add",
			Value: "NET_ADMIN",
		},
	}
	cmd.Uris = []mesosproto.CommandInfo_URI{
		{
			Value:      config.BootstrapURL,
			Extract:    func() *bool { x := false; return &x }(),
			Executable: func() *bool { x := true; return &x }(),
			Cache:      func() *bool { x := false; return &x }(),
			OutputFile: func() *string { x := "bootstrap"; return &x }(),
		},
	}
	cmd.IsK3SAgent = true
	cmd.Volumes = []mesosproto.Volume{
		{
			ContainerPath: "/opt/cni/net.d",
			Mode:          mesosproto.RW.Enum(),
			Source: &mesosproto.Volume_Source{
				Type: mesosproto.Volume_Source_DOCKER_VOLUME,
				DockerVolume: &mesosproto.Volume_Source_DockerVolume{
					Name: "/etc/mesos/cni/net.d",
				},
			},
		},
	}
	if config.DockerSock != "" {
		cmd.Volumes = []mesosproto.Volume{
			{
				ContainerPath: "/var/run/docker.sock",
				Mode:          mesosproto.RW.Enum(),
				Source: &mesosproto.Volume_Source{
					Type: mesosproto.Volume_Source_DOCKER_VOLUME,
					DockerVolume: &mesosproto.Volume_Source_DockerVolume{
						Name: config.DockerSock,
					},
				},
			},
			{
				ContainerPath: "/opt/cni/net.d",
				Mode:          mesosproto.RW.Enum(),
				Source: &mesosproto.Volume_Source{
					Type: mesosproto.Volume_Source_DOCKER_VOLUME,
					DockerVolume: &mesosproto.Volume_Source_DockerVolume{
						Name: "/etc/mesos/cni/net.d",
					},
				},
			},
		}
	}

	cmd.Discovery = mesosproto.DiscoveryInfo{
		Visibility: 2,
		Name:       &cmd.TaskName,
		Ports: &mesosproto.Ports{
			Ports: []mesosproto.Port{
				{
					Number:   cmd.DockerPortMappings[0].HostPort,
					Name:     func() *string { x := strings.ToLower(config.FrameworkName) + "-http"; return &x }(),
					Protocol: cmd.DockerPortMappings[0].Protocol,
				},
				{
					Number:   cmd.DockerPortMappings[1].HostPort,
					Name:     func() *string { x := strings.ToLower(config.FrameworkName) + "-https"; return &x }(),
					Protocol: cmd.DockerPortMappings[1].Protocol,
				},
			},
		},
	}

	cmd.Environment.Variables = []mesosproto.Environment_Variable{
		{
			Name:  "SERVICE_NAME",
			Value: &cmd.TaskName,
		},
		{
			Name:  "K3SFRAMEWORK_TYPE",
			Value: func() *string { x := "agent"; return &x }(),
		},
		{
			Name:  "K3S_TOKEN",
			Value: &config.K3SToken,
		},
		{
			Name:  "K3S_URL",
			Value: &config.K3SServerURL,
		},
		{
			Name:  "MESOS_SANDBOX_VAR",
			Value: &config.MesosSandboxVar,
		},
	}

	if config.K3SAgentLabels != nil {
		cmd.Labels = config.K3SAgentLabels
	}

	if config.K3SAgentLabels != nil {
		cmd.Labels = config.K3SAgentLabels
	}

	d, _ := json.Marshal(&cmd)
	logrus.Debug("Scheduled K3S Agent: ", string(d))

	config.CommandChan <- cmd
	logrus.Info("Scheduled K3S Agent")
}

// The first run have to be in a right sequence
func initStartK3SAgent() {
	serverState := StatusK3SServer(config.K3SServerMax - 1)

	if serverState == nil {
		return
	}

	if config.K3SAgentCount <= (config.K3SAgentMax-1) && serverState.Status.GetState() == 1 {
		if IsK3SServerRunning() {
			StartK3SAgent(config.K3SAgentCount)
			config.K3SAgentCount++
		}
	}
}
