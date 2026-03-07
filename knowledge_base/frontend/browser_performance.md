# 前端工程师 - 浏览器原理与性能优化

> 格式说明：每道题包含「问题 → 规范答案 → 得分点 → 加分项 → 追问方向」

---

## Q1: 浏览器渲染流程是怎样的？

**规范答案：**

**完整渲染流水线：**
1. **解析 HTML** → 构建 DOM Tree
2. **解析 CSS** → 构建 CSSOM Tree
3. **合并** DOM + CSSOM → 生成 Render Tree（不包含 `display: none` 的元素）
4. **Layout（布局/回流）** → 计算每个节点的几何信息（位置、大小）
5. **Paint（绘制）** → 将节点转换为绘制指令
6. **Composite（合成）** → GPU 合成各图层，输出到屏幕

**关键概念：**
- **回流（Reflow）**：元素几何属性变化（大小、位置、显隐）→ 重新计算布局
- **重绘（Repaint）**：元素外观变化（颜色、阴影）→ 重新绘制，不影响布局
- 回流必定触发重绘，重绘不一定触发回流

**减少回流的方法：**
- 批量修改样式（cssText 或 class 切换）
- 使用 `transform` 代替 `top/left`（走合成层，不触发回流）
- 使用 `documentFragment` 批量操作 DOM
- 使用 `will-change` 提升合成层

**得分点：**
- [ ] 完整描述渲染流水线的关键步骤
- [ ] 正确区分回流和重绘
- [ ] 知道哪些操作触发回流
- [ ] 给出减少回流的优化方法

**加分项：**
- 提到合成层和 GPU 加速的原理
- 提到 `requestAnimationFrame` 的作用
- 分析 CSS 放头部、JS 放底部的原因

**追问方向：**
- script 标签的 async 和 defer 有什么区别？
- 首屏性能优化有哪些手段？
- 什么是 CLS、LCP、FID 等核心 Web Vitals 指标？

---

## Q2: JavaScript 事件循环（Event Loop）

**规范答案：**

**核心机制：**
JavaScript 是单线程语言，通过事件循环实现异步非阻塞。

**执行顺序：**
1. 执行同步代码（Call Stack）
2. 清空所有微任务队列（Microtask Queue）
3. 取一个宏任务执行
4. 再次清空微任务
5. 循环以上步骤

**宏任务 vs 微任务：**

| 宏任务（Macrotask） | 微任务（Microtask） |
|-------|-------|
| setTimeout / setInterval | Promise.then/catch/finally |
| setImmediate（Node） | MutationObserver |
| I/O、UI 渲染 | queueMicrotask |
| requestAnimationFrame* | process.nextTick（Node，优先级最高） |

**经典面试题输出顺序：**
```javascript
console.log('1');
setTimeout(() => console.log('2'), 0);
Promise.resolve().then(() => console.log('3'));
console.log('4');
// 输出：1 → 4 → 3 → 2
```

**得分点：**
- [ ] 理解单线程 + 事件循环的工作模型
- [ ] 正确区分宏任务和微任务
- [ ] 知道微任务优先于宏任务执行
- [ ] 能分析嵌套异步代码的执行顺序

**加分项：**
- 分析 Node.js 事件循环的六个阶段
- 提到 `requestAnimationFrame` 的执行时机
- 分析 `async/await` 的执行顺序转换

**追问方向：**
- async/await 的本质是什么？如何转换为 Promise？
- Node.js 的事件循环和浏览器有什么区别？
- 为什么 Vue 的 nextTick 要优先使用 Promise？

---

## Q3: HTTP 缓存机制

**规范答案：**

**强缓存（不发请求直接用缓存）：**
- `Cache-Control: max-age=3600`（优先级高）
- `Expires: Thu, 01 Dec 2025 16:00:00 GMT`（绝对时间，有时区问题）
- 命中返回 200（from cache）

**协商缓存（发请求验证是否可用缓存）：**
- `Last-Modified` + `If-Modified-Since`：基于修改时间，精度1秒
- `ETag` + `If-None-Match`：基于内容哈希，精度更高（优先级高）
- 命中返回 304 Not Modified

**缓存策略实践：**
- HTML 文件：`no-cache`（每次协商）
- CSS/JS（带 hash 文件名）：`max-age=31536000`（长期缓存）
- API 接口：通常 `no-store`（不缓存）

**得分点：**
- [ ] 区分强缓存和协商缓存
- [ ] 知道两组请求/响应头的对应关系
- [ ] 理解 Cache-Control 常用值（no-cache、no-store、max-age）
- [ ] 了解实际项目中的缓存策略

**加分项：**
- 提到 Service Worker 缓存
- 分析 `no-cache` 和 `no-store` 的区别
- 提到 CDN 缓存策略

**追问方向：**
- Service Worker 如何实现离线缓存？
- CDN 的工作原理是什么？
- 如何实现资源的增量更新？

---

## Q4: 跨域问题及解决方案

**规范答案：**

**同源策略：**
浏览器的安全策略，要求协议、域名、端口三者都相同才算同源。非同源的请求会被浏览器拦截。

**解决方案：**

| 方案 | 原理 | 适用场景 |
|------|------|---------|
| **CORS** | 服务端设置 `Access-Control-Allow-Origin` 等响应头 | 最常用、最标准 |
| **代理服务器** | 开发环境配置 devServer proxy，线上用 Nginx 反向代理 | 开发/生产通用 |
| **JSONP** | 利用 `<script>` 标签无跨域限制，只支持 GET | 兼容老浏览器 |
| **postMessage** | 窗口间通信 | iframe 跨域通信 |

**CORS 详细流程：**
- **简单请求**：直接发送，浏览器自动加 Origin 头
- **预检请求（Preflight）**：非简单请求先发 OPTIONS 请求，服务端确认后才发正式请求
- 预检条件：自定义 Header、PUT/DELETE 方法、Content-Type 为 application/json 等

**得分点：**
- [ ] 理解同源策略的三要素
- [ ] 至少知道 CORS 和代理两种方案
- [ ] 了解简单请求和预检请求的区别
- [ ] 知道 CORS 涉及的响应头

**加分项：**
- 能配置完整的 CORS 响应头（Allow-Methods、Allow-Headers、Max-Age）
- 分析 Cookie 跨域传递需要的配置（withCredentials）
- 提到 WebSocket 不受同源策略限制

**追问方向：**
- 什么是 XSS 和 CSRF 攻击？如何防御？
- Nginx 如何配置反向代理实现跨域？
- 为什么 CORS 需要预检请求？

---

## Q5: 前端工程化（Webpack vs Vite）

**规范答案：**

| 特性 | Webpack | Vite |
|------|---------|------|
| 构建方式 | 先打包再启动（Bundle-based） | 开发时按需编译（Native ESM） |
| 开发启动速度 | 慢（全量编译） | 极快（只编译当前页面依赖） |
| HMR 速度 | 慢（全量重建） | 快（精确更新修改的模块） |
| 生产构建 | 成熟稳定 | 使用 Rollup 打包 |
| 生态 | 极其丰富 | 快速增长 |
| 配置复杂度 | 高 | 低（开箱即用） |

**Webpack 核心概念：**
- Entry → Loader → Plugin → Output
- Loader：处理非 JS 文件（babel-loader、css-loader）
- Plugin：扩展功能（HtmlWebpackPlugin、MiniCssExtractPlugin）

**Vite 核心原理：**
- 开发环境：利用浏览器原生 ESM（`import`），按需编译，无需打包
- 预构建：使用 esbuild 预构建 node_modules 中的依赖（CJS → ESM）
- 生产构建：使用 Rollup 进行 Tree Shaking 和代码分割

**得分点：**
- [ ] 知道 Webpack 和 Vite 的核心差异
- [ ] 理解 Vite 开发快的原因（Native ESM）
- [ ] 了解 Webpack 的 Loader 和 Plugin 的区别
- [ ] 知道 Tree Shaking 的概念

**加分项：**
- 有 Webpack 性能优化经验（代码分割、缓存、DLL）
- 了解 esbuild 为什么快（Go 编写、多线程）
- 提到 Module Federation（微前端）

**追问方向：**
- Webpack 的热更新（HMR）原理是什么？
- 如何优化 Webpack 的打包速度？
- Tree Shaking 的原理和限制是什么？

---

## Q6: 前端安全（XSS 和 CSRF）

**规范答案：**

**XSS（跨站脚本攻击）：**
- **存储型**：恶意脚本存到服务器数据库，其他用户访问时执行
- **反射型**：恶意脚本通过 URL 参数传入，服务端直接返回渲染
- **DOM 型**：纯前端 JavaScript 操作 DOM 时注入

**XSS 防御：**
- 输入过滤 + 输出转义（HTML 实体编码）
- 使用 CSP（Content Security Policy）限制脚本执行来源
- Cookie 设置 HttpOnly（防止 JS 读取）
- 使用框架自带的模板引擎（Vue/React 默认转义）

**CSRF（跨站请求伪造）：**
- 攻击者诱导用户在已登录状态下访问恶意页面，该页面携带用户 Cookie 发送请求

**CSRF 防御：**
- 使用 CSRF Token（每次请求携带随机 Token）
- 验证 Referer / Origin 头
- Cookie 设置 SameSite 属性（Lax / Strict）
- 关键操作使用二次验证

**得分点：**
- [ ] 正确区分 XSS 和 CSRF 的攻击原理
- [ ] XSS 的三种类型至少知道两种
- [ ] 每种攻击给出至少 2 种防御方法
- [ ] 知道 HttpOnly 和 SameSite 的作用

**加分项：**
- 提到 CSP 头的配置
- 分析 JWT 方案天然防 CSRF 的原因
- 提到点击劫持（Clickjacking）和 X-Frame-Options

**追问方向：**
- 你的项目中是如何做安全防护的？
- 什么是中间人攻击？如何防止？
- 前端如何做敏感数据加密？
