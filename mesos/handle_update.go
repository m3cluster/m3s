package mesos

import (
	mesosutil "github.com/AVENTER-UG/mesos-util"

	mesosproto "github.com/AVENTER-UG/mesos-util/proto"

	"github.com/sirupsen/logrus"
)

// HandleUpdate will handle the offers event of mesos
func (e *Scheduler) HandleUpdate(event *mesosproto.Event) error {
	logrus.Debug("HandleUpdate")

	update := event.Update

	msg := &mesosproto.Call{
		Type: mesosproto.Call_ACKNOWLEDGE,
		Acknowledge: &mesosproto.Call_Acknowledge{
			AgentID: *update.Status.AgentID,
			TaskID:  update.Status.TaskID,
			UUID:    update.Status.UUID,
		},
	}

	// get the task of the current event, change the state and get some info's we need for later use
	task := e.API.GetTaskFromEvent(update)
	task.State = update.Status.State.String()

	// if these object have not TaskID it, invalid
	if task.TaskID == "" {
		return nil
	}

	logrus.Debug(task.State)

	switch *update.Status.State {
	case mesosproto.TASK_FAILED:
		// restart task
		task.State = ""
	case mesosproto.TASK_KILLED:
		// remove task
		e.API.DelRedisKey(task.TaskName + ":" + task.TaskID)
		return mesosutil.Call(msg)
	case mesosproto.TASK_LOST:
		// restart task
		task.State = ""
	case mesosproto.TASK_ERROR:
		// restart task
		task.State = ""
	case mesosproto.TASK_RUNNING:
		task.MesosAgent = mesosutil.GetAgentInfo(update.Status.GetAgentID().Value)
		task.NetworkInfo = mesosutil.GetNetworkInfo(task.TaskID)
		task.Agent = update.Status.GetAgentID().Value
	}

	// save the new state
	e.API.SaveTaskRedis(task)

	return mesosutil.Call(msg)
}
