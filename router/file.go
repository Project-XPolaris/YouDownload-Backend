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
	taskInfos := make([]TaskStatus,0)
	for _, task := range downloader.DefaultDownloader.Pool.Tasks {
		statusInfo := TaskStatus{
			Id:           task.Id,
			CompleteSize: task.Response.BytesComplete(),
			Total:        task.Response.Size,
			Speed:        task.Response.BytesPerSecond(),
			Progress:     task.Response.Progress(),
			Filename:     task.Response.Filename,
			ETA:          task.Response.ETA().String(),
			SavePath:     task.SavePath,
			Url:          task.Url,
		}
		taskInfos = append(taskInfos, statusInfo)
	}
	WriteResponse(w,JsonFormat{
		"tasks":taskInfos,
	})
}

func handleFile(router *httprouter.Router)  {
	router.GET("/file/tasks", getDownloadStatus)
	router.POST("/file/tasks", newTask)
}