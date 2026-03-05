# Go 后端 - 微服务与云原生面试知识库

## 1. Go 微服务框架

### go-micro / go-kit / Kratos 对比
| 维度 | go-micro | go-kit | Kratos (B站) |
|------|----------|--------|-------------|
| 定位 | 全功能微服务框架 | 工具包 | 企业级微服务框架 |
| 服务发现 | 内置多种（Consul/etcd） | 需自行集成 | 内置（Discovery） |
| RPC | gRPC / HTTP | 任意 | gRPC + HTTP |
| 配置中心 | 内置 | 需自行集成 | 内置 |
| 学习曲线 | 中等 | 陡峭 | 中等 |

### gRPC 与 Protobuf
```protobuf
syntax = "proto3";
package user;

service UserService {
    rpc GetUser (GetUserRequest) returns (UserResponse);
    rpc ListUsers (ListUsersRequest) returns (stream UserResponse); // 服务端流
}

message GetUserRequest {
    int64 id = 1;
}
```

**gRPC 四种通信模式：**
1. Unary（一元）：一请求一响应
2. Server Streaming：一请求多响应
3. Client Streaming：多请求一响应
4. Bidirectional Streaming：双向流

### 标准答案要点（5星回答）
1. Protobuf 序列化为什么比 JSON 快（二进制 + 字段编号）
2. gRPC 的 4 种通信模式
3. 服务发现的方式（客户端发现 vs 服务端发现）
4. 中间件/拦截器的使用
5. 有 Go 微服务项目的实际经验

---

## 2. Docker 与 Kubernetes

### Docker 核心
- **镜像分层**：UnionFS，每层只读，容器层可写
- **多阶段构建**：减小最终镜像体积
```dockerfile
# 构建阶段
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o server .

# 运行阶段
FROM alpine:3.19
COPY --from=builder /app/server /server
CMD ["/server"]
```

### Kubernetes 核心概念
| 资源 | 作用 |
|------|------|
| Pod | 最小部署单元，一个或多个容器 |
| Deployment | 管理 Pod 副本，支持滚动更新 |
| Service | 服务发现和负载均衡 |
| ConfigMap / Secret | 配置和敏感数据管理 |
| Ingress | HTTP 路由入口 |
| HPA | 水平自动扩缩容 |

### 标准答案要点（5星回答）
1. Docker 镜像分层原理
2. K8s Pod 的生命周期和探针（Liveness / Readiness / Startup）
3. Service 的 ClusterIP / NodePort / LoadBalancer 模式
4. Deployment 滚动更新策略
5. 有 K8s 部署和运维经验

---

## 3. etcd 与分布式一致性

### etcd 核心
- 基于 Raft 共识算法的分布式 KV 存储
- 应用场景：服务注册发现、配置中心、分布式锁、Leader 选举

### Raft 算法要点
1. **Leader 选举**：随机超时 → 候选人发起投票 → 获得多数票当选
2. **日志复制**：Leader 接收写请求 → 广播给 Followers → 多数确认后提交
3. **安全性**：只有包含最新日志的节点才能当选 Leader

### 标准答案要点（5星回答）
1. Raft 的三种角色（Leader / Follower / Candidate）
2. 日志复制的流程
3. 脑裂和网络分区的处理
4. etcd 的 Watch 机制和 Lease 机制
5. 与 ZooKeeper (ZAB) 的对比
