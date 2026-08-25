# BENZHI_README

这是一个 Go 后端应用，用于服务于职业本科教师的双师型认定、企业实践、技术服务、实训授课、成果佐证、评价与续认复核。

## 项目说明

- 项目：11DingKing/dual-teacher-practice-go
- 项目用途：该项目服务于职业本科教师的双师型认定、企业实践、技术服务、实训授课、成果佐证、评价与续认复核。学校人事、二级学院、企业导师、教研负责人和教师在同一条可追溯证据链上协作。
- Go 工具链：`golang:1.25.0`
- 前端工具链：无

## 标准构建、运行和测试命令

进入容器后执行：

```bash
# 编译
cd '/app' && GOTOOLCHAIN=local go build ./...

# 启动
cd '/app' && GOTOOLCHAIN=local go run ./cmd/server

# 测试
cd '/app' && GOTOOLCHAIN=local go test ./...
```

## Docker 构建和进入容器

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh benzhi-task-52-amd64 linux/amd64
./build_benzhi_docker.sh benzhi-task-52-arm64 linux/arm64
docker run -it benzhi-task-52-amd64:latest
docker run -it --platform linux/arm64 benzhi-task-52-arm64:latest
```

## 题目验证命令

1. 预期退出码 0：`go test ./internal/service -run '^TestApplicationSubmitRollsBackWhenAuditWriteFails$' -count=1`
2. 预期退出码 0：`go test ./...`
3. 预期退出码 0：`GOTOOLCHAIN=local go build -buildvcs=false ./... && GOTOOLCHAIN=local go vet ./...`
