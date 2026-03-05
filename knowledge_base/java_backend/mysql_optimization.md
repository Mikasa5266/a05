# Java 后端 - MySQL 优化面试知识库

## 1. 索引原理

### 核心知识点

**B+ 树索引结构：**
- 非叶子节点只存索引键，不存数据 → 每页可容纳更多索引项 → 树更矮
- 叶子节点存储数据/主键值，且通过双向链表串联 → 支持范围查询
- InnoDB 聚簇索引：叶子节点存储整行数据；二级索引叶子存主键值（需回表）

**索引类型：**
- 聚簇索引（主键索引）：数据和索引在一起
- 二级索引（辅助索引）：叶子节点存主键 → 需回表查询
- 覆盖索引：查询字段全在索引中 → 无需回表
- 联合索引：遵循最左前缀原则
- 前缀索引：对长字符串取前 N 个字符建索引

**索引失效场景：**
1. 对索引列使用函数或计算：`WHERE YEAR(create_time) = 2024`
2. 隐式类型转换：`WHERE varchar_col = 123`
3. 最左前缀不满足：联合索引 (a, b, c)，查 WHERE b = 1
4. LIKE 以 % 开头：`WHERE name LIKE '%张'`
5. OR 连接非索引列
6. 范围查询后的列不走索引：`a > 1 AND b = 2` 中 b 不走索引

### 标准答案要点（5星回答）
1. 画出 B+ 树结构并解释为什么比 B 树更适合数据库
2. 区分聚簇索引和二级索引的数据存储方式
3. 列举 5+ 种索引失效场景
4. 理解覆盖索引的性能优势和使用方式
5. 有 EXPLAIN 分析 SQL 执行计划的实际经验

---

## 2. SQL 优化

### EXPLAIN 执行计划关键字段
| 字段 | 重要值 |
|------|--------|
| type | system > const > eq_ref > ref > range > index > ALL |
| key | 实际使用的索引 |
| rows | 预估扫描行数 |
| Extra | Using index（覆盖索引）/ Using filesort（需排序）/ Using temporary（临时表）|

### 优化策略
1. **避免 SELECT ***：明确列名，利用覆盖索引
2. **分页优化**：`LIMIT 100000, 10` → 改为 `WHERE id > last_id LIMIT 10`（游标分页）
3. **JOIN 优化**：小表驱动大表，被驱动表关联列加索引
4. **子查询转 JOIN**：避免 MySQL 对子查询物化
5. **批量操作**：`INSERT INTO ... VALUES (), (), ()` 代替循环单条插入
6. **索引下推（ICP）**：MySQL 5.6+ 在索引遍历过程中直接过滤

### 标准答案要点（5星回答）
1. 能看懂 EXPLAIN 输出并指出优化方向
2. 知道深分页的性能问题及解决方案
3. JOIN 的 NLJ/BNL/Hash Join 算法原理
4. 有慢 SQL 排查和优化的实际经验
5. 了解 MySQL 8.0 新特性（窗口函数、CTE、直方图统计）

---

## 3. 事务与锁

### 事务隔离级别
| 隔离级别 | 脏读 | 不可重复读 | 幻读 |
|----------|------|-----------|------|
| READ UNCOMMITTED | ✓ | ✓ | ✓ |
| READ COMMITTED | ✗ | ✓ | ✓ |
| REPEATABLE READ（InnoDB默认）| ✗ | ✗ | ✗（MVCC+间隙锁） |
| SERIALIZABLE | ✗ | ✗ | ✗ |

### MVCC 机制
- 每行数据维护两个隐藏字段：trx_id（最近修改的事务ID）、roll_pointer（undo log 指针）
- Read View 包含：m_ids（活跃事务列表）、min_trx_id、max_trx_id、creator_trx_id
- 可见性判断：trx_id < min_trx_id → 可见；trx_id ∈ m_ids → 不可见（沿 undo log 链找旧版本）

### InnoDB 锁类型
- 行锁：Record Lock（记录锁）、Gap Lock（间隙锁）、Next-Key Lock（临键锁=记录锁+间隙锁）
- 表锁：IS/IX 意向锁、AUTO-INC 锁
- 死锁检测：wait-for graph 等待图检测，回滚代价最小的事务

### 标准答案要点（5星回答）
1. 4 种隔离级别的区别及 RR 如何解决幻读
2. MVCC 的 Read View 可见性判断规则
3. RC vs RR 在 Read View 创建时机的差异
4. 间隙锁和 Next-Key Lock 的加锁范围
5. 死锁排查方法：`SHOW ENGINE INNODB STATUS`

---

## 4. 分库分表

### 核心知识点

**垂直拆分：** 按业务模块拆分（用户库、订单库、商品库）
**水平拆分：** 按分片规则拆分（用户表按 user_id % 16 分到 16 张表）

**分片键选择原则：**
1. 查询频率高的字段
2. 数据分布均匀
3. 尽量避免跨分片查询

**常用中间件：** ShardingSphere、MyCat、Vitess

**分库分表带来的问题：**
- 分布式事务 → Seata / TCC
- 跨分片 JOIN → 冗余数据 / 在应用层组装
- 全局 ID → 雪花算法 / UUID / 号段模式
- 扩容迁移 → 一致性哈希 / 成倍扩容

### 标准答案要点（5星回答）
1. 区分垂直拆分和水平拆分的场景
2. 分片键的选择策略
3. 了解分库分表引入的 4+ 个问题及解决方案
4. 有 ShardingSphere 或类似中间件的使用经验
5. 知道什么情况下不该分库分表（先优化索引和架构）
