# 测试工程师 - 测试方法与工具面试知识库

## 1. 测试基础理论

### 测试类型金字塔
```
        /  E2E 测试  \       ← 少量，验证关键用户路径
       / 集成测试      \      ← 适量，验证模块协作
      / 单元测试         \    ← 大量，验证函数/方法
```

### 测试分类
| 类型 | 目的 | 工具示例 |
|------|------|----------|
| 单元测试 | 验证最小可测单元 | JUnit, pytest, Go testing |
| 集成测试 | 验证模块间协作 | Spring Boot Test, Testcontainers |
| E2E 测试 | 模拟真实用户操作 | Selenium, Cypress, Playwright |
| 性能测试 | 验证系统承载能力 | JMeter, K6, Locust |
| 安全测试 | 发现安全漏洞 | OWASP ZAP, Burp Suite |
| 兼容性测试 | 跨浏览器/设备/OS | BrowserStack, LambdaTest |

### 测试设计方法
1. **等价类划分**：将输入分为有效/无效等价类，每类取代表值
2. **边界值分析**：取边界值及其邻近值（min-1, min, min+1, max-1, max, max+1）
3. **判定表**：多条件组合的完整覆盖
4. **状态转换法**：订单状态流转测试
5. **探索性测试**：基于经验和直觉的自由测试

### 标准答案要点（5星回答）
1. 清楚测试金字塔和各层级的投入比例
2. 至少说出 3 种测试设计方法并举例
3. 理解测试左移（Shift-Left）和测试右移的理念
4. 有实际测试用例设计和执行经验
5. 了解 TDD / BDD 的概念和适用场景

---

## 2. 自动化测试

### Selenium 核心
```python
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

driver = webdriver.Chrome()
driver.get("https://example.com")

# 显式等待
element = WebDriverWait(driver, 10).until(
    EC.presence_of_element_located((By.ID, "login"))
)
element.click()
```

### Playwright（新一代首选）
```python
from playwright.sync_api import sync_playwright

with sync_playwright() as p:
    browser = p.chromium.launch()
    page = browser.new_page()
    page.goto("https://example.com")
    page.click("#login")
    page.fill("#username", "test")
    assert page.title() == "Dashboard"
```

### API 自动化测试
```python
import requests

def test_create_user():
    resp = requests.post("/api/users", json={"name": "test"})
    assert resp.status_code == 201
    assert resp.json()["name"] == "test"
    
    # 验证数据库
    resp = requests.get(f"/api/users/{resp.json()['id']}")
    assert resp.status_code == 200
```

### 标准答案要点（5星回答）
1. Page Object 设计模式
2. 元素定位策略（CSS Selector / XPath / data-testid）
3. 等待机制（显式等待/隐式等待/Fluent Wait）
4. CI/CD 中自动化测试的集成方式
5. 测试报告生成（Allure）

---

## 3. 性能测试

### 关键指标
| 指标 | 定义 | 标准 |
|------|------|------|
| TPS/QPS | 每秒事务/查询数 | 业务相关 |
| RT | 响应时间 | P99 < 500ms |
| 并发数 | 同时在线用户数 | 业务相关 |
| 错误率 | 失败请求占比 | < 0.1% |
| 吞吐量 | 单位时间处理数据量 | 随并发增长 |

### JMeter 核心组件
- Thread Group：线程组（模拟并发用户）
- HTTP Sampler：HTTP 请求
- Assertion：断言验证
- Listener：结果监听器
- Timer：思考时间
- Logic Controller：逻辑控制器（If/Loop/Random）

### 性能测试类型
1. **负载测试**：正常→峰值，找到性能拐点
2. **压力测试**：超出峰值，验证系统稳定性和恢复能力
3. **稳定性测试**：长时间运行，发现内存泄漏等问题
4. **基准测试**：建立性能基准线

### 标准答案要点（5星回答）
1. 关键性能指标的定义和标准
2. 有 JMeter/Locust/K6 实际使用经验
3. 能设计完整的性能测试方案
4. 性能瓶颈分析方法（CPU/内存/IO/网络）
5. 了解 APM 工具（SkyWalking/Prometheus）

---

## 4. Bug 管理与质量度量

### Bug 严重等级
| 等级 | 定义 | 示例 |
|------|------|------|
| P0-致命 | 系统崩溃/数据丢失 | 支付扣款但订单未创建 |
| P1-严重 | 核心功能不可用 | 登录页面白屏 |
| P2-一般 | 功能异常但有绕过方案 | 搜索结果排序错误 |
| P3-轻微 | UI 问题/文案错误 | 按钮文字溢出 |

### 质量度量指标
- 缺陷密度：缺陷数 / 千行代码
- 测试覆盖率：已执行用例 / 总用例
- 代码覆盖率：被测试覆盖的代码行 / 总代码行
- 缺陷逃逸率：线上缺陷 / 总缺陷
- MTTR：平均修复时间

### 标准答案要点（5星回答）
1. Bug 的完整生命周期（新建→确认→修复→验证→关闭）
2. Bug 报告的规范写法（环境/步骤/期望/实际/截图）
3. 质量度量指标的定义和意义
4. 有使用 Jira/禅道 等管理工具的经验
5. 了解质量保障体系的建设
