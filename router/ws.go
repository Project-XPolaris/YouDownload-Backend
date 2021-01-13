package router

import (
	"github.com/projectxpolaris/youdownload/backend/torrent"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//TODO : close handle
func torrentProgress (w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	logger.Info("websocket created!")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Unable to init websocket", err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()
	var tmp torrent.MessageFromWeb
	var resInfo torrent.TorrentProgressInfo
	for {
		select {
			case cmdID := <- RunningEngine.EngineRunningInfo.EngineCMD: {
				logger.Debug("Send CMD Now", cmdID)
				if cmdID == torrent.RefreshInfo {
					resInfo.MessageType = torrent.RefreshInfo
					err = conn.WriteJSON(resInfo)
					if err != nil {
						logger.Error("Unable to write Message", err)
					}
				}
			}
			default:
		}
		err = conn.ReadJSON(&tmp)
		if err != nil {
			logger.Error("Unable to read Message", err)
			break
		}

		if tmp.MessageType == torrent.GetInfo {
			singleTorrent, isExist := RunningEngine.GetOneTorrent(tmp.HexString)
			if isExist {
				singleTorrentLog, _ := RunningEngine.EngineRunningInfo.HashToTorrentLog[singleTorrent.InfoHash()]
				if singleTorrentLog.Status == torrent.RunningStatus || singleTorrentLog.Status == torrent.CompletedStatus {
					singleWebLog := RunningEngine.GenerateInfoFromTorrent(singleTorrent)
					resInfo.MessageType = torrent.GetInfo
					resInfo.HexString = singleWebLog.HexString
					resInfo.Percentage = singleWebLog.Percentage
					resInfo.LeftTime = singleWebLog.LeftTime
					resInfo.DownloadSpeed = singleWebLog.DownloadSpeed
					_ = conn.WriteJSON(resInfo)
				}
			}
		}

	}

}

func handleWS (router *httprouter.Router)  {
	router.GET("/ws", torrentProgress)
}

