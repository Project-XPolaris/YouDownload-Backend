package torrent

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/projectxpolaris/youdownload/backend/setting"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

type Engine struct {
	TorrentEngine     *torrent.Client
	TorrentDB         *TorrentDB
	WebInfo           *WebviewInfo
	EngineRunningInfo *EngineInfo
}

var (
	onlyEngine       Engine
	hasCreatedEngine = false
	clientConfig     = setting.GetClientSetting()
	logger           = clientConfig.LoggerSetting.Logger
)

func GetEngine() *Engine {
	if hasCreatedEngine == false {
		onlyEngine.initAndRunEngine()
		hasCreatedEngine = true
	}
	return &onlyEngine
}

func (engine *Engine) initAndRunEngine() () {
	engine.TorrentDB = GetTorrentDB()

	var tmpErr error
	engine.TorrentEngine, tmpErr = torrent.NewClient(&clientConfig.EngineSetting.TorrentConfig)
	if tmpErr != nil {
		logger.WithFields(log.Fields{"Error": tmpErr}).Error("Failed to Created torrent torrent")
	}

	engine.WebInfo = &WebviewInfo{}
	engine.WebInfo.HashToTorrentWebInfo = make(map[metainfo.Hash]*TorrentWebInfo)

	engine.EngineRunningInfo = &EngineInfo{}
	engine.EngineRunningInfo.init()

	//Get info from storm database
	engine.setEnvironment()
}

func (engine *Engine) setEnvironment() () {

	engine.TorrentDB.GetLogs(&engine.EngineRunningInfo.TorrentLogsAndID)

	logger.Debug("Number of torrent(s) in db is ", len(engine.EngineRunningInfo.TorrentLogs))

	for _, singleLog := range engine.EngineRunningInfo.TorrentLogs {

		if singleLog.Status != CompletedStatus {
			_, tmpErr := engine.TorrentEngine.AddTorrent(&singleLog.MetaInfo)
			if tmpErr != nil {
				logger.WithFields(log.Fields{"Error": tmpErr}).Info("Failed to add torrent to client")
			}

			// set file pior
			tmpTorrent, isExist := engine.TorrentEngine.Torrent(singleLog.HashInfoBytes())
			if isExist {
				for _, torrentFile := range tmpTorrent.Files() {
					for _, fileConfig := range singleLog.Files {
						if fileConfig.Path == torrentFile.Path() {
							ApplyPriority(torrentFile, fileConfig.Priority)
						}
					}

				}
			}
		}
	}
	engine.UpdateInfo()
}

func (engine *Engine) Restart() () {

	logger.Info("Restart torrent")

	//To handle problems caused by change of settings
	for index := range engine.EngineRunningInfo.TorrentLogs {
		torrentLog := engine.EngineRunningInfo.TorrentLogs[index]
		// not complete but store changed
		if torrentLog.Status != CompletedStatus && torrentLog.StoragePath != clientConfig.TorrentConfig.DataDir {
			filePath := filepath.Join(torrentLog.StoragePath, torrentLog.TorrentName)
			log.WithFields(log.Fields{"Path": filePath}).Info("To restart torrent, these unfinished files will be deleted")
			singleTorrent, torrentExist := engine.GetOneTorrent(torrentLog.HashInfoBytes().HexString())
			if torrentExist {
				singleTorrent.Drop()
			}
			torrentLog.StoragePath = clientConfig.TorrentConfig.DataDir
			engine.UpdateInfo()
			delFiles(filePath)
		}
	}
	engine.Cleanup()
	GetEngine()

}

// save Torrent Logs to DB
func (engine *Engine) SaveInfo() () {
	//fmt.Printf("%+v\n", torrent.EngineRunningInfo.TorrentLogsAndID);
	db := GetTorrentDB()
	tmpErr := db.DB.Save(&engine.EngineRunningInfo.TorrentLogsAndID)
	if tmpErr != nil {
		logger.WithFields(log.Fields{"Error": tmpErr}).Fatal("Failed to save torrent queues")
	}
	//fmt.Println("save it successfully")
}

func (engine *Engine) Cleanup() () {
	hasCreatedEngine = false
	engine.UpdateInfo()
	for index := range engine.EngineRunningInfo.TorrentLogs {
		torrentLog := engine.EngineRunningInfo.TorrentLogs[index]
		if torrentLog.Status != CompletedStatus {
			if torrentLog.Status == AnalysingStatus {
				aimLog := torrentLog
				torrentHash := metainfo.Hash{}
				_ = torrentHash.FromHexString(aimLog.TorrentName)
				magnetTorrent, isExist := engine.TorrentEngine.Torrent(torrentHash)
				if isExist {
					logger.Info("One magnet will be deleted " + magnetTorrent.String())
					magnetTorrent.Drop()
				}
			} else if torrentLog.Status == RunningStatus {
				engine.StopOneTorrent(torrentLog.HashInfoBytes().HexString())
				torrentLog.Status = StoppedStatus
			} else if torrentLog.Status == QueuedStatus {
				torrentLog.Status = StoppedStatus
			}
		}
	}

	//Update info in torrentLogs, remove magnet
	tmpLogs := engine.EngineRunningInfo.TorrentLogs
	engine.EngineRunningInfo.TorrentLogs = nil

	for index := range tmpLogs {
		if tmpLogs[index].Status != AnalysingStatus {
			engine.EngineRunningInfo.TorrentLogs = append(engine.EngineRunningInfo.TorrentLogs, tmpLogs[index])
		}
	}

	engine.SaveInfo()

	engine.TorrentEngine.Close()
}
