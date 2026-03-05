# Java 后端 - 并发编程面试知识库

## 1. 线程基础

### 核心知识点

**创建线程的方式：**
1. 继承 Thread 类
2. 实现 Runnable 接口（推荐，避免单继承限制）
3. 实现 Callable 接口 + FutureTask（有返回值）
4. 线程池 ExecutorService（生产推荐）

**线程状态（6 种）：**
NEW → RUNNABLE → (BLOCKED / WAITING / TIMED_WAITING) → TERMINATED

**关键转换：**
- `synchronized` 争锁失败 → BLOCKED
- `Object.wait()` → WAITING
- `Thread.sleep(n)` → TIMED_WAITING
- `LockSupport.park()` → WAITING

### 标准答案要点（5星回答）
1. 清晰描述 6 种状态及转换条件
2. 区分 BLOCKED vs WAITING（一个是争锁失败被动等待，一个是主动释放锁等通知）
3. 理解 interrupt() 的协作中断机制
4. 知道 Thread.join() 底层是 wait/notify

---

## 2. synchronized 与锁升级

### 核心知识点

**synchronized 使用方式：**
- 修饰实例方法：锁对象是 this
- 修饰静态方法：锁对象是 Class 对象
- 同步代码块：锁对象自定义

**JDK6+ 锁升级过程（对象头 Mark Word 变化）：**
无锁 → 偏向锁 → 轻量级锁 → 重量级锁（不可降级）

| 锁状态 | 触发条件 | 实现机制 |
|--------|----------|----------|
| 偏向锁 | 只有一个线程访问 | 在 Mark Word 记录线程 ID，后续同线程进入无需 CAS |
| 轻量级锁 | 少量线程竞争 | CAS 替换 Mark Word 指向锁记录，失败则自旋 |
| 重量级锁 | 自旋达到阈值 | 升级为 ObjectMonitor，线程挂起（内核态切换） |

### ReentrantLock vs synchronized
| 维度 | synchronized | ReentrantLock |
|------|-------------|---------------|
| 实现 | JVM 层面 | API 层面（AQS） |
| 可中断 | 不可中断 | lockInterruptibly() |
| 公平锁 | 非公平 | 可选公平/非公平 |
| 条件变量 | 单条件 wait/notify | 多条件 Condition |
| 自动释放 | 是（退出同步块） | 否（需 finally 手动释放） |

### 标准答案要点（5星回答）
1. 清楚锁升级的 4 个阶段和 Mark Word 变化
2. 能对比 synchronized 和 ReentrantLock 至少 5 个维度
3. 理解 AQS（AbstractQueuedSynchronizer）的核心原理
4. 知道 JDK15 偏向锁被废弃的原因
5. 实际项目中使用锁的经验和避坑

---

## 3. volatile 关键字

### 核心知识点
- **可见性**：写入 volatile 变量时会将工作内存刷回主内存，读取时强制从主内存加载
- **有序性**：通过内存屏障（Memory Barrier）禁止指令重排
- **不保证原子性**：`volatile int count; count++` 不是线程安全的

**典型应用场景：**
1. DCL 单例模式（Double-Check Locking）中的标志位
2. 状态标记（如 `volatile boolean isRunning`）
3. 轻量级发布机制

**内存屏障实现：**
- StoreStore → volatile 写 → StoreLoad
- LoadLoad → volatile 读 → LoadStore

### 标准答案要点（5星回答）
1. 说清可见性和有序性的保证机制
2. 明确 volatile 不保证原子性（举 count++ 反例）
3. 理解 JMM（Java Memory Model）和 happens-before 规则
4. 能说出 DCL 为什么需要 volatile（防止对象半初始化）

---

## 4. 线程池

### 核心知识点

**ThreadPoolExecutor 7 大参数：**
```java
new ThreadPoolExecutor(
    corePoolSize,      // 核心线程数
    maximumPoolSize,   // 最大线程数
    keepAliveTime,     // 非核心线程空闲存活时间
    timeUnit,          // 时间单位
    workQueue,         // 阻塞队列
    threadFactory,     // 线程工厂
    rejectedHandler    // 拒绝策略
);
```

**任务提交流程：**
1. 核心线程未满 → 创建新核心线程执行
2. 核心线程已满 → 放入阻塞队列
3. 队列已满 → 创建非核心线程（不超过 maximumPoolSize）
4. 全部满 → 触发拒绝策略

**4 种拒绝策略：**
- AbortPolicy（默认，抛异常）
- CallerRunsPolicy（调用者线程执行）
- DiscardPolicy（静默丢弃）
- DiscardOldestPolicy（丢弃队列头部任务）

**为什么阿里规范禁止 Executors 创建线程池？**
- `newFixedThreadPool` 使用 LinkedBlockingQueue（无界），会 OOM
- `newCachedThreadPool` 最大线程数 Integer.MAX_VALUE，会创建大量线程

### 线程池大小设定
- **CPU 密集型**：N + 1（N = CPU 核数）
- **IO 密集型**：2N 或 N * (1 + W/C)，W=等待时间，C=计算时间

### 标准答案要点（5星回答）
1. 7 个参数全部说清含义
2. 任务提交 4 步流程画出来
3. 知道阿里规范禁止 Executors 的原因
4. 能根据场景选择合理的线程池配置
5. 有使用 CompletableFuture 进行异步编排的经验

---

## 5. ConcurrentHashMap

### JDK7 vs JDK8 实现差异

| 维度 | JDK7 | JDK8 |
|------|------|------|
| 数据结构 | Segment[] + HashEntry[] + 链表 | Node[] + 链表 + 红黑树 |
| 锁粒度 | Segment 级别（分段锁） | Node 级别（synchronized + CAS） |
| 并发度 | 默认 16 个 Segment | 数组长度（更细粒度） |
| 链表→树 | 不支持 | 链表长度 > 8 且数组 > 64 时转红黑树 |

### JDK8 put 操作流程
1. 计算 hash → 定位桶位置
2. 桶为空 → CAS 写入
3. 桶非空 → synchronized 锁住头节点
4. 链表插入 → 尾插法（JDK7 是头插法）
5. 链表长度 > 8 → 尝试树化
6. 扩容检查 → helpTransfer 协助扩容

### 标准答案要点（5星回答）
1. 对比 JDK7/8 的结构差异
2. 能描述 put 操作的完整流程
3. 理解 CAS + synchronized 协作的精妙设计
4. 知道扩容期间的 ForwardingNode 机制
5. 能和 HashMap、Hashtable 做对比

---

## 6. AQS (AbstractQueuedSynchronizer)

### 核心知识点
AQS 是 JUC 并发工具的基石，ReentrantLock、Semaphore、CountDownLatch 底层都基于 AQS。

**核心思想：**
- 维护一个 volatile int state 表示同步状态
- 维护一个 FIFO 双向等待队列（CLH 队列变种）
- 子类通过 tryAcquire/tryRelease（独占）或 tryAcquireShared/tryReleaseShared（共享）实现同步逻辑

**独占模式（ReentrantLock）：**
1. tryAcquire 尝试获取 → CAS 修改 state
2. 失败 → 封装为 Node 加入等待队列 → park 挂起
3. 前驱释放锁 → unpark 唤醒 → 重新 tryAcquire

### 标准答案要点（5星回答）
1. 说清 state + CLH 队列的核心设计
2. 能描述 ReentrantLock.lock() 的完整调用链
3. 区分独占模式和共享模式
4. 知道公平锁 vs 非公平锁在 AQS 中的实现差异
5. 了解 Condition 的 await/signal 底层实现
