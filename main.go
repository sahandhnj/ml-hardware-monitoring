package main

import (
	"time"

	"github.com/sahandhnj/ml-hardware-monitoring/db"
	"github.com/sahandhnj/ml-hardware-monitoring/gpu"
)

func main() {
	dbService, err := db.NewDBService()
	if err != nil {
		panic(err)
	}

	GPU := &gpu.GPU{
		DBService: dbService,
		Interval:  time.Second * 1,
	}

	GPU.Run()
}
