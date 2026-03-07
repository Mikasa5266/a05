# 前端工程师 - CSS 布局与 TypeScript

> 格式说明：每道题包含「问题 → 规范答案 → 得分点 → 加分项 → 追问方向」

---

## Q1: CSS 盒模型及 box-sizing 的区别

**规范答案：**

**标准盒模型（content-box，默认）：**
- `width/height` 只包含内容区
- 实际宽度 = width + padding + border + margin

**IE 盒模型（border-box）：**
- `width/height` 包含内容 + padding + border
- 实际宽度 = width + margin

**现代开发建议：**
```css
*, *::before, *::after {
  box-sizing: border-box;
}
```
统一使用 `border-box`，更符合直觉，避免计算误差。

**得分点：**
- [ ] 正确区分 content-box 和 border-box
- [ ] 知道各自的宽度计算公式
- [ ] 了解现代项目中统一使用 border-box 的最佳实践

**加分项：**
- 提到 `getComputedStyle` 获取实际计算值
- 提到 `outline` 不占空间不参与盒模型

**追问方向：**
- BFC 是什么？如何触发？
- margin 合并的规则是什么？

---

## Q2: Flexbox 和 Grid 布局

**规范答案：**

**Flexbox（弹性盒布局）— 一维布局：**
- 容器属性：`display: flex`、`flex-direction`、`justify-content`、`align-items`、`flex-wrap`
- 子项属性：`flex: grow shrink basis`、`align-self`、`order`
- `flex: 1` 等同于 `flex: 1 1 0%`，表示等分剩余空间

**Grid（网格布局）— 二维布局：**
- 容器属性：`display: grid`、`grid-template-columns/rows`、`gap`、`grid-template-areas`
- 子项属性：`grid-column`、`grid-row`、`grid-area`
- 强大的行列定义：`repeat(3, 1fr)`、`minmax(200px, 1fr)`

**选择建议：**
- 一维排列（导航栏、工具栏）→ Flexbox
- 二维网格（页面整体布局、卡片列表）→ Grid
- 实际项目中两者结合使用

**得分点：**
- [ ] 知道 Flex 和 Grid 的核心区别（一维 vs 二维）
- [ ] 列出 Flexbox 主要属性并理解含义
- [ ] 知道 Grid 的基本使用方式
- [ ] 给出合理的使用场景选择

**加分项：**
- 能用 Flexbox 实现经典布局（圣杯、双飞翼）
- 能用 Grid 的 `grid-template-areas` 做语义化布局
- 提到 `fr` 单位的含义

**追问方向：**
- 如何实现一个自适应瀑布流布局？
- flex-grow、flex-shrink、flex-basis 各自的作用？
- 如何做移动端响应式布局？

---

## Q3: BFC 是什么？有哪些应用场景？

**规范答案：**

**BFC（Block Formatting Context，块级格式化上下文）：**
一个独立的渲染区域，内部元素的布局不影响外部。

**触发 BFC 的条件：**
- `overflow: hidden/auto/scroll`（非 visible）
- `display: flex/inline-block/grid/table-cell`
- `position: absolute/fixed`
- `float: left/right`（非 none）

**BFC 的应用场景：**
1. **清除浮动**：父元素设置 `overflow: hidden` 触发 BFC，包裹浮动子元素
2. **防止 margin 合并**：将元素包在 BFC 容器中，阻止相邻 margin 折叠
3. **自适应两栏布局**：一侧浮动，另一侧触发 BFC，不被浮动元素覆盖

**得分点：**
- [ ] 知道 BFC 的定义
- [ ] 至少列出 3 种触发方式
- [ ] 至少给出 2 个应用场景
- [ ] 理解 BFC 的隔离性原理

**加分项：**
- 提到 `display: flow-root`（专门触发 BFC 的现代方案）
- 对比 IFC（Inline Formatting Context）
- 结合实际项目举例

**追问方向：**
- 还有哪些 Formatting Context？
- 清除浮动有哪些方式？各自优缺点？
- 包含块（Containing Block）是什么？

---

## Q4: TypeScript 的核心特性和常见类型

**规范答案：**

**核心特性：**
- 静态类型检查：在编译时发现类型错误，减少运行时 bug
- 类型推断：大部分情况不需要显式标注类型
- 渐进式采用：可以逐步将 JS 项目迁移到 TS

**常用类型：**
- 基础类型：`string`、`number`、`boolean`、`null`、`undefined`
- 联合类型：`string | number`
- 交叉类型：`Type1 & Type2`
- 泛型：`function identity<T>(arg: T): T`
- 接口 vs 类型别名：`interface` 可以合并声明，`type` 支持联合/交叉类型

**常用工具类型：**
| 工具类型 | 作用 |
|---------|------|
| `Partial<T>` | 所有属性变为可选 |
| `Required<T>` | 所有属性变为必选 |
| `Pick<T, K>` | 从 T 中选取部分属性 |
| `Omit<T, K>` | 从 T 中排除部分属性 |
| `Record<K, V>` | 构造键值对类型 |
| `ReturnType<T>` | 获取函数返回值类型 |

**得分点：**
- [ ] 理解 TypeScript 的核心价值（类型安全）
- [ ] 知道 interface 和 type 的区别
- [ ] 了解泛型的基本使用
- [ ] 至少知道 3 个工具类型

**加分项：**
- 能实现自定义工具类型（如 DeepPartial）
- 理解协变和逆变
- 提到装饰器和 reflect-metadata

**追问方向：**
- any、unknown、never 三者的区别？
- 如何实现一个类型安全的事件系统？
- TypeScript 的类型体操了解多少？

---

## Q5: 响应式设计方案

**规范答案：**

**媒体查询（Media Query）：**
```css
@media (max-width: 768px) { /* 手机端样式 */ }
@media (min-width: 769px) and (max-width: 1024px) { /* 平板样式 */ }
@media (min-width: 1025px) { /* 桌面端样式 */ }
```

**常见断点：**
- 手机：< 768px
- 平板：768px ~ 1024px
- 桌面：> 1024px

**现代响应式方案：**
1. **rem + flexible**：根据屏幕宽度动态设置 `font-size`，用 rem 单位
2. **vw/vh**：视口单位，`1vw = 视口宽度的 1%`
3. **clamp()**：`font-size: clamp(14px, 2vw, 18px)`，自动在范围内响应
4. **Container Query**：根据父容器尺寸而非视口调整样式（CSS 新特性）
5. **CSS Grid auto-fit/auto-fill**：`grid-template-columns: repeat(auto-fit, minmax(200px, 1fr))`

**得分点：**
- [ ] 知道媒体查询的基本语法
- [ ] 了解常见断点值
- [ ] 至少知道 2 种现代响应式方案
- [ ] 理解 rem 和 vw 的计算方式

**加分项：**
- 提到 Container Query（容器查询）
- 分析 1px 问题和高清屏适配
- 提到 `aspect-ratio` 属性

**追问方向：**
- 如何处理移动端 1px 问题？
- 响应式图片如何优化（srcset / picture）？
- PWA 是什么？如何实现？
