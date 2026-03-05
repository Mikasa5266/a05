# Java 后端 - Spring 框架面试知识库

## 1. IoC 与依赖注入

### 核心知识点
**IoC（控制反转）**：将对象的创建和依赖管理从程序代码转移到 Spring 容器

**DI（依赖注入）方式：**
- 构造器注入（推荐，不可变，利于测试）
- Setter 注入（可选依赖）
- 字段注入 @Autowired（方便但难以测试）

**Bean 生命周期：**
1. 实例化（Constructor）
2. 属性填充（Population）
3. BeanNameAware → BeanFactoryAware → ApplicationContextAware
4. BeanPostProcessor.postProcessBeforeInitialization
5. @PostConstruct → InitializingBean.afterPropertiesSet → init-method
6. BeanPostProcessor.postProcessAfterInitialization（AOP 代理在此产生）
7. 使用中
8. @PreDestroy → DisposableBean.destroy → destroy-method

**Bean 作用域：**
- singleton（默认）：单例，容器启动时创建
- prototype：每次请求创建新实例
- request / session / application：Web 相关

### 标准答案要点（5星回答）
1. 理解 IoC 的本质（好莱坞原则 "Don't call us, we'll call you"）
2. 完整说出 Bean 生命周期 8 个阶段
3. 知道循环依赖的三级缓存解决方案
4. 区分 @Autowired（by type） vs @Resource（by name）
5. 了解 FactoryBean 和 BeanFactory 的区别

---

## 2. Spring AOP

### 核心知识点

**AOP 核心概念：**
- Aspect（切面）：横切关注点的模块化
- Join Point（连接点）：方法执行点
- Pointcut（切点）：匹配连接点的表达式
- Advice（通知）：Before / After / AfterReturning / AfterThrowing / Around
- Weaving（织入）：将切面应用到目标对象

**Spring AOP 实现方式：**
- JDK 动态代理：目标类实现了接口时使用（基于 Proxy + InvocationHandler）
- CGLIB 代理：目标类未实现接口时使用（基于字节码生成子类）
- Spring Boot 2.x 默认使用 CGLIB

**AOP 经典应用场景：**
事务管理、日志记录、权限校验、性能监控、缓存处理

### 标准答案要点（5星回答）
1. 区分 JDK 代理和 CGLIB 的使用条件和实现原理
2. Around 通知的 ProceedingJoinPoint 使用方式
3. 理解 @Transactional 底层是 AOP 实现
4. 知道自调用问题（同类内调用不走代理）的解决方案
5. 了解 AspectJ 编译时织入 vs Spring AOP 运行时织入

---

## 3. Spring Boot 自动配置

### 核心知识点

**@SpringBootApplication 组合注解：**
- @SpringBootConfiguration：标记配置类
- @EnableAutoConfiguration：启用自动配置
- @ComponentScan：组件扫描

**自动配置原理：**
1. @EnableAutoConfiguration 通过 @Import 导入 AutoConfigurationImportSelector
2. 读取 META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports（Spring Boot 3.x）
3. 根据 @ConditionalOnClass / @ConditionalOnMissingBean 等条件注解决定是否生效
4. 用户自定义 Bean 优先（@ConditionalOnMissingBean 保证）

**常用条件注解：**
- @ConditionalOnClass：类路径存在指定类
- @ConditionalOnMissingBean：容器中没有指定 Bean
- @ConditionalOnProperty：配置属性满足条件

### 标准答案要点（5星回答）
1. 完整描述自动配置的加载链路
2. 能手写一个简易的 starter（自定义自动配置类 + spring.factories）
3. 理解 @Conditional 系列注解的作用
4. 知道如何排除某个自动配置：exclude 属性或配置文件
5. Spring Boot 2.7+ 和 3.x 中 imports 文件路径变化

---

## 4. Spring 事务管理

### 核心知识点

**@Transactional 传播行为（7 种）：**
| 传播行为 | 含义 |
|----------|------|
| REQUIRED（默认） | 有事务加入，没有则创建 |
| REQUIRES_NEW | 始终创建新事务，挂起当前事务 |
| NESTED | 在当前事务内创建嵌套事务（Savepoint） |
| SUPPORTS | 有事务加入，没有就非事务运行 |
| NOT_SUPPORTED | 非事务运行，挂起当前事务 |
| MANDATORY | 必须在事务中调用，否则抛异常 |
| NEVER | 非事务运行，有事务就抛异常 |

**事务失效的 8 种场景：**
1. 方法非 public（代理无法拦截）
2. 自调用（同类内调用不走代理）
3. 异常被 catch 吞掉
4. rollbackFor 未配置（默认只回滚 RuntimeException）
5. 数据库不支持事务（MyISAM）
6. 传播行为不当
7. 多线程环境（事务绑定在 ThreadLocal）
8. 未被 Spring 管理的类

### 标准答案要点（5星回答）
1. 说出 7 种传播行为及适用场景
2. 列举至少 5 种事务失效场景
3. 理解事务底层是 AOP + PlatformTransactionManager
4. REQUIRES_NEW vs NESTED 的本质区别
5. 分布式事务的解决方案（Seata / TCC / 本地消息表）

---

## 5. MyBatis 核心原理

### 核心知识点

**架构三层：**
- API 层：SqlSession 接口
- 核心处理层：参数映射 → SQL 解析 → SQL 执行 → 结果映射
- 基础支撑层：连接管理、事务管理、缓存、日志

**MyBatis 缓存：**
- 一级缓存：SqlSession 级别，默认开启，同 Session 同 SQL 直接返回
- 二级缓存：Mapper 级别，需手动开启（@CacheNamespace），跨 Session 共享

**Mapper 代理原理：**
1. 定义 Mapper 接口（无实现类）
2. MyBatis 通过 JDK 动态代理生成接口的代理对象
3. 调用接口方法 → 代理拦截 → 根据方法名定位 MappedStatement → 执行 SQL

**#{} vs ${}：**
- `#{}`：预编译参数绑定，防止 SQL 注入
- `${}`：字符串替换，用于动态表名/列名，存在注入风险

### 标准答案要点（5星回答）
1. 能说出 Mapper 代理的实现原理
2. 区分一级缓存和二级缓存的作用域和使用场景
3. #{} 和 ${} 的区别及安全性考量
4. 了解 MyBatis-Plus 的增强特性
5. 有 XML 和注解两种映射方式的实际使用经验
