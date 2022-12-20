package scheduler

import (
	"time"

	"github.com/sirupsen/logrus"
)

var reviveLock bool
var suppressLock bool

// Heartbeat function for mesos
func (e *Scheduler) Heartbeat() {
	// Check Connection state of Redis
	err := e.Redis.PingRedis()
	if err != nil {
		e.Redis.Connect()
	}

	dsState := e.healthCheckDatastore()
	k3sState := e.healthCheckK3s()
	k3sAgenteState := e.healthCheckAgent()

	if !k3sState || !k3sAgenteState || !dsState {
		if !reviveLock {
			e.Mesos.Revive()
			reviveLock = true
		}
		suppressLock = false
	}

	// if DataStorage container is not running or unhealthy, fix it.
	if !dsState {
		e.StartDatastore("")
	}

	// if Datastorage is running and K3s not, deploy K3s
	if dsState && !k3sState {
		e.StartK3SServer("")
	}

	// if k3s is running, deploy the agent
	if k3sState && !k3sAgenteState {
		e.StartK3SAgent("")
	}

	if k3sState && k3sAgenteState && dsState {
		if !suppressLock {
			e.Mesos.SuppressFramework()
			suppressLock = true
			reviveLock = false
		}
		e.API.ScheduleCleanup()
	}
}

// CheckState check the current state of every task
func (e *Scheduler) CheckState() {
	keys := e.Redis.GetAllRedisKeys(e.Framework.FrameworkName + ":*")

	for keys.Next(e.Redis.CTX) {
		// get the values of the current key
		key := e.Redis.GetRedisKey(keys.Val())

		if e.Redis.CheckIfNotTask(keys) {
			continue
		}

		task := e.Mesos.DecodeTask(key)

		if task.TaskID == "" || task.TaskName == "" {
			continue
		}

		if task.State == "" && e.Redis.CountRedisKey(task.TaskName+":*", "") <= task.Instances {
			e.Mesos.Revive()
			task.State = "__NEW"

			// these will save the current time at the task. we need it to check
			// if the state will change in the next 'n min. if not, we have to
			// give these task a recall.
			task.StateTime = time.Now()

			// Change the Dynamic Host Ports
			task.DockerPortMappings = e.changeDockerPorts(task)
			task.Discovery = e.changeDiscoveryInfo(task)

			// add task to communication channel
			e.Framework.CommandChan <- task

			e.Redis.SaveTaskRedis(task)

			logrus.Info("Scheduled Mesos Task: ", task.TaskName)
		}

		// Remove corrupt tasks
		if task.State == "" && task.StateTime.Year() == 1 {
			e.Redis.DelRedisKey(task.TaskName + ":" + task.TaskID)
		}
	}
}

// HeartbeatLoop - The main loop for the hearbeat
func (e *Scheduler) HeartbeatLoop() {
	suppressLock = false
	reviveLock = true
	ticker := time.NewTicker(e.Config.EventLoopTime)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		go e.Heartbeat()
		go e.CheckState()
	}
}

// ReconcileLoop - The reconcile loop to check periodicly the task state
func (e *Scheduler) ReconcileLoop() {
	ticker := time.NewTicker(e.Config.ReconcileLoopTime)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		go e.reconcile()
		go e.implicitReconcile()
	}
}
