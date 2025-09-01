# 可通过 --build-arg 覆盖 GO_VERSION，例如： --build-arg GO_VERSION=1.24.1
ARG GO_VERSION=1.24.1
FROM golang:${GO_VERSION} AS builder

WORKDIR /app

# 从 build context（通常是仓库根）拷贝 go.mod/go.sum
COPY go.mod go.sum ./

RUN go mod download

# 拷贝整个仓库到镜像（合理利用 .dockerignore）
COPY . .

# 构建 binary（源码位于 ./basic/interactive）
RUN CGO_ENABLED=0 GOOS=linux go build -o interactive ./basic/interactive

# 运行阶段
FROM debian:bullseye AS runner
WORKDIR /app
COPY --from=builder /app/interactive .
CMD ["./interactive", "--config", "/app/configs/interactive.yaml"]
