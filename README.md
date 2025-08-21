Campus Blog - 后端服务（Go）

简要说明：这是一个校园博客系统的后端实现，演示了无状态认证、分布式缓存与异步写入的工程化实践，适合作为中大型系统后端架构练习与演示仓库。

关键特性

  用户认证：HS256 (JWT) + Redis 会话存储与续期，支持分布式无状态认证。
  
  异步写入：Kafka 生产端缓冲 + 消费端批量写入，显著降低数据库压力。
  
  热榜系统：Redis + 本地缓存 + 定时刷新，保证高命中率与低延时查询。
  
  容器化部署：提供 Dockerfile 与基础 k8s/Compose 示例。
  
  监控支持：Prometheus 指标埋点与 Grafana 基本面板。

架构图

  客户端 → API Gateway（Gin） → 后端服务（Go）
  
  后端服务接入 Redis（缓存/Session）、MySQL（持久化）、Kafka（事件流）
  
  消费端批量写入 MySQL，后台定时任务负责热榜刷新
  
快速开始（开发环境）

假设你本机已安装 Go、Docker、Docker Compose

克隆仓库

git clone https://github.com/kritozemu/compus_blog.git
cd compus_blog

启动依赖（使用 docker-compose）

docker-compose up -d
# 将启动：redis, mysql, zipkin, kafka, prometheus, etcd

本地运行服务

# 设置环境变量（示例）
export DB_DSN="root:123@tcp(127.0.0.1:13316)/compusdb?charset=utf8mb4&parseTime=True&loc=Local"
export REDIS_ADDR=127.0.0.1:6379
export KAFKA_ADDR=127.0.0.1:9092

# 运行
go run ./cmd/server

常用命令

运行单元测试：go test ./...

性能压测示例：使用 wrk 或 hey 对接口进行压测，仓库附有示例脚本 scripts/load_test.sh。

性能指标（项目压测结果）

写入峰值稳定达到 5,000+ TPS（在指定硬件与 docker-compose 环境下）。

数据库写入压力降低 ≈10×（通过 Kafka 批量写入与缓冲）。

热榜命中率 > 99%，常规查询延时 < 10ms。

注：性能数据为实验环境下结果，仅供参考。仓库内含压测脚本与说明，方便你复现。

设计要点与实现细节

认证：JWT 负责身份标识，Redis 存储会话与自动续期；采用 HS256 签名，保证轻量与快速验证；对关键接口加入速率限制与黑名单策略。

异步写入流程：API 将写请求发送到 Kafka，生产端采用本地缓冲（批量合并）以减少网络开销；消费端从 Kafka 拉取并批量写入 MySQL。

缓存策略：热数据保存在 Redis Sorted Set 中，并在本地进程做 LRU 缓存以减少网络调用；使用定时任务与事件驱动相结合更新缓存。

部署建议

将服务容器化并部署到 k8s 集群，使用 Horizontal Pod Autoscaler 做弹性扩缩容；

对 Kafka 与 MySQL 做资源隔离与监控告警（Prometheus + Grafana）；

根据流量调整 Kafka 分区数与消费者并发度，确保写入吞吐与数据顺序（若需要）。

如何贡献

Fork 本仓库并新建分支 feature/xxx。

提交 PR，包含用例/文档/变更说明。

若为性能优化，附带压测脚本与基准报告。

联系方式

如有问题或想复现环境，欢迎通过 GitHub Issues 或 Email 联系：2107920928@qq.com
