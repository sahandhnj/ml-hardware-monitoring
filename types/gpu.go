package types

import (
	"time"

	"github.com/sahandhnj/ml-hardware-monitoring/bindings/nvml"
)

type SnapShot struct {
	ID           int       `json:"id"`
	TimeStamp    time.Time `json:"time_stamp"`
	DeviceUUID   string    `json:"device_uuid"`
	Power        uint      `json:"power"`
	Temperature  uint      `json:"temprature"`
	GPU          uint      `json:"gpu"`
	Memory       uint      `json:"memory"`
	Encoder      uint      `json:"encoder"`
	Decoder      uint      `json:"decoder"`
	ClocksMemory uint      `json:"clocks_memory"`
	ClocksCores  uint      `json:"clocks_cores"`
}

func TakeSnapShot(deviceUUID string, st *nvml.DeviceStatus) *SnapShot {
	return &SnapShot{
		// Power:        *st.Power,
		TimeStamp:    time.Now(),
		DeviceUUID:   deviceUUID,
		Temperature:  *st.Temperature,
		GPU:          *st.Utilization.GPU,
		Memory:       *st.Utilization.Memory,
		Encoder:      *st.Utilization.Encoder,
		Decoder:      *st.Utilization.Decoder,
		ClocksMemory: *st.Clocks.Memory,
		ClocksCores:  *st.Clocks.Cores,
	}
}
