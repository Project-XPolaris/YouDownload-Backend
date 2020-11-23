package router

import (
	"github.com/julienschmidt/httprouter"
	"github.com/projectxpolaris/youdownload/backend/downloader"
	"net/http"
)

type TaskStatus struct {
	Id           string  `json:"id"`
	SavePath     string  `json:"save_path"`
	Url          string  `json:"url"`
	CompleteSize int64   `json:"complete_size"`
	Total        int64   `json:"total"`
	Speed        float64 `json:"speed"`
	Progress     float64 `json:"progress"`
	Filename     string  `json:"filename"`
	ETA          string  `json:"eta"`
	Status       string  `json:"status"`
}

func newTask(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	dest := r.FormValue("dest")
	downloadUrl := r.FormValue("url")
	downloader.DefaultDownloader.Pool.NewTaskChan <- downloader.NewTaskConfig{
		Url:  downloadUrl,
		Dest: dest,
	}
	w.Write([]byte("success"))
}
func getDownloadStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	taskInfos := make([]TaskStatus, 0)
	for _, task := range downloader.DefaultDownloader.Pool.Tasks {
		statusInfo := TaskStatus{
			Id:           task.Id,
			SavePath:     task.SavePath,
			Url:          task.Url,
			Status:       downloader.TaskStatusToTextMapping[task.Status],
		}
		if task.Response == nil {
			statusInfo.CompleteSize = task.SaveComplete
			statusInfo.Total = task.SaveTotal
			statusInfo.Progress = float64(task.SaveComplete) / float64(task.SaveTotal)
			statusInfo.Filename =  task.SaveFileName

		}else{
			statusInfo.CompleteSize = task.Response.BytesComplete()
			statusInfo.Total = task.Response.Size
			statusInfo.Progress = task.Response.Progress()
			statusInfo.ETA = task.Response.ETA().String()
			statusInfo.Filename =  task.Response.Filename
		}
		taskInfos = append(taskInfos, statusInfo)
	}
	WriteResponse(w, JsonFormat{
		"tasks": taskInfos,
	})
}
func pauseTask(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	taskId := r.FormValue("id")
	downloader.DefaultDownloader.Pool.PauseTask(taskId)
	WriteResponse(w, JsonFormat{
		"result": "success",
	})
}
func handleFile(router *httprouter.Router) {
	router.GET("/file/tasks", getDownloadStatus)
	router.POST("/file/tasks", newTask)
	router.POST("/file/task/pause", pauseTask)
}
