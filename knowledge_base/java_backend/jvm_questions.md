# Java 后端 - JVM 面试知识库

## 1. JVM 内存模型

### 核心知识点
Java 虚拟机运行时数据区域分为以下几个部分：

**线程私有区域：**
- **程序计数器 (PC Register)**：记录当前线程执行的字节码指令地址，唯一不会 OOM 的区域
- **虚拟机栈 (VM Stack)**：每个方法对应一个栈帧（局部变量表、操作数栈、动态链接、方法返回地址），默认大小 1MB（-Xss），过深递归会 StackOverflowError
- **本地方法栈 (Native Method Stack)**：为 native 方法服务

**线程共享区域：**
- **堆 (Heap)**：对象实例分配的主要区域，GC 管理的核心区域。分为新生代（Eden + S0 + S1）和老年代
- **方法区 (Method Area)**：存储类信息、常量、静态变量。JDK8+ 使用 Metaspace 替代永久代，存放于本地内存
- **运行时常量池**：方法区的一部分，存放编译期生成的字面量和符号引用

### 标准答案要点（5星回答）
1. 能画出完整的 JVM 运行时数据区结构图
2. 清楚每个区域存储的内容和生命周期
3. 知道 JDK7→JDK8 的变化（永久代→元空间）
4. 能说出各区域可能抛出的异常类型
5. 结合实际调优经验说明参数配置（-Xms/-Xmx/-Xss/-XX:MetaspaceSize）

---

## 2. 垃圾回收 (GC)

### 核心知识点

**垃圾判定算法：**
- **引用计数法**：简单高效但无法解决循环引用（Python 用此方案 + 分代回收）
- **可达性分析 (GC Roots Tracing)**：Java 采用此方案。GC Root 包括：虚拟机栈引用、静态变量引用、JNI 引用、活跃线程、同步锁持有的对象

**回收算法：**
- **标记-清除 (Mark-Sweep)**：产生内存碎片
- **标记-整理 (Mark-Compact)**：老年代常用，无碎片但效率较低
- **复制算法 (Copying)**：新生代使用，Eden:S0:S1 = 8:1:1
- **分代收集**：结合以上算法，新生代用复制，老年代用标记-整理/清除

**主流垃圾收集器：**
| 收集器 | 区域 | 算法 | 特点 |
|--------|------|------|------|
| Serial | 新生代 | 复制 | 单线程，Client 模式默认 |
| ParNew | 新生代 | 复制 | Serial 的多线程版本 |
| Parallel Scavenge | 新生代 | 复制 | 吞吐量优先 |
| CMS | 老年代 | 标记-清除 | 低延迟，4 阶段 |
| G1 | 全堆 | Region + 混合 | JDK9+ 默认，可预测停顿 |
| ZGC | 全堆 | 染色指针 | JDK15+ 生产就绪，<10ms 停顿 |

### CMS 四阶段详解
1. **初始标记**（STW）：标记 GC Roots 直接关联的对象，极快
2. **并发标记**：遍历对象图，与用户线程并发
3. **重新标记**（STW）：修正并发标记期间变化的引用
4. **并发清除**：清理垃圾对象，与用户线程并发

### G1 核心机制
- 将堆划分为等大的 Region（1-32MB），每个 Region 可独立充当 Eden/Survivor/Old/Humongous
- 维护 Remembered Set 跟踪跨 Region 引用
- Mixed GC：同时回收新生代和部分老年代 Region
- 通过 -XX:MaxGCPauseMillis 设置期望停顿时间

### 标准答案要点（5星回答）
1. 完整描述可达性分析 + 四种 GC Root
2. 说清各回收算法的优缺点及适用场景
3. 至少深入讲解一个收集器的工作机制（CMS/G1/ZGC）
4. 结合实际 GC 调优经验（日志分析、参数调整）
5. 了解 ZGC/Shenandoah 等最新收集器方向

---

## 3. 类加载机制

### 核心知识点

**类加载过程：** 加载 → 验证 → 准备 → 解析 → 初始化 → 使用 → 卸载

**双亲委派模型 (Parent Delegation Model)：**
- Bootstrap ClassLoader → Extension ClassLoader → Application ClassLoader → Custom ClassLoader
- 工作原理：收到类加载请求时，先委派给父类加载器处理，层层上推。父类加载器无法处理时才由当前加载器自行尝试
- 作用：保证核心类库安全（防止自定义 java.lang.String 替换 JDK 原生类）
- 打破场景：SPI（ServiceLoader）、OSGi、Tomcat 多 WebApp 隔离

### 标准答案要点（5星回答）
1. 7 个阶段能逐一说明每步做什么
2. 双亲委派的链路和目的说清楚
3. 举出 2-3 个打破双亲委派的实际场景
4. 了解类的主动引用 vs 被动引用（何时触发初始化）
5. 对 Java 9 模块化系统对类加载的影响有认知

---

## 4. JVM 调优实战

### 常用参数
```
-Xms256m            # 堆初始大小
-Xmx512m            # 堆最大大小
-Xss256k            # 线程栈大小
-XX:MetaspaceSize=128m    # 初始元空间大小
-XX:MaxMetaspaceSize=256m # 最大元空间大小
-XX:NewRatio=2            # 老年代:新生代 = 2:1
-XX:SurvivorRatio=8       # Eden:S0:S1 = 8:1:1
-XX:+UseG1GC              # 使用 G1 收集器
-XX:MaxGCPauseMillis=200  # G1 目标停顿
-XX:+HeapDumpOnOutOfMemoryError    # OOM 时自动 dump
-XX:HeapDumpPath=/tmp/oom.hprof    # dump 文件路径
```

### 常见问题排查
| 问题 | 工具 | 关键指标 |
|------|------|----------|
| OOM | jmap + MAT | 大对象、泄漏引用链 |
| CPU 飙高 | top + jstack | 长时间运行的线程、死锁 |
| GC 频繁 | jstat -gcutil | GC 次数、停顿时间 |
| 类加载异常 | -verbose:class | 重复加载、类冲突 |

### 标准答案要点（5星回答）
1. 能说出 10+ 个常用 JVM 参数及含义
2. 描述一次完整的 OOM 排查流程
3. 使用过 Arthas/MAT/JVisualVM 等工具
4. 有实际调优经验（如将 GC 停顿从 500ms 降到 50ms）
5. 理解不同场景的最优 GC 选型
