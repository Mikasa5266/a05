# AI 工程师 - 大模型与 LLM 应用

> 格式说明：每道题包含「问题 → 规范答案 → 得分点 → 加分项 → 追问方向」

---

## Q1: Transformer 架构的核心机制

**规范答案：**

**整体架构：**
Transformer 由 Encoder 和 Decoder 组成（如原始的机器翻译模型），现代 LLM 主要使用 Decoder-only 架构（GPT 系列）或 Encoder-only 架构（BERT）。

**自注意力机制（Self-Attention）：**
1. 输入经过三个线性变换生成 Query(Q)、Key(K)、Value(V)
2. 计算注意力权重：Attention(Q,K,V) = softmax(QK^T / √d_k) · V
3. √d_k 是缩放因子，防止点积过大导致 softmax 梯度消失

**多头注意力（Multi-Head Attention）：**
- 将 Q/K/V 分成多个头，各自计算注意力后拼接
- 使模型关注不同子空间的特征

**其他核心组件：**
- **位置编码（Positional Encoding）**：正弦/余弦函数或可学习的位置编码，注入序列位置信息
- **Layer Normalization**：稳定训练，加速收敛
- **残差连接（Residual Connection）**：缓解深层网络的梯度消失
- **FFN（前馈网络）**：两层全连接 + 激活函数，增加非线性表达

**得分点：**
- [ ] 理解 Self-Attention 的 Q/K/V 计算过程
- [ ] 知道缩放因子 √d_k 的作用
- [ ] 理解多头注意力的意义
- [ ] 知道位置编码的必要性（Transformer 本身不感知位置）

**加分项：**
- 分析 Self-Attention 的 O(n²) 复杂度及其优化方案（Flash Attention 等）
- 对比 RoPE vs 正弦位置编码
- 提到 KV Cache 在推理中的加速作用

**追问方向：**
- 为什么 Transformer 替代了 RNN/LSTM？
- Flash Attention 是怎么优化的？
- 什么是因果注意力（Causal Attention）和掩码？

---

## Q2: 大模型的预训练与微调方法

**规范答案：**

**预训练阶段：**
- **自回归预训练（GPT 系列）**：预测下一个 token，左到右生成
- **掩码语言模型（BERT）**：随机遮盖 token，双向预测
- 训练数据：大规模无标注文本（CommonCrawl、Books 等）
- 训练目标：学习通用的语言表示

**指令微调（Instruction Tuning / SFT）：**
- 使用高质量的 instruction-response 数据对进行微调
- 让模型学会遵循指令、理解用户意图
- 代表：ChatGPT 的 SFT 阶段

**RLHF（基于人类反馈的强化学习）：**
1. 收集多个模型回答，人类标注偏好排序
2. 训练奖励模型（Reward Model）
3. 用 PPO 算法对齐人类偏好
- 目的：让模型输出更有帮助、更安全、更诚实

**参数高效微调（PEFT）：**
| 方法 | 原理 | 参数量 |
|------|------|--------|
| LoRA | 在权重矩阵旁加低秩分解矩阵 | <1% |
| QLoRA | LoRA + 4-bit 量化 | 更少 |
| Adapter | 在层间插入小型适配器 | ~2-5% |
| Prefix Tuning | 在输入前添加可训练的虚拟 token | <1% |
| Prompt Tuning | 只训练 soft prompt | 极少 |

**得分点：**
- [ ] 理解预训练和微调的区别
- [ ] 知道 SFT 和 RLHF 的作用
- [ ] 至少了解 2 种 PEFT 方法
- [ ] 理解 LoRA 的基本原理

**加分项：**
- 分析 DPO 替代 RLHF 的趋势
- 讨论 LoRA 的秩 r 选择对效果的影响
- 提到 MoE（Mixture of Experts）架构

**追问方向：**
- LoRA 为什么有效？低秩假设是什么？
- RLHF 和 DPO 的区别？
- 如何评估微调后的模型效果？

---

## Q3: RAG（检索增强生成）系统

**规范答案：**

**RAG 的核心思想：**
将外部知识库的检索结果作为上下文，输入大模型生成回答。解决 LLM 的知识过时和幻觉问题。

**标准 RAG 流程：**
1. **文档处理**：加载 → 分块（Chunking）→ 向量化（Embedding）→ 存入向量数据库
2. **检索**：用户查询 → 向量化 → 向量相似度搜索（如余弦相似度）→ 返回 Top-K 文档
3. **生成**：将检索到的文档 + 用户问题组合成 Prompt → LLM 生成回答

**关键技术选型：**
| 环节 | 常见方案 |
|------|---------|
| Embedding 模型 | text-embedding-3-small, BGE, m3e |
| 向量数据库 | Milvus, Pinecone, ChromaDB, FAISS |
| 分块策略 | 固定大小、语义分割、递归分割 |
| 检索增强 | 混合检索（向量+关键词）、Re-rank |

**高级 RAG 优化：**
- **Query 改写**：用 LLM 优化用户原始查询
- **HyDE**：让 LLM 先生成假设文档，再用假设文档检索
- **Re-ranking**：用交叉编码器对检索结果重排序
- **Chunk 优化**：重叠切分、父子文档策略

**得分点：**
- [ ] 理解 RAG 的核心流程（索引→检索→生成）
- [ ] 知道为什么需要 RAG（知识更新、减少幻觉）
- [ ] 了解向量数据库的作用
- [ ] 知道 Embedding 和语义检索的原理

**加分项：**
- 讨论分块大小对检索效果的影响
- 提到混合检索（BM25 + 向量检索）
- 分析 RAG 和微调的适用场景对比

**追问方向：**
- 如何评估 RAG 系统的效果？
- 检索结果不相关时如何处理？
- RAG vs 微调，各自的优劣和适用场景？

---

## Q4: Prompt Engineering 技巧

**规范答案：**

**核心 Prompt 策略：**

| 策略 | 说明 | 效果 |
|------|------|------|
| **Zero-Shot** | 直接提问，不给示例 | 简单任务有效 |
| **Few-Shot** | 提供 2-5 个示例 | 显著提升复杂任务效果 |
| **CoT（思维链）** | 要求模型"一步步思考" | 提升推理任务准确率 |
| **Role Prompting** | 给模型设定角色 | 引导输出风格和知识域 |
| **Self-Consistency** | 多次采样取众数 | 提升推理稳定性 |
| **ReAct** | 推理（Reason）+ 行动（Act）交替 | 工具调用、复杂任务 |

**Prompt 设计原则：**
1. **明确具体**：清楚说明任务、格式、限制
2. **结构化**：使用标记分隔输入的不同部分
3. **给出正例和反例**：让模型理解边界
4. **控制输出格式**：要求 JSON/Markdown/表格等结构化输出
5. **迭代优化**：根据输出结果不断调整 Prompt

**System Prompt 设计：**
- 包含角色定义、任务说明、约束条件、输出格式要求
- 保持简洁但完整

**得分点：**
- [ ] 知道至少 3 种 Prompt 策略
- [ ] 理解 CoT 的原理和适用场景
- [ ] 了解 Few-Shot 的样本选择影响
- [ ] 能给出结构化的 Prompt 设计方法

**加分项：**
- 分析 CoT 在小模型上效果不佳的原因
- 提到 Tree of Thought（ToT）和 Graph of Thought
- 讨论 Prompt 注入攻击和防御

**追问方向：**
- CoT 为什么有效？有什么局限？
- 如何防止 Prompt 注入攻击？
- Function Calling 的原理是什么？

---

## Q5: AI Agent 架构与工具调用

**规范答案：**

**AI Agent 的定义：**
以 LLM 为核心推理引擎，能感知环境、规划行动、调用工具、自主完成复杂任务的系统。

**核心组件：**
1. **规划（Planning）**：任务分解、制定行动计划
   - 方法：ReAct、Plan-and-Execute、Reflexion
2. **记忆（Memory）**：
   - 短期记忆：对话上下文（context window）
   - 长期记忆：向量数据库存储的历史信息
3. **工具使用（Tool Use）**：
   - 搜索引擎、代码执行、API 调用、数据库查询
   - 通过 Function Calling 或 JSON 指令调用
4. **行动（Action）**：执行具体操作并获取反馈

**常见 Agent 框架：**
| 框架 | 特点 |
|------|------|
| LangChain | 生态丰富，组件化设计 |
| LangGraph | 基于图的工作流编排 |
| AutoGPT | 自主规划和执行 |
| CrewAI | 多 Agent 协作 |
| Dify / Coze | 低代码 Agent 平台 |

**Multi-Agent 系统：**
- 多个 Agent 分工协作，各自负责不同领域
- 协作模式：分层结构、对等协商、投票决策

**得分点：**
- [ ] 理解 Agent 的核心组件（规划、记忆、工具、行动）
- [ ] 知道 ReAct 框架的工作方式
- [ ] 了解 Function Calling 的机制
- [ ] 能分析 Agent 的适用场景

**加分项：**
- 有实际的 Agent 项目经验
- 分析 Agent 的可靠性和错误恢复机制
- 提到 MCP（Model Context Protocol）

**追问方向：**
- Agent 的幻觉问题如何解决？
- 如何评估 Agent 的效果？
- Multi-Agent 的通信协议如何设计？

---

## Q6: 模型部署与推理优化

**规范答案：**

**模型量化：**
| 精度 | 内存占用（7B 模型） | 说明 |
|------|-------------------|------|
| FP32 | ~28GB | 原始精度 |
| FP16 | ~14GB | 半精度，几乎无损 |
| INT8 | ~7GB | 8-bit 量化 |
| INT4 | ~3.5GB | 4-bit 量化（GPTQ / AWQ） |

**推理优化技术：**
1. **KV Cache**：缓存已计算的 Key/Value，避免重复计算
2. **Flash Attention**：通过分块计算减少显存访问，降低 O(n²) 空间到 O(n)
3. **Continuous Batching**：动态批处理请求，提高 GPU 利用率
4. **投机采样（Speculative Decoding）**：小模型先生成草稿，大模型并行验证
5. **张量并行 / 流水线并行**：多 GPU 部署

**推理框架：**
| 框架 | 特点 |
|------|------|
| vLLM | PagedAttention，高吞吐推理 |
| TensorRT-LLM | NVIDIA 优化，低延迟 |
| llama.cpp | CPU/边缘设备推理 |
| Ollama | 本地部署，易用 |
| Triton | 推理服务框架 |

**得分点：**
- [ ] 理解模型量化的基本概念和效果
- [ ] 知道 KV Cache 的作用
- [ ] 至少了解 2 种推理优化技术
- [ ] 知道常见推理框架

**加分项：**
- 分析 vLLM 的 PagedAttention 原理
- 讨论延迟 vs 吞吐量的权衡
- 有实际部署大模型的经验

**追问方向：**
- 量化会损失多少精度？如何评估？
- vLLM 和 TensorRT-LLM 各适合什么场景？
- 如何实现大模型的流式输出？
