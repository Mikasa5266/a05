# 后端工程师 - 框架与中间件

> 格式说明：每道题包含「问题 → 规范答案 → 得分点 → 加分项 → 追问方向」

---

## Q1: Spring Boot 的自动配置原理

**规范答案：**

Spring Boot 自动配置的核心入口是 `@SpringBootApplication` 注解，它包含 `@EnableAutoConfiguration`。

**自动配置链路：**
1. `@EnableAutoConfiguration` 导入 `AutoConfigurationImportSelector`
2. 该 Selector 读取 `META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports` 文件（Spring Boot 3.x）
3. 文件中列出了所有候选的自动配置类
4. 每个配置类通过 `@ConditionalOnXxx` 条件注解判断是否生效

**常用条件注解：**
- `@ConditionalOnClass`：类路径中存在某个类时生效
- `@ConditionalOnMissingBean`：容器中不存在某个 Bean 时生效
- `@ConditionalOnProperty`：配置文件中某个属性为指定值时生效

**执行流程：** 引入 starter 依赖 → 类路径中有对应的类 → 条件注解匹配 → 自动注入默认配置的 Bean → 用户可通过配置文件覆盖默认值。

**得分点：**
- [ ] 知道从 `@SpringBootApplication` 到自动配置类的链路
- [ ] 理解条件注解的过滤机制
- [ ] 知道 starter 的作用（依赖管理 + 自动配置）
- [ ] 理解"约定大于配置"的思想

**加分项：**
- 提到 Spring Boot 2.x 和 3.x 配置文件的变化（spring.factories → imports）
- 介绍如何自定义一个 starter
- 提到 `@AutoConfigureOrder` 和 `@AutoConfigureAfter/Before`

**追问方向：**
- 如何自定义一个 Spring Boot Starter？
- Spring 的 Bean 生命周期是怎样的？
- Spring 是如何解决循环依赖的？

---

## Q2: Spring 的 IoC 和 AOP 原理

**规范答案：**

**IoC（控制反转）：**
- 核心思想：把对象的创建和依赖关系管理交给 Spring 容器，而非程序员手动 new
- 实现方式：DI（依赖注入），通过构造器注入、Setter 注入、字段注入等
- 容器：`BeanFactory`（懒加载） → `ApplicationContext`（预加载 + 增强功能）

**Bean 生命周期（核心阶段）：**
1. 实例化（通过构造方法或工厂方法）
2. 属性填充（依赖注入）
3. Aware 接口回调（BeanNameAware、ApplicationContextAware 等）
4. BeanPostProcessor 前置处理
5. InitializingBean / `@PostConstruct` 初始化
6. BeanPostProcessor 后置处理（AOP 代理在此生成）
7. 使用
8. DisposableBean / `@PreDestroy` 销毁

**AOP（面向切面编程）：**
- 核心概念：切面（Aspect）、切入点（Pointcut）、通知（Advice）、织入（Weaving）
- 实现方式：JDK 动态代理（接口）vs CGLIB 代理（类）
- 五种通知类型：Before、After、AfterReturning、AfterThrowing、Around

**得分点：**
- [ ] 理解 IoC 的本质是依赖关系的反转
- [ ] 描述 Bean 生命周期的核心阶段
- [ ] 知道 AOP 的代理实现方式（JDK vs CGLIB）
- [ ] 了解五种通知类型

**加分项：**
- 解释三级缓存解决循环依赖的原理
- 分析 Spring Boot 默认使用 CGLIB 代理的原因
- 提到 AspectJ 编译时织入 vs Spring AOP 运行时织入

**追问方向：**
- Spring 是如何解决循环依赖的？三级缓存分别存什么？
- JDK 动态代理和 CGLIB 的区别？为什么 Spring Boot 默认用 CGLIB？
- Spring 事务的传播行为有哪些？

---

## Q3: Go 语言的 goroutine 和 GMP 调度模型

**规范答案：**

**Goroutine：**
- Go 语言的轻量级协程，初始栈仅 2KB（可动态扩缩）
- 由 Go runtime 调度，属于用户态线程
- `go func()` 即可创建，成本极低

**GMP 调度模型：**
- **G（Goroutine）**：待执行的 goroutine
- **M（Machine）**：操作系统线程，执行 goroutine 的载体
- **P（Processor）**：逻辑处理器，持有本地 goroutine 队列，连接 G 和 M

**调度流程：**
1. 新创建的 G 优先放入当前 P 的本地队列（最多 256 个）
2. 本地队列满则放入全局队列
3. M 从绑定的 P 的本地队列取 G 执行
4. 本地队列为空时，从全局队列或其他 P 偷取（Work Stealing）

**抢占式调度：**
- Go 1.14+ 引入基于信号的异步抢占，解决了长时间运行的 goroutine 阻塞问题

**得分点：**
- [ ] 知道 G、M、P 分别代表什么
- [ ] 理解 P 的本地队列和全局队列
- [ ] 了解 Work Stealing 机制
- [ ] 知道 goroutine 与线程的区别（栈大小、调度方式）

**加分项：**
- 提到 Go 1.14 的异步抢占式调度
- 解释 `GOMAXPROCS` 的作用
- 分析 M 的阻塞处理（系统调用时 hand off P）

**追问方向：**
- goroutine 泄漏如何检测和避免？
- channel 的底层实现是怎样的？
- context 包是如何实现取消传播的？

---

## Q4: Go 的 Channel 与并发模式

**规范答案：**

**Channel 基础：**
- Go 中 goroutine 之间通信的管道，遵循 CSP（Communicating Sequential Processes）模型
- 无缓冲 channel：发送和接收同步阻塞
- 有缓冲 channel：缓冲区满时发送阻塞，缓冲区空时接收阻塞

**常见并发模式：**

1. **Fan-Out / Fan-In**：多个 goroutine 从同一 channel 读取（扇出），多个 channel 结果汇聚到一个 channel（扇入）
2. **Pipeline**：多个阶段串联，每个阶段通过 channel 传递数据
3. **Worker Pool**：固定数量的 worker goroutine 从 job channel 取任务处理
4. **Select 多路复用**：监听多个 channel，哪个就绪就执行哪个
5. **Context 控制**：通过 `context.WithCancel/Timeout/Deadline` 控制 goroutine 生命周期

**得分点：**
- [ ] 理解有缓冲和无缓冲 channel 的区别
- [ ] 至少说出 2 种并发模式
- [ ] 知道 select 的用法
- [ ] 了解 context 控制取消的方式

**加分项：**
- 提到 channel 的底层实现（环形缓冲区 + 等待队列）
- 分析向已关闭的 channel 发送数据的 panic 问题
- 提到 `sync.Once`、`sync.WaitGroup`、`sync.Pool` 的使用

**追问方向：**
- channel 和 mutex 各适合什么场景？
- 如何优雅地关闭 channel？
- sync.Map 和 map+mutex 的性能差异？

---

## Q5: Gin 框架的核心特性和中间件机制

**规范答案：**

**Gin 框架核心特性：**
- 基于 httprouter 的高性能 HTTP 框架，使用前缀树（Trie）路由
- 支持路由分组、参数绑定、JSON 渲染
- 中间件支持（洋葱模型）

**中间件机制：**
- 中间件函数签名：`func(c *gin.Context)`
- 通过 `c.Next()` 调用下一个中间件，形成洋葱模型
- 通过 `c.Abort()` 终止后续中间件执行
- 全局中间件 `r.Use()` 和路由组中间件 `group.Use()`

**常用内置中间件：**
- `gin.Logger()`：请求日志
- `gin.Recovery()`：panic 恢复
- 自定义：认证、CORS、限流、请求ID

**得分点：**
- [ ] 知道 Gin 的路由实现（前缀树）
- [ ] 理解中间件的洋葱模型执行顺序
- [ ] 知道 `c.Next()` 和 `c.Abort()` 的区别
- [ ] 了解路由分组和中间件的作用域

**加分项：**
- 对比 Gin、Echo、Fiber 等框架
- 提到 Gin 的性能优势（零内存分配路由）
- 分析如何优雅关闭 Gin 服务

**追问方向：**
- 如何实现一个限流中间件？
- GORM 的预加载和懒加载有什么区别？
- Go 的 HTTP 标准库和 Gin 有什么关系？

---

## Q6: Python Web 框架对比（Django / Flask / FastAPI）

**规范答案：**

| 特性 | Django | Flask | FastAPI |
|------|--------|-------|---------|
| 定位 | 全功能框架 | 微框架 | 现代异步框架 |
| 异步支持 | Django 3.1+ ASGI | 需要扩展 | 原生 async/await |
| ORM | Django ORM（内置） | SQLAlchemy（第三方） | SQLAlchemy/Tortoise |
| API 文档 | 需要 DRF + Swagger | Flask-RESTful + Swagger | 自动生成 OpenAPI 文档 |
| 学习曲线 | 中等 | 低 | 低 |
| 性能 | 较低 | 中等 | 高（基于 Starlette） |
| 适用场景 | 大型全栈项目、CMS | 小型项目、快速原型 | API 服务、微服务 |

**选型建议：**
- 需要快速搭建有管理后台的全栈项目 → Django
- 需要灵活轻量，自由选择组件 → Flask
- 需要高性能 API 服务，重视类型安全 → FastAPI

**得分点：**
- [ ] 知道三个框架的核心定位差异
- [ ] 给出合理的选型建议和场景分析
- [ ] 了解异步支持的差异
- [ ] 知道 FastAPI 的类型提示和自动文档特性

**加分项：**
- 有实际项目使用经验
- 提到 WSGI vs ASGI 的区别
- 分析 FastAPI 性能优势的底层原因（Starlette + Pydantic）

**追问方向：**
- Django 的 ORM N+1 查询问题如何解决？
- FastAPI 的依赖注入是怎么实现的？
- Python 的 GIL 对 Web 服务有什么影响？
