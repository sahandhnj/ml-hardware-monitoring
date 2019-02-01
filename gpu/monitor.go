package gpu

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sahandhnj/ml-hardware-monitoring/bindings/nvml"
	"github.com/sahandhnj/ml-hardware-monitoring/db"
	"github.com/sahandhnj/ml-hardware-monitoring/types"
)

type GPU struct {
	DBService *db.DBService
	Interval  time.Duration
	Broadcast chan types.Message
}

func (g *GPU) Run() {
	nvml.Init()
	defer nvml.Shutdown()
	fmt.Println("Start recording...")

	count, err := nvml.GetDeviceCount()
	if err != nil {
		log.Panicln("Error getting device count:", err)
	}

	fmt.Printf("Found %d devices\n", count)

	var devices []*nvml.Device
	for i := uint(0); i < count; i++ {
		device, err := nvml.NewDevice(i)
		if err != nil {
			log.Panicf("Error getting device %d: %v\n", i, err)
		}
		devices = append(devices, device)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(g.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for i, device := range devices {
				st, err := device.Status()
				if err != nil {
					log.Panicf("Error getting device %d status: %v\n", i, err)
				}
				go func() {
					snapshot := types.TakeSnapShot(device.UUID, st)
					message := types.TakeMessage(st)
					fmt.Println(message)

					g.DBService.SnapShotService.CreateSnapshot(snapshot)
					g.Broadcast <- *message
				}()

			}
		case <-sigs:
			panic("BOOOOO")
			return
		}
	}
}
