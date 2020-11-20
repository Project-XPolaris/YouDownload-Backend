package torrent

import (
	"github.com/asdine/storm"
	"github.com/projectxpolaris/youdownload/backend/database"
	log "github.com/sirupsen/logrus"
)

type TorrentDB struct {
	DB 		*storm.DB
}

func GetTorrentDB() *TorrentDB {
	var torrentDB	TorrentDB
	torrentDB.DB = database.Instance
	return &torrentDB
}

func (TorrentDB *TorrentDB)Cleanup()() {
	if TorrentDB.DB != nil {
		err := TorrentDB.DB.Close()
		if err != nil {
			logger.WithFields(log.Fields{"Detail":err}).Error("Failed to closed database")
		}
	}
}

func (TorrentDB *TorrentDB)GetLogs(torrentLogs *TorrentLogsAndID)() {
	torrentLogs.ID = TorrentLogsID
	err := TorrentDB.DB.One("ID", TorrentLogsID, torrentLogs)
	if err != nil {
		logger.WithFields(log.Fields{"Error":err}).Info("Init running queue now")
	}
	return
}









