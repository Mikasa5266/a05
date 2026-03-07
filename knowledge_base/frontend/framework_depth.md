# 前端工程师 - 框架深度（Vue & React）

> 格式说明：每道题包含「问题 → 规范答案 → 得分点 → 加分项 → 追问方向」

---

## Q1: Vue 3 的响应式原理是什么？

**规范答案：**

Vue 3 使用 **Proxy** 替代 Vue 2 的 `Object.defineProperty` 实现响应式。

**核心机制：**
1. `reactive()` 内部使用 `new Proxy(target, handler)` 创建代理对象
2. **依赖收集（track）**：访问属性时（get 拦截），将当前副作用函数（effect）收集到该属性的依赖集合中
3. **派发更新（trigger）**：修改属性时（set 拦截），触发该属性所有依赖的副作用函数重新执行

**数据结构：**
```
targetMap: WeakMap<target, Map<key, Set<effect>>>
```
- WeakMap 以原始对象为 key
- Map 以属性名为 key
- Set 存储该属性的所有 effect

**ref vs reactive：**
- `ref()`：包装基本类型，通过 `.value` 访问，内部使用 getter/setter
- `reactive()`：处理对象类型，返回 Proxy 代理

**Vue 3 Proxy 的优势：**
- 能监听新增/删除属性（Vue 2 不行，需要 `$set`）
- 能监听数组下标和 length 变化
- 性能更好（懒代理，访问时才递归代理深层对象）

**得分点：**
- [ ] 知道 Vue 3 使用 Proxy 实现
- [ ] 理解 track（依赖收集）和 trigger（派发更新）的流程
- [ ] 能区分 ref 和 reactive 的使用场景
- [ ] 知道 Proxy 相比 Object.defineProperty 的优势

**加分项：**
- 能画出 targetMap → depsMap → dep（Set）的数据结构
- 提到 `effect`、`computed`、`watch` 的关系
- 分析 shallowReactive 和 shallowRef 的使用场景

**追问方向：**
- Vue 3 的 computed 是如何实现惰性求值的？
- Vue 3 的 watch 和 watchEffect 有什么区别？
- Vue 2 的 Object.defineProperty 有哪些局限性？

---

## Q2: Virtual DOM 和 Diff 算法

**规范答案：**

**Virtual DOM（虚拟 DOM）：**
- 用 JavaScript 对象描述真实 DOM 结构，包含 tag、props、children 等字段
- 作用：减少直接操作真实 DOM 的次数，通过 diff 计算最小更新量

**Diff 算法（Vue 3）：**
1. **同层比较**：只比较同一层级的节点，不跨层移动
2. **类型判断**：tag 不同则直接替换，不深入比较
3. **key 的作用**：通过 key 标识节点身份，提高复用率
4. **最长递增子序列（LIS）**：Vue 3 的优化，找出不需要移动的最长稳定序列，最小化 DOM 操作

**Vue 3 Diff 优化（与 Vue 2 的区别）：**
- 静态提升（Static Hoisting）：静态节点只创建一次
- Patch Flag：标记动态绑定的类型，diff 时跳过静态内容
- 缓存事件处理函数
- Block Tree：将动态节点收集为扁平数组直接 diff

**得分点：**
- [ ] 理解虚拟 DOM 的本质和存在意义
- [ ] 知道 diff 是同层比较
- [ ] 理解 key 的作用（不要说"提高性能"，要说明白原理）
- [ ] 了解 Vue 3 的编译时优化

**加分项：**
- 能解释最长递增子序列算法在 Diff 中的应用
- 对比 Vue 和 React 的 Diff 策略差异
- 提到编译器优化带来的性能收益

**追问方向：**
- 为什么不推荐用 index 作为 key？
- React 的 Fiber 架构是什么？解决了什么问题？
- 直接操作 DOM vs Virtual DOM，哪个更快？

---

## Q3: React Hooks 原理和常见 Hooks

**规范答案：**

**Hooks 的本质：**
- 让函数组件也能使用状态和生命周期特性
- 底层通过链表结构存储，每次渲染按顺序取值

**常见 Hooks：**
| Hook | 作用 | 使用场景 |
|------|------|---------|
| useState | 状态管理 | 组件内部状态 |
| useEffect | 副作用处理 | 数据请求、订阅、DOM 操作 |
| useContext | 跨组件通信 | 全局状态（主题、语言） |
| useMemo | 缓存计算结果 | 昂贵计算避免重复执行 |
| useCallback | 缓存函数引用 | 避免子组件不必要的重渲染 |
| useRef | 引用值/DOM | 操作 DOM、存储不引起渲染的值 |
| useReducer | 复杂状态逻辑 | 多状态关联更新的场景 |

**Hooks 规则：**
1. 只能在函数组件或自定义 Hook 的最顶层调用
2. 不能在条件语句、循环中调用（因为链表顺序要稳定）

**为什么有这些规则：**
React 通过调用顺序（链表）来匹配 state 和 hook 的对应关系。条件语句会打乱顺序，导致 state 错乱。

**得分点：**
- [ ] 至少了解 useState、useEffect、useMemo/useCallback
- [ ] 知道 Hooks 的两条规则和原因
- [ ] 理解 useEffect 的依赖数组机制
- [ ] 知道 useMemo 和 useCallback 的区别

**加分项：**
- 能自定义一个实用 Hook（如 useDebounce、useFetch）
- 理解 React Fiber 中 Hooks 链表的存储结构
- 分析 useEffect 的清理函数的执行时机

**追问方向：**
- useEffect 和 useLayoutEffect 的区别？
- 如何避免 useEffect 的无限循环？
- React 的状态更新是同步还是异步的？

---

## Q4: 前端状态管理方案对比

**规范答案：**

**Vue 生态：**
- **Vuex**：集中式状态管理，State/Getters/Mutations/Actions/Modules
- **Pinia**：Vue 3 推荐方案，更简洁，支持 Composition API，去掉 Mutations

**React 生态：**
- **Redux**：单向数据流，Action → Reducer → Store，通过 middleware 处理异步
- **Redux Toolkit（RTK）**：Redux 官方推荐简化方案
- **Zustand**：轻量级方案，API 简洁
- **Recoil / Jotai**：原子化状态管理

**选型原则：**
- 小型项目：组件内状态 + Context / provide/inject 即可
- 中型项目：Pinia / Zustand
- 大型项目：Pinia / Redux Toolkit

**得分点：**
- [ ] 至少了解一种 Vue 和一种 React 的状态管理方案
- [ ] 理解状态管理解决的核心问题（跨组件共享状态）
- [ ] 知道 Vuex 和 Pinia 的区别
- [ ] 给出合理的选型建议

**加分项：**
- 分析 Redux 的中间件机制（如 redux-thunk、redux-saga）
- 提到 Pinia 的 SSR 支持
- 讨论原子化状态管理的优势

**追问方向：**
- Vuex 的 Mutation 为什么必须是同步的？
- Redux 的中间件原理是什么？
- 如何处理全局状态和局部状态的划分？

---

## Q5: 前端路由原理（Hash 模式 vs History 模式）

**规范答案：**

**Hash 模式：**
- URL 格式：`http://example.com/#/path`
- 原理：监听 `hashchange` 事件，`#` 后面的内容不会发送给服务器
- 优点：兼容性好，不需要服务器配置
- 缺点：URL 不美观，SEO 不友好

**History 模式：**
- URL 格式：`http://example.com/path`
- 原理：使用 HTML5 的 `pushState()` 和 `replaceState()` API 操作浏览器历史记录
- 监听 `popstate` 事件处理前进/后退
- 优点：URL 美观
- 缺点：需要服务器配置 fallback（所有路由返回 index.html），否则刷新 404

**Vue Router 的导航守卫：**
- 全局守卫：`beforeEach`、`afterEach`
- 路由独享守卫：`beforeEnter`
- 组件内守卫：`beforeRouteEnter`、`beforeRouteUpdate`、`beforeRouteLeave`

**得分点：**
- [ ] 清楚 Hash 和 History 模式的原理差异
- [ ] 知道 History 模式需要服务器配置
- [ ] 了解导航守卫的分类和执行顺序
- [ ] 理解 SPA 路由和传统路由的区别

**加分项：**
- 分析 React Router v6 的新特性
- 提到路由懒加载的实现方式
- 提到 SSR 场景下的路由处理

**追问方向：**
- 如何实现路由懒加载？原理是什么？
- 前端路由如何处理权限控制？
- SPA 的 SEO 问题如何解决？

---

## Q6: 组件设计原则与通信方式

**规范答案：**

**组件设计原则：**
1. **单一职责**：每个组件只负责一个功能
2. **可复用性**：通过 props 和 slot 增强灵活性
3. **高内聚低耦合**：内部逻辑独立，通过明确的接口通信

**Vue 组件通信方式：**
| 方式 | 适用场景 | 方向 |
|------|---------|------|
| props / emit | 父子通信 | 父→子 / 子→父 |
| provide / inject | 跨层级通信 | 祖先→后代 |
| EventBus（mitt） | 任意组件通信 | 任意 |
| Pinia / Vuex | 全局状态 | 任意 |
| ref / expose | 父调子方法 | 父→子 |

**React 组件通信方式：**
| 方式 | 适用场景 |
|------|---------|
| props / callback | 父子通信 |
| Context API | 跨层级通信 |
| State Management | 全局通信 |
| Ref | 父调子方法 |

**得分点：**
- [ ] 至少列出 3 种组件通信方式
- [ ] 知道不同通信方式的适用场景
- [ ] 理解 props 单向数据流的意义
- [ ] 了解组件设计的基本原则

**加分项：**
- 提到 Compound Component 模式
- 分析 render props 和 HOC 的区别
- 提到 Vue 3 的 Teleport 和 Suspense

**追问方向：**
- 如何设计一个高复用性的组件库？
- Vue 3 的 Composition API 相比 Options API 有什么优势？
- React 的 HOC 和 Hooks 哪个更好？为什么？
