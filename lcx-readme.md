CGO_ENABLED=0 GOOS=linux go build  -o ./nvidia_gpu_exporter ./cmd/nvidia_gpu_exporter/main.go

docker build -t 10.19.193.67:5000/arena-addon/gpu-exporter:v1 .