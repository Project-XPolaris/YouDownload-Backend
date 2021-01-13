package router

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/projectxpolaris/youdownload/backend/torrent"
	"github.com/projectxpolaris/youdownload/backend/setting"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func getSetting(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	WriteResponse(w, clientConfig.GetWebSetting())
}

func getStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	RunningEngine.TorrentEngine.WriteStatus(w)
}

func getRunningQueue(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var tmp torrent.TorrentLogsAndID
	RunningEngine.TorrentDB.GetLogs(&tmp)
	WriteResponse(w, tmp)
}

func applySetting(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	isApplied := false
	var newSettings setting.WebSetting
	err := decoder.Decode(&newSettings)
	if err != nil {
		logger.WithFields(log.Fields{"Error": err}).Error("Failed to get new settings")
	}else{
		if RunningEngine.EngineRunningInfo.HasRestarted == false {
			RunningEngine.EngineRunningInfo.HasRestarted = true
			clientConfig.UpdateConfig(newSettings)
			logger.WithFields(log.Fields{"Settings": newSettings}).Info("Setting update")
			isApplied = true
			RunningEngine.Restart()
			RunningEngine.EngineRunningInfo.HasRestarted = false
		}
	}
	WriteResponse(w, JsonFormat{
		"IsApplied":isApplied,
	})
}

func handleSetting(router *httprouter.Router)  {
	router.GET("/settings/config", getSetting)
	router.GET("/settings/status", getStatus)
	router.GET("/settings/queue", getRunningQueue)
	router.POST("/settings/apply", applySetting)
}