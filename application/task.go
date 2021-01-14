package application

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/projectxpolaris/youdownload/backend/downloader"
	"github.com/projectxpolaris/youdownload/backend/torrent"
	"sync"
	"time"
)

var DefaultWatcher = Watcher{
	tasks: []*TaskStatus{},
}
var (
	TaskTypeTorrent = "Torrent"
	TaskTypeFile    = "File"
)

type TaskStatus struct {
	Id        string  `json:"id"`
	Name      string  `json:"name"`
	Progress  float64 `json:"progress"`
	TotalSize string  `json:"total_size"`
	Status    string  `json:"status"`
	Speed     string  `json:"speed"`
	Type      string  `json:"type"`
}
type Watcher struct {
	sync.RWMutex
	tasks []*TaskStatus
}

func (w *Watcher) RunEngineWatcher() {
	go func() {
		for {
			<-time.After(1 * time.Second)
			w.Lock()
			w.tasks = []*TaskStatus{}
			resInfo := []torrent.TorrentWebInfo{}
			resInfo = AppendRunningTorrents(resInfo)
			resInfo = AppendCompletedTorrents(resInfo)
			for _, torrentTask := range resInfo {
				newTask := &TaskStatus{}
				newTask.UpdateWithTorrent(&torrentTask)
				w.tasks = append(w.tasks, newTask)
			}
			for _, fileDownloadTask := range downloader.DefaultDownloader.Pool.Tasks {
				newTask := &TaskStatus{}
				newTask.UpdateWithFileDownloadTask(fileDownloadTask)
				w.tasks = append(w.tasks, newTask)
			}
			w.Unlock()
		}
	}()
}

func (t *TaskStatus) UpdateWithTorrent(torrentTask *torrent.TorrentWebInfo) {
	t.Id = torrentTask.HexString
	t.Name = torrentTask.TorrentName
	t.Progress = torrentTask.Percentage
	t.Speed = torrentTask.DownloadSpeed
	t.TotalSize = torrentTask.TotalLength
	t.Status = torrentTask.Status
	t.Type = TaskTypeTorrent
}

func (t *TaskStatus) UpdateWithFileDownloadTask(task *downloader.Task) {
	t.Id = task.Id
	t.Type = TaskTypeFile
	if task.Response != nil {
		t.Progress = task.Response.Progress()
		t.Speed = fmt.Sprintf("%s/s", humanize.Bytes(uint64(task.Response.BytesPerSecond())))
		t.TotalSize = humanize.Bytes(uint64(task.Response.Size))
	} else {
		t.Progress = float64(task.SaveComplete) / float64(task.SaveTotal)
		t.TotalSize = humanize.Bytes(uint64(task.SaveTotal))
	}
	t.Name = task.SaveFileName
	t.Status = downloader.TaskStatusToTextMapping[task.Status]

}

func AppendRunningTorrents(resInfo []torrent.TorrentWebInfo) []torrent.TorrentWebInfo {
	for _, singleTorrent := range torrent.GetEngine().TorrentEngine.Torrents() {
		singleTorrentLog, isExist := torrent.GetEngine().EngineRunningInfo.HashToTorrentLog[singleTorrent.InfoHash()]
		if isExist && singleTorrentLog.Status != torrent.CompletedStatus {
			resInfo = append(resInfo, *torrent.GetEngine().GenerateInfoFromTorrent(singleTorrent))
		}
	}
	return resInfo
}

func AppendCompletedTorrents(resInfo []torrent.TorrentWebInfo) []torrent.TorrentWebInfo {
	for _, singleTorrentLog := range torrent.GetEngine().EngineRunningInfo.TorrentLogs {
		if singleTorrentLog.Status == torrent.CompletedStatus {
			resInfo = append(resInfo, *torrent.GetEngine().GenerateInfoFromLog(singleTorrentLog))
		}
	}
	return resInfo
}
