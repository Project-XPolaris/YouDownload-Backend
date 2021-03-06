package router

import (
	"github.com/julienschmidt/httprouter"
	"github.com/projectxpolaris/youdownload/backend/torrent"
	"net/http"
	"time"
)

func startPlay(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hexString := ps.ByName("hexString")
	singleTorrent, isExist := RunningEngine.GetOneTorrent(hexString)
	fileServed := false
	if isExist {
		singleTorrentLog := RunningEngine.EngineRunningInfo.HashToTorrentLog[singleTorrent.InfoHash()]
		if singleTorrentLog.Status == torrent.RunningStatus || singleTorrentLog.Status == torrent.CompletedStatus {
			fileEntry, target, err := RunningEngine.GetReaderFromTorrent(singleTorrent, "")
			if err != nil {
				logger.Error("Unable to get reader : ", err)
			} else {
				defer fileEntry.Close()
				fileServed = true
				w.Header().Set("Content-Disposition", "attachment; filename=\""+singleTorrent.Info().Name+"\"")
				logger.Info("serve it now")
				http.ServeContent(w, r, target.DisplayPath(), time.Now(), fileEntry)
			}
		}
	}
	if !fileServed {
		w.WriteHeader(http.StatusNotFound)
	}
	logger.Debug("Play has done")
}

func handlePlayer(router *httprouter.Router)  {
	router.GET("/player/:hexString", startPlay)
}