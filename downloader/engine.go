package downloader

import (
	"context"
	"fmt"
	"github.com/asdine/storm/q"
	"github.com/cavaliercoder/grab"
	"github.com/projectxpolaris/youdownload/backend/database"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"log"
	"sync"
	"time"
)

var Logger = logrus.New().WithField("scope", "TaskManager")

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
	for _, task := range p.Tasks {
		if task.Id == taskId {
			task.Cancel()
			task.Status = TaskStatusStopped
			DefaultDownloader.TaskStore.SaveChan <- NewTaskSaveInfoFromTask(task)
		}
	}
}

func (p *TaskPool) StartTask(taskId string) {

	for _, task := range p.Tasks {
		if task.Id == taskId {
			p.NewTaskChan <- NewTaskConfig{
				Url:  task.Url,
				Dest: task.SavePath,
			}
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
		task := p.Tasks[targetIndex]
		if task.Status == TaskStatusRunning {
			task.Cancel()
		}
		p.Tasks = append(p.Tasks[:targetIndex], p.Tasks[targetIndex+1:]...)
	}
}

func (p *TaskPool) DeleteTask(taskId string) error {
	targetIndex := -1
	for index, task := range p.Tasks {
		if task.Id == taskId {
			targetIndex = index
			break
		}
	}
	if targetIndex != -1 {
		task := p.Tasks[targetIndex]
		if task.Status == TaskStatusRunning {
			task.Cancel()
		}
		err := database.Instance.DeleteStruct(&TaskSaveInfo{TaskId: task.Id})
		if err != nil {
			return err
		}
		p.Tasks = append(p.Tasks[:targetIndex], p.Tasks[targetIndex+1:]...)
	}
	return nil
}

func (p *TaskPool) UpdateTaskLimiter(taskId string, rate int) error {
	targetIndex := -1
	for index, task := range p.Tasks {
		if task.Id == taskId {
			targetIndex = index
			break
		}
	}
	if targetIndex == -1 {
		return nil
	}

	// restart with limiter
	task := p.Tasks[targetIndex]
	if task.Status == TaskStatusCompleted {
		return nil
	}



	if task.Status == TaskStatusRunning {
		task.Cancel()
		p.NewTaskChan <- NewTaskConfig{
			Url:  task.Url,
			Dest: task.SavePath,
			UseLimit: true,
			Limit: rate,
		}
	}
	task.Limit = rate
	DefaultDownloader.TaskStore.SaveChan <- NewTaskSaveInfoFromTask(task)
	return nil
}

const (
	TaskStatusRunning = iota + 1
	TaskStatusStopped
	TaskStatusCompleted
)

var TaskStatusToTextMapping = map[int64]string{
	TaskStatusRunning:   "Running",
	TaskStatusStopped:   "Stopped",
	TaskStatusCompleted: "Completed",
}

type Task struct {
	Id           string
	Request      *grab.Request
	Response     *grab.Response
	Limit        int
	Url          string
	SavePath     string
	Err          error
	Cancel       context.CancelFunc
	Status       int64
	SaveComplete int64
	SaveTotal    int64
	SaveFileName string
}
type NewTaskConfig struct {
	Url   string
	Dest  string
	Limit int
	UseLimit bool
}

func (e *FileDownloaderEngine) Run() {
	// import save task
	var saveTasks []TaskSaveInfo
	err := database.Instance.All(&saveTasks)
	if err != nil {
		log.Fatal(err)
	}
	for _, saveTask := range saveTasks {
		e.Pool.Tasks = append(e.Pool.Tasks, &Task{
			Id:           saveTask.TaskId,
			Url:          saveTask.Url,
			SavePath:     saveTask.Dest,
			Status:       saveTask.Status,
			SaveComplete: saveTask.CompleteSize,
			SaveTotal:    saveTask.Total,
			SaveFileName: saveTask.Filename,
			Limit: saveTask.Limit,
		})
	}
	Logger.Info(fmt.Sprintf("success load %d task from database", len(e.Pool.Tasks)))
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
					taskSaveInfo := NewTaskSaveInfoFromTask(task)
					e.TaskStore.SaveChan <- taskSaveInfo
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
					Limit:    taskConfig.Limit,
				}



				// check duplicate in save task
				var saveTask TaskSaveInfo
				err := database.Instance.Select(q.And(q.Eq("Url", taskConfig.Url), q.Eq("Dest", taskConfig.Dest))).First(&saveTask)
				if err == nil && len(saveTask.TaskId) > 0 {
					// exist
					task.Id = saveTask.TaskId
					task.SaveFileName = saveTask.Filename

					if !taskConfig.UseLimit {
						task.Limit = saveTask.Limit
					}
				}

				// make request
				request, err := grab.NewRequest(taskConfig.Dest, taskConfig.Url)
				if err != nil {
					fmt.Println(err)
					task.Err = err
					return
				}
				if task.Limit != 0 {
					request.RateLimiter = NewLimiter(task.Limit)
				}
				ctx, cancel := context.WithCancel(context.Background())
				task.Cancel = cancel
				request = request.WithContext(ctx)
				Logger.WithField("id", task.Id).WithField("url", request.URL()).Info("Downloading")
				task.Request = request

				// request download url
				response := e.Client.Do(request)
				task.Response = response

				// update with request result
				task.SaveFileName = response.Filename
				e.Pool.Lock()
				e.Pool.RemoveTask(task.Id)
				e.Pool.Tasks = append(e.Pool.Tasks, task)
				e.Pool.Unlock()

				// update task save info
				saveInfo := NewTaskSaveInfoFromTask(task)
				e.TaskStore.SaveChan <- saveInfo

				//run for done chan
				go func() {
					select {
					case <-response.Done:
						task.Status = TaskStatusCompleted
						Logger.WithField("id", task.Id).Info("task complete")
						saveInfo = NewTaskSaveInfoFromTask(task)
						e.TaskStore.SaveChan <- saveInfo
					case <-ctx.Done():
						task.Status = TaskStatusStopped
						saveInfo = NewTaskSaveInfoFromTask(task)
						e.TaskStore.SaveChan <- saveInfo
						Logger.WithField("id", task.Id).Info("task interrupt")
					}
				}()
			}
		}
	}()
	select {}
}
