# 前端开发 - 浏览器原理面试知识库

## 1. 浏览器渲染流程

### 关键渲染路径 (Critical Rendering Path)
```
HTML → DOM Tree
              ↘
               Render Tree → Layout → Paint → Composite
              ↗
CSS → CSSOM Tree
```

**详细步骤：**
1. **解析 HTML → DOM Tree**：遇到 `<script>` 阻塞解析（除非 async/defer）
2. **解析 CSS → CSSOM Tree**：CSS 解析不阻塞 DOM 构建，但阻塞渲染
3. **合并 → Render Tree**：只包含可见元素（display:none 不参与）
4. **Layout (回流/重排)**：计算每个元素的位置和大小
5. **Paint (重绘)**：填充像素，生成位图
6. **Composite (合成)**：多个图层合成最终画面（GPU 加速）

### 回流 vs 重绘
| 操作 | 触发回流 | 仅触发重绘 |
|------|---------|-----------|
| width/height 变化 | ✓ | |
| margin/padding 变化 | ✓ | |
| 读取 offsetTop/scrollTop | ✓（强制同步布局） | |
| color/background 变化 | | ✓ |
| visibility: hidden | | ✓ |
| transform/opacity | | ✓（合成层优化） |

### 性能优化策略
1. **减少回流**：批量修改 DOM（DocumentFragment）、避免逐条修改样式
2. **利用合成层**：`transform` + `opacity` 动画走 GPU，不触发 Layout/Paint
3. **will-change**：提前告知浏览器元素将变化
4. **虚拟列表**：只渲染可视区域的 DOM 节点

### 标准答案要点（5星回答）
1. 完整描述从 URL 输入到页面显示的过程
2. 区分回流和重绘，知道哪些操作会触发
3. 理解 Composite 合成层的 GPU 加速原理
4. 有实际的性能优化经验（Lighthouse 分数提升）
5. 了解 Core Web Vitals（LCP / FID / CLS）指标

---

## 2. JavaScript 事件循环

### 浏览器事件循环
```
Call Stack (执行栈)
      ↓ 执行完毕
检查 微任务队列 (Microtask Queue) → 全部执行完
      ↓
取一个 宏任务 (Macrotask Queue) → 执行
      ↓
检查是否需要渲染 → 渲染
      ↓ 循环
```

**宏任务 (Macrotask)**：setTimeout、setInterval、I/O、UI rendering、MessageChannel
**微任务 (Microtask)**：Promise.then、MutationObserver、queueMicrotask、async/await（await 之后的部分）

### 经典面试题
```javascript
console.log('1')
setTimeout(() => console.log('2'), 0)
Promise.resolve().then(() => console.log('3'))
console.log('4')

// 输出：1 4 3 2
// 解析：同步 1→4，微任务 3，宏任务 2
```

### Node.js 事件循环（6 阶段）
```
timers → pending callbacks → idle/prepare → poll → check → close callbacks
```
- poll 阶段：处理 I/O 回调
- check 阶段：setImmediate 回调
- timers 阶段：setTimeout/setInterval 回调

### 标准答案要点（5星回答）
1. 画出事件循环的完整流程图
2. 区分宏任务和微任务的执行时机
3. 能正确分析复杂的异步执行顺序题
4. 浏览器 vs Node.js 事件循环的差异
5. requestAnimationFrame 在事件循环中的位置

---

## 3. HTTP 与网络

### HTTP 缓存
**强缓存：**
- `Cache-Control: max-age=31536000`（秒）
- `Expires: Thu, 01 Jan 2027 00:00:00 GMT`
- 命中强缓存 → 200 (from cache)，不发请求

**协商缓存：**
- `ETag / If-None-Match`：基于内容哈希
- `Last-Modified / If-Modified-Since`：基于修改时间
- 命中协商缓存 → 304 Not Modified

### HTTP/1.1 vs HTTP/2 vs HTTP/3
| 版本 | 关键特性 |
|------|---------|
| HTTP/1.1 | 持久连接、管线化（但有队头阻塞）|
| HTTP/2 | 多路复用、头部压缩（HPACK）、服务器推送 |
| HTTP/3 | 基于 QUIC（UDP），无队头阻塞、0-RTT 握手 |

### HTTPS = HTTP + TLS
**TLS 1.3 握手（1-RTT）：**
1. ClientHello（支持的加密套件 + 密钥份额）
2. ServerHello（选定套件 + 密钥份额 + 证书）
3. 双方计算对称密钥 → 加密通信

### 跨域解决方案
1. **CORS**：服务端 `Access-Control-Allow-Origin`
2. **代理**：Nginx 反向代理 / Vite proxy
3. **JSONP**：利用 script 标签（只支持 GET）
4. **WebSocket**：不受同源策略限制

### 标准答案要点（5星回答）
1. 强缓存和协商缓存的完整流程
2. HTTP/2 多路复用如何解决队头阻塞
3. HTTPS 握手流程
4. CORS 预检请求（OPTIONS）的触发条件
5. 有实际的网络性能优化经验

---

## 4. TypeScript

### 核心类型系统
```typescript
// 联合类型与类型守卫
type Result = Success | Error;
function handle(result: Result) {
    if ('data' in result) { /* Success */ }
}

// 泛型
function identity<T>(arg: T): T { return arg; }

// 工具类型
type Partial<T> = { [P in keyof T]?: T[P] };
type Required<T> = { [P in keyof T]-?: T[P] };
type Pick<T, K extends keyof T> = { [P in K]: T[P] };
type Omit<T, K extends keyof T> = Pick<T, Exclude<keyof T, K>>;

// 条件类型
type IsString<T> = T extends string ? true : false;
```

### 标准答案要点（5星回答）
1. interface vs type 的区别（interface 可合并声明）
2. 泛型约束 `<T extends HasLength>`
3. 内置工具类型的实现原理
4. 类型推断和类型收窄
5. 有 TypeScript 项目的实际开发经验

---

## 5. 工程化

### Webpack vs Vite
| 维度 | Webpack | Vite |
|------|---------|------|
| 开发启动 | 全量打包后启动（慢） | 基于 ESM 按需加载（快） |
| HMR | 重新构建受影响模块 | 精确到模块级别 |
| 打包 | 自身打包 | 生产用 Rollup |
| 生态 | 最成熟 | 快速增长 |

### 性能优化清单
1. **代码分割**：dynamic import、React.lazy、Vue async components
2. **Tree Shaking**：移除未使用代码（需 ESM）
3. **资源优化**：图片懒加载、WebP、CDN
4. **首屏优化**：骨架屏、SSR/SSG、preload/prefetch
5. **运行时优化**：虚拟列表、Web Worker、debounce/throttle
