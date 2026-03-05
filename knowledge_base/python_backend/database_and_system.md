# Python 后端 - 数据库与系统设计面试知识库

## 1. SQLAlchemy ORM

### 核心概念
- Engine：数据库连接池入口
- Session：工作单元（Unit of Work），管理对象状态
- Model：声明式映射（Declarative Mapping）
- Query：查询构建器，链式调用

### N+1 查询问题
```python
# 问题：循环中每次查询关联对象
users = session.query(User).all()
for user in users:
    print(user.orders)  # 每个用户触发一次 SQL

# 解决：预加载
# 方案1：joinedload（LEFT JOIN 一次查出）
users = session.query(User).options(joinedload(User.orders)).all()

# 方案2：subqueryload（子查询批量加载）
users = session.query(User).options(subqueryload(User.orders)).all()
```

### 标准答案要点（5星回答）
1. 理解 Session 的生命周期管理
2. 知道 N+1 问题及预加载解决方案
3. 了解 SQLAlchemy 2.0 的 select() 新风格
4. 事务管理和 rollback 处理
5. 连接池配置（pool_size, max_overflow, pool_recycle）

---

## 2. Celery 异步任务

### 核心架构
```
Producer → Broker (Redis/RabbitMQ) → Worker → Result Backend
```

### 关键配置
- `task_serializer`: json（安全）vs pickle（支持复杂对象）
- `task_acks_late`: 消费确认策略（True = 执行完再 ACK）
- `task_reject_on_worker_lost`: Worker 异常重启是否重试
- `worker_concurrency`: 并发 Worker 数

### 常见问题
1. **任务幂等性**：同一任务可能被多次执行（Worker 重启），需设计幂等
2. **任务超时**：`task_time_limit` 硬超时 + `task_soft_time_limit` 软超时
3. **任务优先级**：多队列 + 路由实现

### 标准答案要点（5星回答）
1. Celery 的架构和消息流转
2. 定时任务 celery beat 的使用
3. 任务链/组合：chain、group、chord
4. 生产环境的监控和运维（Flower）
5. 替代方案了解：FastAPI BackgroundTasks、Dramatiq、Huey

---

## 3. Python 系统设计

### 微服务通信
- **HTTP REST**：requests / httpx（异步）
- **gRPC**：protobuf 高性能序列化，Python 适合做 gRPC 客户端
- **消息队列**：Redis Pub/Sub、RabbitMQ（pika）、Kafka（confluent-kafka）

### 部署架构
```
Nginx → Gunicorn (WSGI) → Django/Flask
Nginx → Uvicorn (ASGI) → FastAPI

# Gunicorn 推荐配置
workers = multiprocessing.cpu_count() * 2 + 1
worker_class = "gthread"  # 线程模式
timeout = 30
```

### 标准答案要点（5星回答）
1. WSGI vs ASGI 的区别和使用场景
2. Gunicorn 的 Worker 模型（sync/gthread/gevent）
3. Docker 容器化部署经验
4. CI/CD 流水线配置
5. 日志与监控（ELK / Prometheus + Grafana）
