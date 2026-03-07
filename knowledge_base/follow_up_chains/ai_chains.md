# 追问逻辑链 - AI 工程师

> 每条追问链为 3 层递进，用于评估候选人的知识深度

---

## 链1: Transformer 架构

**L1 基础**: Self-Attention 的计算过程是怎样的？Q/K/V 分别是什么？
**L2 原理**: 为什么需要缩放因子 √d_k？多头注意力有什么优势？
**L3 实战**: Self-Attention 的 O(n²) 复杂度在长文本场景下有什么问题？Flash Attention 怎么优化的？

---

## 链2: 大模型微调

**L1 基础**: 什么是 SFT（指令微调）？和预训练的区别？
**L2 原理**: LoRA 的原理是什么？为什么有效？秩 r 怎么选？
**L3 实战**: 你做过模型微调吗？数据怎么准备的？效果怎么评估？遇到什么问题？

---

## 链3: RAG 系统

**L1 基础**: RAG 是什么？为什么需要它（相比直接使用 LLM）？
**L2 原理**: 向量检索的原理是什么？Embedding 是怎么工作的？
**L3 实战**: RAG 检索结果不相关时怎么优化？你用过什么 Re-ranking 策略？

---

## 链4: Prompt Engineering

**L1 基础**: Few-Shot 和 Zero-Shot Prompting 有什么区别？
**L2 原理**: Chain of Thought（思维链）为什么能提升推理效果？
**L3 实战**: 如何防止 Prompt 注入攻击？你的项目中做了哪些安全防护？

---

## 链5: AI Agent

**L1 基础**: AI Agent 的核心组件有哪些（规划、记忆、工具、行动）？
**L2 原理**: ReAct 框架是怎么工作的？Function Calling 的机制是什么？
**L3 实战**: 你开发过 Agent 应用吗？如何处理 Agent 的幻觉和错误恢复？

---

## 链6: 模型部署

**L1 基础**: 模型量化是什么？INT8 和 INT4 量化的效果差异？
**L2 原理**: KV Cache 是怎么加速推理的？PagedAttention 的原理？
**L3 实战**: 你部署过大模型吗？用的什么框架？延迟优化做了什么？

---

## 链7: Diffusion Model

**L1 基础**: Diffusion Model 的基本原理是什么（加噪和去噪）？
**L2 原理**: Stable Diffusion 为什么在潜空间操作？Classifier-Free Guidance 的作用？
**L3 实战**: 如何用 ControlNet 实现条件控制生成？LoRA 在图像生成中怎么用？
