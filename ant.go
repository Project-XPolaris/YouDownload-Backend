package main

import (
	"github.com/projectxpolaris/youdownload/backend/application"
	"github.com/projectxpolaris/youdownload/backend/database"
	"github.com/projectxpolaris/youdownload/backend/downloader"
	"github.com/projectxpolaris/youdownload/backend/torrent"
	"github.com/projectxpolaris/youdownload/backend/router"
	"github.com/projectxpolaris/youdownload/backend/setting"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var (
	clientConfig   = setting.GetClientSetting()
	logger         = clientConfig.LoggerSetting.Logger
	torrentEngine  *torrent.Engine
	nRouter        *negroni.Negroni
)

func runAPP() {
	go func() {
		// Init server router
		nRouter = router.InitRouter()
		err := http.ListenAndServe(clientConfig.ConnectSetting.Addr, nRouter)
		if err != nil {
			logger.WithFields(log.Fields{"Error": err}).Fatal("Failed to created http service")
		}

	}()
}

func runFileDownloader () {
	go func() {
		downloader.DefaultDownloader.Run()
	}()
}
func cleanUp() {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)
		<-c
		log.Info("The progame will stop!")
		torrentEngine.Cleanup()
		os.Exit(0)
	}()
}


func main() {
	database.InitDB(clientConfig.EngineSetting.DBPath)
	runFileDownloader()
	application.RunApiService()
	cleanUp()
	runtime.Goexit()
}
