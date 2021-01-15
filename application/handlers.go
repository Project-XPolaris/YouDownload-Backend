package application

import (
	"fmt"
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youdownload/backend/downloader"
	"github.com/projectxpolaris/youdownload/backend/service"
	"github.com/projectxpolaris/youdownload/backend/setting"
	"github.com/projectxpolaris/youdownload/backend/torrent"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
)

var taskList haruka.RequestHandler = func(context *haruka.Context) {
	err := context.JSON(map[string]interface{}{
		"tasks": DefaultWatcher.tasks,
	})
	if err != nil {
		logrus.Error(err)
	}
}

type AddMagnetRequestBody struct {
	Link string `json:"link"`
}

var addMagnet haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody AddMagnetRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		Abort500Error(err, context)
		return
	}
	e := torrent.GetEngine()
	_, err = e.AddOneTorrentFromMagnet(requestBody.Link)
	if err != nil {
		Abort500Error(err, context)
		return
	}
	SendSuccessResponse(context)
}
var addTorrentFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	//Get torrent file from form
	r := context.Request
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		Abort500Error(err, context)
		return
	}
	file, handler, err := r.FormFile("oneTorrentFile")

	if err != nil {
		Abort500Error(err, context)
		return
	}

	defer file.Close()
	conf := setting.GetClientSetting()
	filePath := filepath.Join(conf.EngineSetting.Tmpdir, handler.Filename)
	filePathAbs, _ := filepath.Abs(filePath)

	f, err := os.OpenFile(filePathAbs, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		Abort500Error(err, context)
		return
	}

	//Start to add to client
	e := torrent.GetEngine()
	tmpTorrent, err := e.AddOneTorrentFromFile(filePathAbs)
	if err != nil {
		Abort500Error(err, context)
		return
	}
	if tmpTorrent != nil {
		e.GenerateInfoFromTorrent(tmpTorrent)
		e.StartDownloadTorrent(tmpTorrent.InfoHash().HexString())
	}
	SendSuccessResponse(context)
}

type StartTorrentDownloadRequestBody struct {
	Id string `json:"id"`
}

var startTorrentDownloadHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody StartTorrentDownloadRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		Abort500Error(err, context)
		return
	}
	e := torrent.GetEngine()
	e.StartDownloadTorrent(requestBody.Id)
	SendSuccessResponse(context)
}

type StopTorrentDownloadRequestBody struct {
	Id string `json:"id"`
}

var stopTorrentDownloadHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody StartTorrentDownloadRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		Abort500Error(err, context)
		return
	}
	e := torrent.GetEngine()
	e.StopOneTorrent(requestBody.Id)
	SendSuccessResponse(context)
}
type DeleteTorrentDownloadRequestBody struct {
	Id string `json:"id"`
}

var deleteTorrentDownloadHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody DeleteTorrentDownloadRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		Abort500Error(err, context)
		return
	}
	e := torrent.GetEngine()
	e.DelOneTorrent(requestBody.Id)
	SendSuccessResponse(context)
}

var getTorrentSettingHandler haruka.RequestHandler = func(context *haruka.Context) {
	context.JSON(setting.GetClientSetting().GetWebSetting())
}

var applyTorrentSettingHandler haruka.RequestHandler = func(context *haruka.Context) {
	var settingRequestBody setting.WebSetting
	err := context.ParseJson(&settingRequestBody)
	if err != nil {
		Abort500Error(err, context)
		return
	}
	e := torrent.GetEngine()
	config := setting.GetClientSetting()
	if e.EngineRunningInfo.HasRestarted == false {
		e.EngineRunningInfo.HasRestarted = true
		config.UpdateConfig(settingRequestBody)
		e.Restart()
		e.EngineRunningInfo.HasRestarted = false
	}
	SendSuccessResponse(context)
}
type AddLinkRequestBody struct {
	Link string `json:"link"`
	Dest string `json:"dest"`
}

var addLinkHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody AddLinkRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		Abort500Error(err, context)
		return
	}
	downloader.DefaultDownloader.Pool.NewTaskChan <- downloader.NewTaskConfig{
		Url:  requestBody.Link,
		Dest: requestBody.Dest,
	}
	SendSuccessResponse(context)
}


type StartFileDownloadRequestBody struct {
	Id string `json:"id"`
}

var startFileDownloadTask haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody StartFileDownloadRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		Abort500Error(err, context)
		return
	}
	downloader.DefaultDownloader.Pool.StartTask(requestBody.Id)
	SendSuccessResponse(context)
}
type PauseFileDownloadRequestBody struct {
	Id string `json:"id"`
}

var pauseFileDownloadTask haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody PauseFileDownloadRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		Abort500Error(err, context)
		return
	}
	downloader.DefaultDownloader.Pool.PauseTask(requestBody.Id)
	SendSuccessResponse(context)
}
type DeleteFileDownloadRequestBody struct {
	Id string `json:"id"`
}
var deleteDownloadTaskHandler haruka.RequestHandler  = func(context *haruka.Context) {
	var requestBody DeleteFileDownloadRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		Abort500Error(err, context)
		return
	}
	err = downloader.DefaultDownloader.Pool.DeleteTask(requestBody.Id)
	if err != nil {
		Abort500Error(err, context)
		return
	}
	SendSuccessResponse(context)
}

type ReadDirectoryRequestBody struct {
	Path string `json:"path"`
}
var readDirectoryHandler haruka.RequestHandler  = func(context *haruka.Context) {
	var requestBody ReadDirectoryRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		Abort500Error(err, context)
		return
	}
	if len(requestBody.Path) == 0{
		homePath,err := os.UserHomeDir()
		if err != nil {
			Abort500Error(err, context)
			return
		}
		requestBody.Path = homePath
	}
		items,err := service.ReadDirectory(requestBody.Path)
	if err != nil {
		Abort500Error(err, context)
		return
	}
	abs,_ := filepath.Abs(requestBody.Path)
	context.JSON(map[string]interface{}{
		"path":abs,
		"sep":string(filepath.Separator),
		"files":items,
	})
}