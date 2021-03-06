package downloader

import (
	"github.com/projectxpolaris/youdownload/backend/database"
	"github.com/sirupsen/logrus"
)

type TaskSaveInfo struct {
	TaskId       string `storm:"id"`
	Url          string `storm:"index"`
	Dest         string `storm:"index"`
	CompleteSize int64  `storm:"index"`
	Total        int64  `storm:"index"`
	Status       int64  `storm:"index"`
	Filename     string `storm:"index"`
	Limit        int    `storm:"index"`
}

func NewTaskSaveInfoFromTask(task *Task) *TaskSaveInfo {
	info := &TaskSaveInfo{
		TaskId: task.Id,
		Url:    task.Url,
		Dest:   task.SavePath,
		Status: task.Status,
		Limit:  task.Limit,
	}
	if task.Response != nil {
		info.CompleteSize = task.Response.BytesComplete()
		info.Total = task.Response.Size
		info.Filename = task.Response.Filename
	} else {
		info.CompleteSize = task.SaveComplete
		info.Total = task.SaveTotal
		info.Filename = task.SaveFileName
	}

	return info
}

type TaskStore struct {
	SaveChan chan *TaskSaveInfo
}

func NewTaskStore() *TaskStore {
	return &TaskStore{
		SaveChan: make(chan *TaskSaveInfo),
	}
}

func saveInfo(info *TaskSaveInfo) {
	err := database.Instance.Save(info)
	if err != nil {
		logrus.WithError(err)
	}
}
func (s *TaskStore) Run() {
	for {
		select {
		case info := <-s.SaveChan:
			//logrus.Info(fmt.Sprintf("save id = %s", info.TaskId))
			saveInfo(info)
		}
	}
}
