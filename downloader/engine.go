package downloader

import (
	"context"
	"fmt"
	"github.com/asdine/storm/q"
	"github.com/cavaliercoder/grab"
	"github.com/projectxpolaris/youdownload/backend/database"
	"github.com/rs/xid"
	"sync"
	"time"
)

type FileDownloaderEngine struct {
	Client    *grab.Client
	Pool      *TaskPool
	TaskStore *TaskStore
}

func NewFileDownloaderEngine() *FileDownloaderEngine {
	return &FileDownloaderEngine{
		Client: grab.NewClient(),
		Pool: &TaskPool{
			Tasks:       make([]*Task, 0),
			NewTaskChan: make(chan NewTaskConfig),
		},
		TaskStore: NewTaskStore(),
	}
}

type TaskPool struct {
	Tasks       []*Task
	NewTaskChan chan NewTaskConfig
	sync.RWMutex
}

func (p *TaskPool) PauseTask(taskId string) {
	defer p.Unlock()
	for _, task := range p.Tasks {
		if task.Id == taskId {
			p.Lock()
			task.Cancel()
			task.Status = TaskStatusPause
			DefaultDownloader.TaskStore.SaveChan <- NewTaskSaveInfoFromTask(task)
		}
	}
}
func (p *TaskPool) RemoveTask(taskId string) {
	targetIndex := -1
	for index, task := range p.Tasks {
		if task.Id == taskId {
			targetIndex = index
			break
		}
	}
	if targetIndex != -1 {
		p.Tasks = append(p.Tasks[:targetIndex], p.Tasks[targetIndex+1:]...)
	}
}

const (
	TaskStatusRunning = iota + 1
	TaskStatusPause
)

var TaskStatusToTextMapping = map[int64]string{
	TaskStatusRunning: "Running",
	TaskStatusPause:   "Pause",
}

type Task struct {
	Id       string
	Request  *grab.Request
	Response *grab.Response
	Url      string
	SavePath string
	Err      error
	Cancel   context.CancelFunc
	Status   int64
}
type NewTaskConfig struct {
	Url  string
	Dest string
}

func (e *FileDownloaderEngine) Run() {
	// for task store
	go func() {
		e.TaskStore.Run()
	}()
	// for interval save to database
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				for _, task := range e.Pool.Tasks {
					if task.Status == TaskStatusRunning {
						taskSaveInfo := NewTaskSaveInfoFromTask(task)
						e.TaskStore.SaveChan <- taskSaveInfo
					}
				}
			}
		}
	}()
	// for task pool
	go func() {
		for {
			select {
			case taskConfig := <-e.Pool.NewTaskChan:

				task := &Task{
					Id:       xid.New().String(),
					SavePath: taskConfig.Dest,
					Url:      taskConfig.Url,
					Status:   TaskStatusRunning,
				}

				var saveTask TaskSaveInfo
				err := database.Instance.Select(q.And(q.Eq("Url", taskConfig.Url), q.Eq("Dest", taskConfig.Dest))).First(&saveTask)
				if err == nil && len(saveTask.TaskId) > 0 {
					// exist
					task.Id = saveTask.TaskId
				}

				request, err := grab.NewRequest(taskConfig.Dest, taskConfig.Url)
				if err != nil {
					fmt.Println(err)
					task.Err = err
					return
				}
				ctx, cancel := context.WithCancel(context.Background())
				task.Cancel = cancel
				request = request.WithContext(ctx)
				fmt.Printf("Downloading %v...\n", request.URL())
				task.Request = request
				response := e.Client.Do(request)
				task.Response = response
				e.Pool.Lock()
				e.Pool.RemoveTask(task.Id)
				e.Pool.Tasks = append(e.Pool.Tasks, task)
				e.Pool.Unlock()
			}
		}
	}()
	select {}
}
