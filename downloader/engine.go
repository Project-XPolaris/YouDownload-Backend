package downloader

import (
	"context"
	"fmt"
	"github.com/cavaliercoder/grab"
	"github.com/rs/xid"
	"sync"
)

type FileDownloaderEngine struct {
	Client *grab.Client
	Pool *TaskPool
}

func NewFileDownloaderEngine() *FileDownloaderEngine {
	return &FileDownloaderEngine{
		Client: grab.NewClient(),
		Pool: &TaskPool{
			Tasks: make([]*Task,0),
			NewTaskChan: make(chan NewTaskConfig),
		},
	}
}

type TaskPool struct {
	Tasks []*Task
	NewTaskChan chan NewTaskConfig
	sync.RWMutex
}
type Task struct {
	Id string
	Request *grab.Request
	Response *grab.Response
	Url string
	SavePath string
	Err error
	Cancel context.CancelFunc
}
type NewTaskConfig struct {
	Url string
	Dest string
}
func (e *FileDownloaderEngine) Run() {
	go func() {
		for {
			taskConfig := <- e.Pool.NewTaskChan
			task := &Task{
				Id: xid.New().String(),
				SavePath: taskConfig.Dest,
				Url: taskConfig.Url,
			}

			request,err  := grab.NewRequest(taskConfig.Dest,taskConfig.Url)
			if err != nil {
				fmt.Println(err)
				task.Err = err
				return
			}
			ctx,cancel := context.WithCancel(context.Background())
			task.Cancel = cancel
			request = request.WithContext(ctx)
			fmt.Printf("Downloading %v...\n", request.URL())
			task.Request = request
			response := e.Client.Do(request)
			task.Response = response
			e.Pool.Lock()
			defer e.Pool.Unlock()
			e.Pool.Tasks = append(e.Pool.Tasks, task)
		}
	}()
	select {

	}
}