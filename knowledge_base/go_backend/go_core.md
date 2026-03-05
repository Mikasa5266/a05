# Go 后端 - 语言核心面试知识库

## 1. Goroutine 与调度模型

### GMP 调度模型
- **G (Goroutine)**：轻量级协程，初始栈 2KB（可动态扩缩）
- **M (Machine)**：OS 线程，执行 G 的实体
- **P (Processor)**：逻辑处理器，维护本地 G 队列，数量默认等于 CPU 核数

**调度流程：**
1. 新 G 创建 → 放入当前 P 的本地队列
2. P 的本地队列满 → 打散一半放入全局队列
3. M 绑定 P 执行 G
4. G 执行系统调用 → M 与 P 分离 → P 寻找空闲 M 继续执行
5. 本地队列为空 → Work Stealing 偷取其他 P 的 G

**抢占式调度（Go 1.14+）：**
- 基于信号的异步抢占（SIGURG），解决了死循环 Goroutine 无法被调度的问题

### 标准答案要点（5星回答）
1. GMP 三者的角色和关系画出来
2. 描述 Work Stealing 和 Hand Off 机制
3. Go 1.14 信号抢占的触发条件
4. Goroutine vs 线程的区别（栈大小、切换成本、调度方式）
5. 有高并发场景使用 Goroutine 的经验

---

## 2. Channel 与并发模式

### Channel 基础
```go
ch := make(chan int)     // 无缓冲：发送阻塞到接收方就绪
ch := make(chan int, 10) // 有缓冲：缓冲满时阻塞发送

// 方向限制
func producer(ch chan<- int) { ... } // 只发送
func consumer(ch <-chan int) { ... } // 只接收
```

### 常见并发模式
```go
// 1. Fan-Out / Fan-In
jobs := make(chan Job)
results := make(chan Result)
for i := 0; i < numWorkers; i++ {
    go worker(jobs, results)
}

// 2. Pipeline
func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- n * n
        }
    }()
    return out
}

// 3. Context 超时控制
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
select {
case result := <-doWork(ctx):
    // 正常完成
case <-ctx.Done():
    // 超时处理
}
```

### 标准答案要点（5星回答）
1. 无缓冲 vs 有缓冲 Channel 的行为差异
2. select 多路复用的工作方式
3. 至少说出 3 种并发模式（Pipeline、Fan-Out/Fan-In、Worker Pool）
4. Context 的使用（WithTimeout/WithCancel/WithValue）
5. 了解 Channel 底层的 hchan 结构和环形缓冲区

---

## 3. 内存管理与 GC

### 内存分配
- **TCMalloc 启发**：mcache（每个 P 本地）→ mcentral（全局）→ mheap（堆）
- **大小分类**：tiny（<16B）→ small（16B-32KB）→ large（>32KB 直接堆分配）
- **逃逸分析**：编译器决定变量分配在栈还是堆（`go build -gcflags="-m"`）

### 垃圾回收（三色标记 + 混合写屏障）
**三色标记法：**
- 白色：未访问，GC 结束后回收
- 灰色：已访问但引用对象未全部扫描
- 黑色：已访问且引用对象全部扫描完

**并发 GC 流程：**
1. STW → 开启写屏障 → 标记根对象为灰色
2. 并发标记（与 mutator 并发）
3. STW → 重新扫描标记变化 → 关闭写屏障
4. 并发清除

**GC 触发条件：**
- 堆内存达到 GOGC 阈值（默认 100%，即堆翻倍时触发）
- 手动 `runtime.GC()`
- 2 分钟未触发则强制

### 标准答案要点（5星回答）
1. 三色标记的工作流程
2. 混合写屏障解决什么问题（黑色对象引用白色对象）
3. GOGC 参数的含义和调优
4. 逃逸分析的规则和查看方法
5. 有 pprof 性能分析的实际经验

---

## 4. Interface 与反射

### Interface 底层
```go
// 空接口 eface
type eface struct {
    _type *_type  // 类型信息
    data  unsafe.Pointer // 数据指针
}

// 非空接口 iface
type iface struct {
    tab  *itab    // 类型+方法表
    data unsafe.Pointer
}
```

### 类型断言与类型选择
```go
// 类型断言
val, ok := i.(string)

// 类型选择
switch v := i.(type) {
case string: ...
case int: ...
default: ...
}
```

### 反射 (reflect)
- `reflect.TypeOf()` → 获取类型信息
- `reflect.ValueOf()` → 获取值信息
- 三大法则：反射可以从接口值获取反射对象；反射对象可以恢复为接口值；要修改反射对象，值必须可设置（传指针）

### 标准答案要点（5星回答）
1. eface 和 iface 的区别
2. 值接收者 vs 指针接收者对接口实现的影响
3. 反射的三大法则
4. 反射的性能代价和使用场景
5. Go 1.18 泛型对减少反射使用的影响

---

## 5. Go Web 开发（Gin 框架）

### Gin 核心
- **路由**：基于 httprouter（前缀树），高性能
- **中间件**：洋葱模型，`c.Next()` 链式调用
- **参数绑定**：`ShouldBindJSON` / `ShouldBindQuery`
- **分组路由**：`r.Group("/api/v1")`

### 常用中间件实现
```go
// 鉴权中间件
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        claims, err := ParseJWT(token)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
            return
        }
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
```

### GORM ORM
- 自动迁移：`db.AutoMigrate(&User{})`
- 预加载：`db.Preload("Orders").Find(&users)`
- 事务：`db.Transaction(func(tx *gorm.DB) error { ... })`
- Hook：BeforeCreate / AfterCreate / BeforeUpdate

### 标准答案要点（5星回答）
1. Gin 中间件的执行顺序和 Abort 机制
2. GORM 的 Hook 和 Callback 机制
3. 连接池配置（SetMaxIdleConns / SetMaxOpenConns）
4. 优雅关停 (Graceful Shutdown) 的实现
5. 有完整的 Go Web 项目开发经验
