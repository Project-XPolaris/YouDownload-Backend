package router

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/projectxpolaris/youdownload/backend/torrent"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func addOneTorrentFromFile(w http.ResponseWriter, r *http.Request, ps httprouter.Params)  {
	//Get torrent file from form
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		logger.WithFields(log.Fields{"Error":err}).Error("Unable to parse form")
		return
	}
	file, handler, err := r.FormFile("oneTorrentFile")

	if err != nil {
		logger.WithFields(log.Fields{"Error":err}).Error("Unable to get file from form")
		return
	}

	defer file.Close()

	filePath := filepath.Join(clientConfig.EngineSetting.Tmpdir, handler.Filename)
	filePathAbs, _ := filepath.Abs(filePath)

	f, err := os.OpenFile(filePathAbs, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		logger.WithFields(log.Fields{"Error":err}).Error("Unable to copy file from form")
		return
	}

	//Start to add to client
	tmpTorrent, err := RunningEngine.AddOneTorrentFromFile(filePathAbs)
	
	var isAdded bool
	if err != nil {
		logger.WithFields(log.Fields{"Error":err}).Error("unable to add a torrent")
		isAdded = false
	}else{
		if tmpTorrent != nil {
			RunningEngine.GenerateInfoFromTorrent(tmpTorrent)
			RunningEngine.StartDownloadTorrent(tmpTorrent.InfoHash().HexString())
			isAdded = true
		}
	}

	WriteResponse(w, JsonFormat{
		"IsAdded":isAdded,
	})

}

func getOneTorrent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hexString := r.FormValue("hexString")
	singleTorrent, isExist := RunningEngine.GetOneTorrent(hexString)
	if isExist {
		torrentWebInfo := RunningEngine.GenerateInfoFromTorrent(singleTorrent)
		WriteResponse(w, torrentWebInfo)
	}else{
		w.WriteHeader(http.StatusNotFound)
	}
}


func AppendRunningTorrents(resInfo []torrent.TorrentWebInfo)([]torrent.TorrentWebInfo) {
	for _, singleTorrent := range RunningEngine.TorrentEngine.Torrents() {
		singleTorrentLog, isExist := RunningEngine.EngineRunningInfo.HashToTorrentLog[singleTorrent.InfoHash()]
		if isExist && singleTorrentLog.Status != torrent.CompletedStatus {
			resInfo = append(resInfo, *RunningEngine.GenerateInfoFromTorrent(singleTorrent))
		}
	}
	return resInfo
}

func AppendCompletedTorrents(resInfo []torrent.TorrentWebInfo)([]torrent.TorrentWebInfo) {
	for _, singleTorrentLog := range RunningEngine.EngineRunningInfo.TorrentLogs {
		if singleTorrentLog.Status == torrent.CompletedStatus {
			resInfo = append(resInfo, *RunningEngine.GenerateInfoFromLog(singleTorrentLog))
		}
	}
	return resInfo
}


func getAllTorrents(w http.ResponseWriter, r *http.Request, ps httprouter.Params)  {
	resInfo := []torrent.TorrentWebInfo{}
	resInfo = AppendRunningTorrents(resInfo)
	resInfo = AppendCompletedTorrents(resInfo)
	WriteResponse(w, resInfo)
}

func getCompletedTorrents(w http.ResponseWriter, r *http.Request, ps httprouter.Params)  {
	resInfo := []torrent.TorrentWebInfo{}
	resInfo = AppendCompletedTorrents(resInfo)
	WriteResponse(w, resInfo)
}



func getAllEngineTorrents(w http.ResponseWriter, r *http.Request, ps httprouter.Params){
	resInfo := []torrent.TorrentWebInfo{}
	resInfo = AppendRunningTorrents(resInfo)
	WriteResponse(w, resInfo)
}

func delOneTorrent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hexString := r.FormValue("hexString")
	deleted := RunningEngine.DelOneTorrent(hexString)
	WriteResponse(w, JsonFormat{
		"IsDeleted":deleted,
	})
}

func stopOneTorrent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hexString := r.FormValue("hexString")
	stopped := RunningEngine.StopOneTorrent(hexString)
	WriteResponse(w, JsonFormat{
		"IsStopped":stopped,
	})
}

func setTorrentFilePriority(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hexString := r.FormValue("hexString")
	filePath := r.FormValue("filePath")
	rawLevel := r.FormValue("level")
	level,err := strconv.Atoi(rawLevel)
	if err != nil {
		WriteResponse(w, JsonFormat{
			"success":false,
		})
		return
	}
	RunningEngine.SetFilePriority(hexString,filePath,level)
	WriteResponse(w, JsonFormat{
		"success":true,
	})
}
func startDownloadTorrent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hexString := r.FormValue("hexString")
	downloaded := RunningEngine.StartDownloadTorrent(hexString)
	WriteResponse(w, JsonFormat{
		"IsDownloading":downloaded,
	})
}

func test(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

func handleTorrent(router *httprouter.Router)  {
	router.POST("/torrent/addOneFile", addOneTorrentFromFile)
	router.POST("/torrent/getOne", getOneTorrent)
	router.GET("/torrent/getAllEngineTorrents", getAllEngineTorrents)
	router.GET("/torrent/getAllTorrents", getAllTorrents)
	router.GET("/torrent/getCompletedTorrents", getCompletedTorrents)
	router.POST("/torrent/delOne", delOneTorrent)
	router.POST("/torrent/startDownload", startDownloadTorrent)
	router.POST("/torrent/stopDownload", stopOneTorrent)
	router.POST("/torrent/setFilePriority", setTorrentFilePriority)
	router.GET("/torrent/test", test)
}
