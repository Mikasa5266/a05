# Python 后端 - 语言核心面试知识库

## 1. Python 基础特性

### GIL（全局解释器锁）
**定义**：CPython 中的互斥锁，同一时刻只有一个线程执行 Python 字节码
**影响**：
- CPU 密集型任务：多线程无法利用多核 → 用 multiprocessing 或 C 扩展
- IO 密集型任务：多线程仍有效（IO 等待时释放 GIL）
- 替代方案：asyncio 协程、ProcessPoolExecutor

### 深浅拷贝
- **浅拷贝** `copy.copy()`：只复制一层，内部对象仍是引用
- **深拷贝** `copy.deepcopy()`：递归复制所有层级，完全独立
- **赋值**：仅绑定变量名到同一对象

### 可变与不可变类型
| 不可变 | 可变 |
|--------|------|
| int, float, str, tuple, frozenset | list, dict, set |
- 不可变类型可作为 dict 的 key
- 函数默认参数使用可变类型的陷阱：`def f(lst=[])`

### 标准答案要点（5星回答）
1. 清楚 GIL 的本质和绕过方式
2. 深浅拷贝的区别及嵌套列表场景
3. 理解 Python 的引用语义和对象模型

---

## 2. 装饰器与元编程

### 装饰器原理
```python
# 装饰器本质是高阶函数
def timer(func):
    @functools.wraps(func)  # 保留原函数元信息
    def wrapper(*args, **kwargs):
        start = time.time()
        result = func(*args, **kwargs)
        print(f"{func.__name__} took {time.time()-start:.3f}s")
        return result
    return wrapper

@timer  # 等价于 my_func = timer(my_func)
def my_func(): ...
```

### 类装饰器与带参数装饰器
```python
# 带参数装饰器（三层嵌套）
def retry(max_attempts=3):
    def decorator(func):
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            for i in range(max_attempts):
                try:
                    return func(*args, **kwargs)
                except Exception as e:
                    if i == max_attempts - 1:
                        raise
        return wrapper
    return decorator
```

### 元类 (Metaclass)
- `type` 是所有类的元类
- `__new__` 控制类的创建，`__init__` 控制类的初始化
- 应用：ORM 框架（Django Model）、API 自动注册、单例模式

### 标准答案要点（5星回答）
1. 能手写一个带参数的装饰器
2. 理解 `@functools.wraps` 的作用
3. 知道装饰器的执行时机（模块加载时即执行）
4. 元类的使用场景和 `__new__` vs `__init__`

---

## 3. 生成器与协程

### 生成器 Generator
```python
def fibonacci(n):
    a, b = 0, 1
    for _ in range(n):
        yield a       # yield 暂停执行并返回值
        a, b = b, a + b

# 惰性求值，节省内存
gen = fibonacci(1000000)  # 不会占用大量内存
```

### asyncio 协程
```python
import asyncio

async def fetch_data(url):
    async with aiohttp.ClientSession() as session:
        async with session.get(url) as resp:
            return await resp.json()

# 并发执行多个 IO 任务
results = await asyncio.gather(
    fetch_data(url1),
    fetch_data(url2),
    fetch_data(url3)
)
```

### 标准答案要点（5星回答）
1. yield 的暂停与恢复机制
2. 生成器表达式 vs 列表推导式的内存差异
3. async/await 的事件循环模型
4. asyncio 与多线程在 IO 密集任务中的性能对比

---

## 4. Python Web 框架

### Django vs Flask vs FastAPI
| 维度 | Django | Flask | FastAPI |
|------|--------|-------|---------|
| 定位 | 全栈框架 | 微框架 | 高性能 API |
| ORM | 内置 | SQLAlchemy | SQLAlchemy/Tortoise |
| 异步 | Django 3.1+ 部分支持 | 不原生支持 | 原生 async |
| 性能 | 较低 | 中等 | 接近 Go/Node |
| 自动文档 | 无（需 DRF+Swagger） | 无 | 自带 Swagger/ReDoc |

### Django 核心
- MTV 架构：Model-Template-View
- ORM：`Model.objects.filter()` → SQL
- Middleware：请求生命周期钩子
- Admin：自动后台管理
- Signal：松耦合的事件通知

### FastAPI 核心
- 基于 Pydantic 做数据校验
- 依赖注入系统 Depends()
- ASGI 服务器（Uvicorn/Gunicorn）
- 自动生成 OpenAPI 3.0 文档

### 标准答案要点（5星回答）
1. 根据场景选择合适的框架
2. Django ORM 的 N+1 查询问题及 select_related / prefetch_related
3. FastAPI 的类型提示和自动校验
4. WSGI vs ASGI 的区别
5. 有实际框架使用和部署经验

---

## 5. 数据处理与算法

### 核心库
- **NumPy**：多维数组运算，向量化操作比纯 Python 循环快 100x
- **Pandas**：DataFrame 数据处理，groupby/merge/pivot
- **Scikit-learn**：经典机器学习算法

### 常见面试算法题 Python 解法技巧
1. 列表推导式简化代码
2. `collections.Counter` 统计频率
3. `heapq` 实现 Top-K
4. `bisect` 二分查找
5. `functools.lru_cache` 记忆化搜索

### 标准答案要点（5星回答）
1. NumPy 向量化 vs Python 循环的性能差异
2. Pandas 大数据集优化（chunksize、dtypes 优化）
3. 至少掌握 3 种 Python 独有的算法技巧
