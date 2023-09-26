package app

import (
	"log"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/zhigulin-io/word-stat/internal/server"
	"github.com/zhigulin-io/word-stat/internal/storage"
	"github.com/zhigulin-io/word-stat/internal/task"
)

func Run() {
	log.Println("Loading configs...")
	storageConf := storage.LoadConfig("config/storage.yaml")
	serverConf := server.LoadConfig("config/server.yaml")
	log.Println("Configs loaded!")

	log.Println("Creating storage...")
	storage := storage.NewPGStorage(storageConf)
	defer storage.Close()
	log.Println("Storage created!")

	log.Println("Creating server...")
	serv := server.New(serverConf, storage)
	log.Println("Server created!")

	statUpdateTask := task.StatUpdateTask{
		Storage: storage,
	}

	log.Println("Registring tasks...")
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.Cron("*/5 * * * *").Do(statUpdateTask.Action)
	log.Println("Tasks registred!")

	log.Println("=== LET'S START! ===")
	scheduler.StartAsync()
	serv.Serve()
}
