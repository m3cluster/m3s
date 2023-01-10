package api

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// V0ScaleK3SServer will scale the k3s server service
// example:
// curl -X GET http://user:password@127.0.0.1:10000/v0/server/scale/{count of instances} -d 'JSON'
func (e *API) V0ScaleK3SServer(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0ScaleK3SServer").Debug("Call")

	vars := mux.Vars(r)

	if vars == nil || !e.CheckAuth(r, w) {
		return
	}

	d := e.ErrorMessage(0, "V0ScaleK3SServer", "ok")

	if vars["count"] != "" {
		count, err := strconv.Atoi(vars["count"])
		if err != nil {
			logrus.WithField("func", "api.V0K3SServer").Error("Error: ", err.Error())
			return
		}
		d = e.scaleServer(count)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write(d)
}

func (e *API) scaleServer(count int) []byte {
	r := e.scale(count, e.Config.K3SServerMax, ":server:*")
	e.Config.K3SServerMax = count
	return r
}

func (e *API) scale(newCount int, oldCount int, key string) []byte {
	logrus.WithField("func", "api.scale").Debug("Scale "+key+" current: ", oldCount)

	d := []byte(strconv.Itoa(newCount - oldCount))

	// Save current config
	e.Redis.SaveConfig(*e.Config)

	// if scale down, kill not needes agents
	keys := e.Redis.GetAllRedisKeys(e.Framework.FrameworkName + key)

	for keys.Next(e.Redis.CTX) {
		key := e.Redis.GetRedisKey(keys.Val())
		task := e.Mesos.DecodeTask(key)
		task.Instances = newCount
		e.Redis.SaveTaskRedis(task)

		if newCount < oldCount {
			e.Mesos.Kill(task.TaskID, task.Agent)
			logrus.WithField("func", "api.scale").Debug("TaskID: ", task.TaskID)
		}
		if newCount > oldCount {
			e.Mesos.Revive()
		}
		oldCount = oldCount - 1
	}

	return d
}
