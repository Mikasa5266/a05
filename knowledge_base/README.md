# 智聘AI 面试知识库（RAG Knowledge Base）

## 目录结构

```
knowledge_base/
├── java_backend/              # Java后端开发 (6文件: JVM/并发/Spring/MySQL/Redis/项目)
├── python_backend/            # Python后端 (2文件: 核心语法/数据库与系统)
├── go_backend/                # Go后端开发 (2文件: 核心/微服务云原生)
├── web_frontend/              # 前端开发 (3文件: React&Vue/浏览器/CSS)
├── qa_testing/                # 测试工程师 (1文件: 测试方法论)
├── cs_fundamentals/           # 计算机通识 (2文件: 数据结构算法/设计模式)
├── behavioral/                # 行为面试 & HR面 (1文件: 软技能)
├── scoring_rubrics/           # 评分维度量表 (2文件: 通用维度/岗位权重)
│   ├── dimension_rubrics.json   # 4维度(技术深度/表达/逻辑/完整度)×5级评分锚点
│   └── position_scoring.json    # 5岗位×能力权重+等级期望
├── follow_up_chains/          # 追问逻辑链 (5文件: Java/Python/Go/前端/通用)
│   └── *_chains.md              # 每知识点3层递进: 概念→原理→实战
└── learning_resources/        # 推荐学习资源 (1文件)
    └── resources_index.md       # 书籍/视频/博客按知识点索引
```

## 文件格式说明

- `.md` 文件：结构化知识点文档，用于 RAG 检索上下文
- `.json` 文件：结构化数据（评分量表、追问链、资源索引）

## 评分等级定义

| 等级 | 分数段 | 含义 |
|------|--------|------|
| 1 星 | 0-20   | 仅知道名词，无法展开 |
| 2 星 | 21-40  | 知道基本概念但有严重错误 |
| 3 星 | 41-60  | 能答出主干，缺乏深度 |
| 4 星 | 61-80  | 准确完整，有一定展开 |
| 5 星 | 81-100 | 深入原理+实践+权衡分析 |

## 使用方式

知识库通过 RAG (Retrieval Augmented Generation) 检索服务接入 AI 面试评估流程：
1. 面试问题触发检索 → 匹配最相关知识文档
2. 检索结果注入评估 Prompt → AI 基于标准答案进行对比评分
3. 追问逻辑链驱动深度面试 → 逐层追问验证真实能力
