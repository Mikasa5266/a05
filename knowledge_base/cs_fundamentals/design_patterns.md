# 计算机基础 - 设计模式面试知识库

## 1. 创建型模式

### 单例模式 (Singleton)
**目的**：确保一个类只有一个实例

**Java 推荐：枚举实现**
```java
public enum Singleton {
    INSTANCE;
    public void doSomething() { ... }
}
```

**Go 实现：sync.Once**
```go
var (
    instance *Singleton
    once     sync.Once
)
func GetInstance() *Singleton {
    once.Do(func() {
        instance = &Singleton{}
    })
    return instance
}
```

### 工厂模式 (Factory)
- **简单工厂**：一个工厂类根据参数创建不同产品
- **工厂方法**：每种产品有对应的工厂类
- **抽象工厂**：创建一系列相关产品族

### 建造者模式 (Builder)
适用场景：参数多且可选的复杂对象构建
```java
User user = User.builder()
    .name("John")
    .email("john@example.com")
    .age(25)
    .build();
```

---

## 2. 结构型模式

### 代理模式 (Proxy)
- **静态代理**：手动编写代理类
- **动态代理**：JDK Proxy（接口）/ CGLIB（子类）
- **应用**：Spring AOP、RPC 远程调用、延迟加载

### 装饰器模式 (Decorator)
- 动态地给对象添加职责
- Java IO：`new BufferedReader(new InputStreamReader(new FileInputStream(file)))`
- 与代理的区别：装饰器增强原有行为，代理控制访问

### 适配器模式 (Adapter)
- 将不兼容的接口转换为目标接口
- 应用：旧系统集成、第三方库适配

---

## 3. 行为型模式

### 观察者模式 (Observer)
- 一对多依赖关系，主题状态变化通知所有观察者
- 应用：事件监听、消息通知、Vue 响应式系统

### 策略模式 (Strategy)
- 定义一系列算法并使其可互换
- 消除大量 if-else / switch-case
- 应用：不同支付方式、不同排序策略

### 模板方法模式 (Template Method)
- 定义算法骨架，子类实现具体步骤
- 应用：Spring 生命周期回调、JdbcTemplate

### 责任链模式 (Chain of Responsibility)
- 多个处理器形成链条，依次处理请求
- 应用：FilterChain、中间件链、审批流程

---

## 4. 面试高频模式总结

| 模式 | 核心场景 | 面试关联 |
|------|---------|---------|
| 单例 | 全局唯一实例 | DCL + volatile |
| 工厂方法 | 解耦创建逻辑 | Spring BeanFactory |
| 代理 | 增强/控制访问 | AOP / RPC |
| 装饰器 | 动态增强 | Java IO |
| 观察者 | 事件驱动 | EventBus / 响应式 |
| 策略 | 算法替换 | 消除 if-else |
| 模板方法 | 流程固定，步骤可变 | Spring 模板类 |

### 标准答案要点（5星回答）
1. 每种模式的意图和适用场景
2. 能画出 UML 类图
3. 结合实际框架源码说明应用
4. 模式之间的区别和联系
5. 遵循 SOLID 原则选择合适的模式
