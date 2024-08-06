package util

import (
	"encoding/json"
	"os"
	"strings"
)

const DefaultCheckpointFile = "/var/lib/kubelet/device-plugins/kubelet_internal_checkpoint"

// DevicesPerNUMA represents device ids obtained from device plugin per NUMA node id
type DevicesPerNUMA map[int64][]string

// Checksum is the data to be stored as checkpoint
type Checksum uint64

// PodDevicesEntry connects pod information to devices
type PodDevicesEntry struct {
	PodUID        string
	ContainerName string
	ResourceName  string
	DeviceIDs     DevicesPerNUMA
	AllocResp     []byte
}

// checkpointData struct is used to store pod to device allocation information
// in a checkpoint file.
// TODO: add version control when we need to change checkpoint format.
type checkpointData struct {
	PodDeviceEntries  []PodDevicesEntry
	RegisteredDevices map[string][]string
}

// Data holds checkpoint data and its checksum
type Data struct {
	Data     checkpointData
	Checksum Checksum
}

func New() *Data {
	registeredDevs := make(map[string][]string)
	devEntries := make([]PodDevicesEntry, 0)
	return &Data{
		Data: checkpointData{
			PodDeviceEntries:  devEntries,
			RegisteredDevices: registeredDevs,
		},
	}
}

func GetPodUIDByDeviceID(deviceID string) (string, error) {
	data, err := getData()
	if err != nil {
		return "", err
	}
	for _, entry := range data.Data.PodDeviceEntries {
		for _, deviceIDs := range entry.DeviceIDs {
			for _, id := range deviceIDs {
				// 包含deviceID
				if strings.Contains(id, deviceID) {
					return entry.PodUID, nil
				}
			}
		}
	}
	return "", nil
}

// Get data from file /var/lib/kubelet/device-plugins/kubelet_internal_checkpoint
func getData() (*Data, error) {
	//read checkpoint file path from env
	checkpointFile := os.Getenv("CHECKPOINT_FILE")
	if checkpointFile == "" {
		checkpointFile = DefaultCheckpointFile
	}

	// Read checkpoint file from disk
	blob, err := os.ReadFile(checkpointFile)
	if err != nil {
		return nil, err
	}
	// Unmarshal checkpoint data
	data := New()
	err = json.Unmarshal(blob, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
