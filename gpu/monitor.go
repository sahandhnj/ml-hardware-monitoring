package gpu

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
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

func (g *GPU) RunForProcess() {
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
					message := types.TakeMessage(st, getCPUUsage())
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

				pInfo, err := device.GetAllRunningProcesses()
				if err != nil {
					log.Panicf("Error getting device %d processes: %v\n", i, err)
				}
				if len(pInfo) == 0 && false {
					fmt.Printf("%5v %5s %5s %5s %-5s\n", i, "-", "-", "-", "-")
				}
				// for j := range pInfo {
				// 	fmt.Printf("%5v %5v %5v %5v %-5v\n",
				// 		i, pInfo[j].PID, pInfo[j].Type, pInfo[j].MemoryUsed, pInfo[j].Name)
				// }

				go func() {
					// snapshot := types.TakeSnapShot(device.UUID, st)
					message := types.TakeMessage(st, getCPUUsage())

					// g.DBService.SnapShotService.CreateSnapshot(snapshot)
					g.Broadcast <- *message
				}()

			}
		case <-sigs:
			panic("BOOOOO")
			return
		}
	}
}

func getCPUSample() (idle, total uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					fmt.Println("Error: ", i, fields[i], err)
				}
				total += val
				if i == 4 {
					idle = val
				}
			}
			return
		}
	}
	return
}

func getCPUUsage() float64 {
	idle0, total0 := getCPUSample()
	time.Sleep(5 * time.Millisecond)
	idle1, total1 := getCPUSample()

	idleTicks := float64(idle1 - idle0)
	totalTicks := float64(total1 - total0)
	cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks

	return cpuUsage
}
