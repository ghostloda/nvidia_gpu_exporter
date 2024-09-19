CGO_ENABLED=0 GOOS=linux go build  -o ./nvidia_gpu_exporter ./cmd/nvidia_gpu_exporter/main.go

docker build -t 10.19.193.67:5000/arena-addon/gpu-exporter:v1 .

```
 volumeMounts:
        - mountPath: /var/lib/kubelet/pod-resources
          name: pod-gpu-resources
          readOnly: true
        - mountPath: /var/lib/kubelet/device-plugins
          name: device-plugin
...
...
...
 volumes:
      - hostPath:
          path: /var/lib/kubelet/pod-resources
          type: ""
        name: pod-gpu-resources
      - hostPath:
          path: /var/lib/kubelet/device-plugins
          type: ""
        name: device-plugin     

```


support rename resource name

```
        env:
        - name: NVIDIA_RESOURCE_NAME
          value: arena.com/nvidia-gpu-t4
```