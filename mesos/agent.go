package mesos

import (
	"encoding/json"
	"strings"

	mesosutil "github.com/AVENTER-UG/mesos-util"
	mesosproto "github.com/AVENTER-UG/mesos-util/proto"
	"github.com/AVENTER-UG/util"

	"github.com/sirupsen/logrus"
)

// StartK3SAgent is starting a agent container with the given IDs
func StartK3SAgent(taskID string) {
	var cmd mesosutil.Command

	// if taskID is 0, then its a new task and we have to create a new ID
	newTaskID := taskID
	if taskID == "" {
		newTaskID, _ = util.GenUUID()
	}

	cni := framework.MesosCNI

	hostport := getRandomHostPort(2)
	if hostport == 0 {
		logrus.WithField("func", "StartK3SAgent").Error("Could not find free ports")
		return
	}
	protocol := "tcp"

	cmd.TaskID = newTaskID

	cmd.ContainerType = "DOCKER"
	cmd.ContainerImage = config.ImageK3S

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
	cmd.Memory = config.K3SAgentMEM
	cmd.CPU = config.K3SAgentCPU
	cmd.TaskName = framework.FrameworkName + ":agent"
	cmd.Hostname = framework.FrameworkName + "agent" + config.Domain
	cmd.Command = "$MESOS_SANDBOX/bootstrap '" + config.K3SAgentString + config.K3SDocker + " --with-node-id " + newTaskID + "'"
	cmd.DockerParameter = addDockerParameter(make([]mesosproto.Parameter, 0), mesosproto.Parameter{Key: "cap-add", Value: "NET_ADMIN"})
	cmd.DockerParameter = addDockerParameter(make([]mesosproto.Parameter, 0), mesosproto.Parameter{Key: "cap-add", Value: "SYS_ADMIN"})
	cmd.DockerParameter = addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "shm-size", Value: config.DockerSHMSize})
	// if mesos cni is unset, then use docker cni
	if framework.MesosCNI == "" {
		// net-alias is only supported onuser-defined networks
		if config.DockerCNI != "bridge" {
			cmd.NetworkMode = "user"
			cni = config.DockerCNI
			cmd.DockerParameter = addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "net-alias", Value: framework.FrameworkName + "agent"})
		}
	}
	cmd.NetworkInfo = []mesosproto.NetworkInfo{{
		Name: &cni,
	}}

	cmd.Uris = []mesosproto.CommandInfo_URI{
		{
			Value:      config.BootstrapURL,
			Extract:    func() *bool { x := false; return &x }(),
			Executable: func() *bool { x := true; return &x }(),
			Cache:      func() *bool { x := false; return &x }(),
			OutputFile: func() *string { x := "bootstrap"; return &x }(),
		},
	}

	cmd.Discovery = mesosproto.DiscoveryInfo{
		Visibility: 2,
		Name:       &cmd.TaskName,
		Ports: &mesosproto.Ports{
			Ports: []mesosproto.Port{
				{
					Number:   cmd.DockerPortMappings[0].HostPort,
					Name:     func() *string { x := strings.ToLower(framework.FrameworkName) + "-http"; return &x }(),
					Protocol: cmd.DockerPortMappings[0].Protocol,
				},
				{
					Number:   cmd.DockerPortMappings[1].HostPort,
					Name:     func() *string { x := strings.ToLower(framework.FrameworkName) + "-https"; return &x }(),
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
	// store mesos task in DB
	d, _ := json.Marshal(&cmd)
	logrus.Debug("Scheduled K3S Agent: ", string(d))
	logrus.Info("Scheduled K3S Agent")
	err := config.RedisClient.Set(config.RedisCTX, cmd.TaskName+":"+newTaskID, d, 0).Err()
	if err != nil {
		logrus.Error("Cloud not store Mesos Task in Redis: ", err)
	}
}
