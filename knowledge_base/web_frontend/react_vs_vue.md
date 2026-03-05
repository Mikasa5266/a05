# 前端开发 - 框架对比面试知识库

## 1. React vs Vue

### 核心差异
| 维度 | React | Vue |
|------|-------|-----|
| 设计理念 | UI = f(state)，函数式倾向 | 渐进式框架，声明式模板 |
| 模板 | JSX（JavaScript 中写 HTML） | SFC 模板（Template + Script + Style） |
| 数据驱动 | 不可变数据 + setState | 可变数据 + 响应式代理 |
| 状态管理 | useState / useReducer + Context / Redux | reactive / ref + Pinia |
| 生态 | 灵活组合（百花齐放） | 官方统一（Vue Router + Pinia） |
| 学习曲线 | 较陡（需理解 Hooks 规则） | 较平缓（模板语法直观） |
| 性能优化 | useMemo / useCallback / React.memo | 编译时优化，自动依赖追踪 |
| 移动端 | React Native | uni-app / Weex |

### React Hooks 核心
```jsx
function Counter() {
  const [count, setCount] = useState(0);
  
  useEffect(() => {
    document.title = `Count: ${count}`;
    return () => { /* cleanup */ };
  }, [count]); // 依赖数组
  
  const doubleCount = useMemo(() => count * 2, [count]);
  const handleClick = useCallback(() => setCount(c => c + 1), []);
  
  return <button onClick={handleClick}>{doubleCount}</button>;
}
```

### Vue 3 Composition API 核心
```vue
<script setup>
import { ref, computed, onMounted, watch } from 'vue'

const count = ref(0)
const doubleCount = computed(() => count.value * 2)

watch(count, (newVal) => {
  document.title = `Count: ${newVal}`
})

onMounted(() => console.log('mounted'))
</script>
```

### 标准答案要点（5星回答）
1. 两者响应式系统的底层差异（React 的 Fiber reconciliation vs Vue 的 Proxy 依赖追踪）
2. React Hooks 的规则（只能在顶层调用、只能在函数组件中使用）
3. Vue 3 的编译时优化（静态提升、patchFlag、Block Tree）
4. 根据项目场景选择框架的考量
5. 有实际项目迁移或技术选型的经验

---

## 2. Vue 3 深入

### 响应式原理（Proxy）
```javascript
// Vue 3 使用 Proxy（而非 Vue 2 的 Object.defineProperty）
const handler = {
  get(target, key, receiver) {
    track(target, key) // 依赖收集
    return Reflect.get(target, key, receiver)
  },
  set(target, key, value, receiver) {
    const result = Reflect.set(target, key, value, receiver)
    trigger(target, key) // 派发更新
    return result
  }
}
const observed = new Proxy(target, handler)
```

**Proxy vs Object.defineProperty：**
- Proxy 可以拦截数组下标修改和 length 变化
- Proxy 可以拦截新增属性（defineProperty 需要 Vue.set）
- Proxy 是惰性代理（只有访问时才代理嵌套对象）

### 虚拟 DOM 与 Diff 算法
1. 同层级比较（不跨层）
2. 相同类型节点 → patch 更新属性
3. 不同类型 → 销毁旧节点，创建新节点
4. 列表 Diff → 使用 key 做最长递增子序列（LIS）优化移动

### 标准答案要点（5星回答）
1. Proxy 响应式的 track/trigger 流程
2. ref vs reactive 的区别和使用场景
3. Vue 3 编译优化（PatchFlags 标记动态节点）
4. nextTick 的实现原理（微任务队列）
5. Teleport / Suspense 等 Vue 3 新特性

---

## 3. React 深入

### Fiber 架构
- 目的：解决 React 15 Stack Reconciler 的同步递归问题
- Fiber 节点：一个 Fiber 对应一个组件，组成链表树
- 双缓冲：current tree + workInProgress tree
- 时间切片：将渲染拆分为小任务，每帧空闲时执行，可中断

### React 18 新特性
- **并发模式 (Concurrent Mode)**：多任务优先级调度
- **useTransition**：标记低优先级更新
- **Suspense**：异步组件加载的声明式写法
- **Automatic Batching**：自动批量更新（promise/setTimeout 中也生效）
- **Server Components**：部分组件在服务端渲染

### 标准答案要点（5星回答）
1. Fiber 的三个阶段：Schedule → Reconcile → Commit
2. 并发模式如何实现可中断渲染
3. useMemo vs useCallback 的使用场景
4. 受控组件 vs 非受控组件
5. Next.js 等 SSR 框架的使用经验
