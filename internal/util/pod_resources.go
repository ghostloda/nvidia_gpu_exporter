package util

import (
	"context"
	"fmt"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	podresourcesapi "k8s.io/kubelet/pkg/apis/podresources/v1"
)

var (
	connectionTimeout  = 10 * time.Second
	nvidiaResourceName = "nvidia.com/gpu"
	kubeletSocketPath  = "/var/lib/kubelet/pod-resources/kubelet.sock"
)

type PodInfo struct {
	Name      string
	Namespace string
	Container string
}

func GetPodByID(deviceID string) (PodInfo, error) {
	// read the kubelet socket path from the environment
	socketPath := os.Getenv("KUBELET_SOCKET_PATH")
	if socketPath == "" {
		socketPath = kubeletSocketPath
	}
	podInfo := PodInfo{}

	_, err := os.Stat(socketPath)
	if os.IsNotExist(err) {
		return podInfo, fmt.Errorf("kubelet socket path %s does not exist", socketPath)
	}

	conn, cleanup, err := connectToServer(socketPath)
	if err != nil {
		return podInfo, err
	}
	defer cleanup()

	devicePods, err := listPods(conn)
	if err != nil {
		return podInfo, err
	}

	deviceToPodMap := toDeviceToPod(devicePods)
	podInfo, ok := deviceToPodMap[deviceID]
	if !ok {
		return podInfo, fmt.Errorf("deviceID %s not found in pod resources", deviceID)
	}

	return podInfo, nil
}

func connectToServer(socket string) (*grpc.ClientConn, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx,
		socket,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "unix", addr)
		}),
	)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failure connecting to '%s'; err: %w", socket, err)
	}

	return conn, func() { conn.Close() }, nil
}

func listPods(conn *grpc.ClientConn) (*podresourcesapi.ListPodResourcesResponse, error) {
	client := podresourcesapi.NewPodResourcesListerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	resp, err := client.List(ctx, &podresourcesapi.ListPodResourcesRequest{})
	if err != nil {
		return nil, fmt.Errorf("failure getting pod resources; err: %w", err)
	}

	return resp, nil
}

func toDeviceToPod(devicePods *podresourcesapi.ListPodResourcesResponse) map[string]PodInfo {
	deviceToPodMap := make(map[string]PodInfo)
	for _, pod := range devicePods.GetPodResources() {
		for _, container := range pod.GetContainers() {
			for _, device := range container.GetDevices() {
				resourceName := device.GetResourceName()
				if resourceName != nvidiaResourceName {
					continue
				}
				podInfo := PodInfo{
					Name:      pod.GetName(),
					Namespace: pod.GetNamespace(),
					Container: container.GetName(),
				}

				for _, deviceID := range device.GetDeviceIds() {
					if strings.Contains(deviceID, "::") {
						gpuInstanceID := strings.Split(deviceID, "::")[0]
						deviceToPodMap[gpuInstanceID] = podInfo
					}
					// Default mapping between deviceID and pod information
					deviceToPodMap[deviceID] = podInfo
				}
			}
		}
	}
	return deviceToPodMap
}
