# DHS 构建说明

分布式任务心跳服务，提供节点注册、心跳、失联巡检、恢复和审计查询 API。

```bash
go build ./...
go run ./cmd/heartbeat -config config.yaml
go test ./...
./build_benzhi_docker.sh dhs linux/amd64
./build_benzhi_docker.sh dhs linux/arm64
```

服务默认监听 `:8080`，数据存储在 `./data/heartbeat.db`。前端访问 `/`。
