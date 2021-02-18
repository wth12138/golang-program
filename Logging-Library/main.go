package main

import (
	"Logging-Library/mylog"
)

//var log mylog.Logger

func main() {
	//log := mylog.NewFileLogger("INFO", "./", "wth.log", 10*1024*1024)
	log := mylog.NewFileLogger("INFO", "./", "wth.log", 1*1024*1024)
	log := mylog.
	for {
		log.Debug("this is debug log")
		log.Info("this is info log")
		log.Warning("this is warning log")
		id := 10010
		name := "test"
		log.Error("this is error log,id:%d,name:%s", id, name)
		log.Fatal("this is fatal log")

	}
}
