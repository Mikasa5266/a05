# 前端开发 - CSS 与布局面试知识库

## 1. CSS 盒模型

### 标准盒模型 vs IE 盒模型
- **标准盒模型**（`box-sizing: content-box`）：width = content
- **IE 盒模型**（`box-sizing: border-box`）：width = content + padding + border

### BFC（Block Formatting Context）
**触发条件：**
- float 不为 none
- overflow 不为 visible
- display: inline-block / flex / grid / table-cell
- position: absolute / fixed

**应用场景：**
1. 清除浮动（overflow: hidden）
2. 防止 margin 合并
3. 自适应多栏布局

---

## 2. Flex 布局

### 核心属性
```css
.container {
    display: flex;
    flex-direction: row | column;       /* 主轴方向 */
    justify-content: center | space-between; /* 主轴对齐 */
    align-items: center | stretch;      /* 交叉轴对齐 */
    flex-wrap: wrap;                    /* 换行 */
    gap: 16px;                         /* 间距 */
}
.item {
    flex: 1;            /* flex-grow: 1; flex-shrink: 1; flex-basis: 0% */
    align-self: center; /* 个体覆盖交叉轴对齐 */
}
```

### flex 简写属性
- `flex: 1` = `flex: 1 1 0%`
- `flex: auto` = `flex: 1 1 auto`
- `flex: none` = `flex: 0 0 auto`

---

## 3. Grid 布局

```css
.grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);  /* 三等分 */
    grid-template-rows: auto 1fr auto;      /* 头部自适应+中间填充+底部自适应 */
    grid-template-areas:
        "header header header"
        "sidebar main aside"
        "footer footer footer";
    gap: 16px;
}
```

---

## 4. 响应式设计

### 媒体查询断点
```css
/* 移动优先 */
@media (min-width: 640px) { /* tablet */ }
@media (min-width: 1024px) { /* desktop */ }
@media (min-width: 1280px) { /* wide desktop */ }
```

### 现代方案
- `clamp(min, preferred, max)`：流式排版
- `container queries`：基于容器而非视口
- CSS 自定义属性（变量）主题切换

### 标准答案要点（5星回答）
1. 盒模型两种模式的区别
2. BFC 的触发条件和应用
3. Flex 与 Grid 的使用场景区分
4. 移动端适配方案（rem / vw / clamp）
5. 了解 CSS Houdini、container queries 等新特性
