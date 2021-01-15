package application

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/middleware"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

var (
	AppLogger *logrus.Logger = logrus.New()
)

func RunApiService() {
	e := haruka.NewEngine()
	e.UseMiddleware(middleware.NewLoggerMiddleware())
	e.Router.POST("/torrent/magnet", addMagnet)
	e.Router.POST("/torrent/file", addTorrentFileHandler)
	e.Router.POST("/torrent/start", startTorrentDownloadHandler)
	e.Router.POST("/torrent/stop", stopTorrentDownloadHandler)
	e.Router.POST("/torrent/del", deleteTorrentDownloadHandler)
	e.Router.GET("/torrent/setting",getTorrentSettingHandler)
	e.Router.POST("/torrent/setting",applyTorrentSettingHandler)
	e.Router.GET("/tasks", taskList)
	e.Router.POST("/file/task", addLinkHandler)
	e.Router.POST("/file/pause", pauseFileDownloadTask)
	e.Router.POST("/file/start", startFileDownloadTask)
	e.Router.POST("/file/delete", deleteDownloadTaskHandler)
	e.Router.POST("/file/limit", updateDownloadTaskLimitHandler)
	e.Router.POST("/util/readDir", readDirectoryHandler)
	go DefaultWatcher.RunEngineWatcher()
	e.UseCors(cors.AllowAll())
	e.RunAndListen(":7500")
}
