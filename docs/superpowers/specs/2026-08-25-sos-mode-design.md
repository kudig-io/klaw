# SOS 模式（语音应急快速对话）设计文档

- 日期：2026-08-25
- 状态：已评审待实施
- 范围：Klaw Web 控制台新增 SOS 语音快速对话能力（后端 + 前端 + 配置）

## 1. 背景与目标

Klaw 已具备 Web 控制台、ChatOps 机器人、诊断流水线与 AI 文本诊断助手。运维同学在事故现场最需要的往往不是打开层层页面，而是"喊一嗓子就有答案"。

SOS 模式为 Klaw 增加一个**应急语音入口**：在 Web 控制台通过悬浮按钮或导航菜单一键进入全屏通话页（交互形态参照豆包电话聊天），与实时全双工语音模型对话，支持边说边答与智能打断；既能秒答产品与运维常识类问题，也能实时查询集群真实状态并触发应急诊断。

目标：

1. 全屏语音通话体验：实时流式对话、双向字幕、智能打断、静音/挂断控制。
2. 三层兜底回答逻辑（预置语料 → 集群工具 → 通用知识），保证应急问题"问得到、答得准"。
3. 密钥与集群访问全部收敛在 Klaw 后端，浏览器不接触 DashScope API Key。
4. 语料与配置沿用现有 YAML 体系，运维可独立维护，改配置即可更新答案口径。

## 2. 已确认决策

| 决策点 | 结论 |
|---|---|
| 形态 | Web UI 语音快速对话：全局悬浮 SOS 按钮 + 导航菜单双入口，进入全屏通话页（类豆包电话聊天） |
| 回答引擎 | 三层兜底：预置语料 → 集群工具（tools/function calling）→ 模型通用知识 |
| 语音链路 | 实时全双工 Realtime API（非 ASR+LLM+TTS 拼接，非浏览器 Web Speech） |
| 服务商 | 固定阿里云百炼 DashScope，Qwen-Omni-Realtime 系列（OpenAI Realtime 兼容事件协议） |
| 集群数据 | 接入实时集群查询与应急诊断工具（SOS 定位为应急运维入口） |
| 语料存储 | `configs/sos-faq.yaml`，Go embed 嵌入默认语料，支持外部文件覆盖，重启生效 |
| 浏览器接入方式 | 经 Klaw 后端 WebSocket 代理（方案 A），不浏览器直连、本期不做 WebRTC |

## 3. 三层兜底逻辑（核心）

SOS 会话建立时，后端将「系统提示 + 预置语料 + 工具清单」一次性注入 Realtime 会话（`session.update` 的 `instructions` 与 `tools` 字段）。模型按以下优先级组织回答：

```text
用户语音提问
   │
   ▼
第 1 层：预置语料（FAQ）
   命中主题（产品定位、能力对比、常见操作问题等）→ 严格按语料标准口径作答
   例：「klaw 和市面上的对话机器人有什么不同？」→ 直接背诵三大核心差异
   │ 未命中
   ▼
第 2 层：集群工具（function calling）
   涉及集群实时状态、异常排查、应急操作 → 模型调用 tools 查询 Klaw 后端真实数据后作答
   例：「集群现在健康吗？」「default 命名空间有没有异常 Pod？」→ get_cluster_status / list_pods
   │ 不需要/不适用
   ▼
第 3 层：模型通用知识
   通用 Kubernetes/运维知识等 → 模型凭自身知识作答，并声明未查询实时数据
```

说明：

- 第 1 层不是独立的关键词匹配服务，而是把语料全文注入 system instructions，由模型语义匹配命中——保留 Realtime 全双工低延迟特性，避免"先识别→再匹配→再合成"的串联开销。
- 第 2 层工具在后端进程内执行，模型只拿到脱敏后的 JSON 结果；工具执行失败时错误信息以 `function_call_output` 回传，模型口头告知用户。
- 第 3 层要求模型在未查询实时数据的情况下明确说明"这是通用建议，非当前集群实测数据"，避免误导应急判断。

## 4. 总体架构

```text
┌────────────────────────────┐         ┌─────────────────────────┐         ┌──────────────────────────┐
│  浏览器（SosCallPage）      │  WS(1)  │  Klaw 后端               │  WS(2)  │  DashScope Realtime       │
│  - AudioWorklet 采集 PCM16k │ ◄─────► │  internal/sos            │ ◄─────► │  Qwen-Omni-Realtime       │
│  - 24k 播放队列 + 打断停播   │         │  - session 桥接/事件翻译  │         │  (OpenAI Realtime 兼容协议)│
│  - 字幕/状态机/控制条        │         │  - tools 进程内执行       │         │  semantic_vad 智能打断     │
└────────────────────────────┘         │  - FAQ 注入 instructions │         └──────────────────────────┘
                                       │  依赖: kubernetes.Manager │
                                       │        diag pipeline      │
                                       └─────────────────────────┘
```

- 链路 1（浏览器 ↔ Klaw）：同源 WebSocket `/api/v1/sos/session`，复用现有 Bearer Token 鉴权中间件；音频用二进制帧，控制/事件用 JSON 文本帧。
- 链路 2（Klaw ↔ DashScope）：`wss://{workspace-id}.{region}.maas.aliyuncs.com/api-ws/v1/realtime?model={model}`，`Authorization: Bearer {DASHSCOPE_API_KEY}`。
- 会话配置（`session.update`）：`voice`（默认 Ethan）、`turn_detection.type=semantic_vad`（语义打断）、输入音频 `pcm/16000`、输出音频 `pcm/24000`、`instructions`（系统提示+语料）、`tools`（集群工具 schema）。
- 智能打断：服务端 semantic_vad 检测到用户开口即下发 `input_audio_buffer.speech_started`，后端转发给浏览器，浏览器立即停播本地音频队列并清空待播缓冲。

方案对比结论（为何选后端代理）：

| | A. 后端 WS 代理（选定） | B. 浏览器直连 | C. WebRTC 信令 |
|---|---|---|---|
| 密钥安全 | 留在服务端 | Key 暴露给前端 | 留在服务端 |
| 集群 tools | 进程内直调 Manager | 需前端回调中转 | 进程内直调 |
| 复杂度 | 中（双 WS 生命周期） | 低但不可接受 | 高（SDP/ICE，内网 kind 不友好） |
| 结论 | ✅ 延迟仅增加同源转发一跳 | ❌ | 二期可选 |

## 5. 后端设计（新模块 `internal/sos/`）

| 文件 | 职责 |
|---|---|
| `faq.go` | 加载/校验语料（embed 默认 + 可选外部覆盖），拼装 instructions 片段 |
| `dashscope.go` | Realtime 客户端：建连、鉴权、事件读写、`session.update` 下发 |
| `tools.go` | 集群工具注册表与进程内执行：schema 定义 + 执行器 |
| `session.go` | 会话生命周期：浏览器 WS ↔ DashScope WS 桥接、事件翻译、断线重连（1 次）、超时清理 |
| `server.go`（挂载点） | 在 `internal/api/server.go` 注册路由，复用鉴权中间件 |

依赖方向：`api → sos → kubernetes.Manager / diag pipeline / config`，单向依赖，不反向引用 api 包。

### 5.1 集群工具清单（第 2 层）

| 工具名 | 参数 | 行为 |
|---|---|---|
| `get_cluster_status` | `cluster?` | 集群健康概览：节点/Pod 统计、异常计数、最近 Warning 事件摘要 |
| `list_pods` | `namespace?`, `status?` | 列出 Pod，默认聚焦异常（Pending/CrashLoopBackOff/OOMKilled 等） |
| `get_pod_logs` | `namespace`, `pod`, `tail_lines?`（默认 100，上限 500） | 取最近日志，截断后以文本回传 |
| `list_events` | `namespace?`, `severity?` | 最近 Warning/Error 事件列表 |
| `run_diagnosis` | `namespace?` | 触发现有诊断流水线，同步等待（上限 30s），返回问题摘要与修复建议 |

约束：全部为只读或诊断类操作；本期不注册任何变更类工具（扩缩容/删除等），应急场景避免语音误操作。

### 5.2 浏览器 ↔ 后端 WS 帧协议

- 文本帧（JSON）：
  - 上行：`{"type":"start"}`（携带可选 cluster）、`{"type":"mute"|"unmute"}`、`{"type":"end"}`
  - 下行：`{"type":"session","model":...,"voice":...}`、`{"type":"error","code":...,"message":...}`、事件转发（`user.transcript.delta`、`assistant.transcript.delta`、`tool_call`、`speech_started` 等，统一小写驼峰命名）
- 二进制帧：上行 = 浏览器采集的 PCM16k 单声道裸数据（小端 Int16）；下行 = DashScope 返回的 PCM24k 音频段。后端负责裸 PCM ↔ DashScope base64 事件的编码转换。

## 6. API

| 路由 | 方法 | 说明 |
|---|---|---|
| `/api/v1/sos/status` | GET | `{enabled, ready, model, voice, faq_count}`；未配置 Key 时 `ready=false`，前端据此展示配置引导 |
| `/api/v1/sos/session` | GET（WebSocket Upgrade） | 建立语音会话，复用现有 Token 鉴权中间件 |

## 7. 配置（`config.yaml` 新增 `sos:` 段）

```yaml
sos:
  enabled: true
  dashscope:
    api_key: ""            # 优先读环境变量 KLAW_SOS_DASHSCOPE_API_KEY（与 KLAW_API_TOKEN 注入模式一致）
    workspace_id: ""       # 百炼 Workspace ID（端点子域名）
    region: "cn-beijing"   # 可选 cn-beijing / ap-southeast-1
    model: "qwen3.5-omni-plus-realtime"
    voice: "Ethan"
  faq_file: ""             # 可选外部语料路径；为空时使用 embed 默认语料
  instructions_prefix: ""  # 可选自定义系统提示前缀（追加在默认系统提示之前）
```

默认值原则：`enabled=false` 时全部路由返回 unavailable，保证未使用该功能的部署零开销、零行为变化。

## 8. 前端设计（`web/src/`）

| 新增 | 说明 |
|---|---|
| `components/SosFloatingButton.tsx` | 全局右下角红色悬浮按钮（所有页面可见），点击进入 `/sos` |
| `pages/SosCallPage.tsx` | 全屏通话页：中央头像 + 呼吸/波形动画、连接状态（连接中/通话中/重连中/已结束）、双向实时字幕、底部控制条（静音、挂断） |
| `hooks/useSosSession.ts` | 会话状态机：WS 连接、AudioWorklet 采集（PCM16k 单声道）、24k 播放队列（AudioBufferSourceNode 调度）、`speech_started` 停播打断、字幕累积 |
| `lib/sosApi.ts` | status 探测与会话 WS 封装 |
| 路由/导航 | `App.tsx` 新增 `/sos` 路由与 navItems 菜单项（SOS，红色标识） |

麦克风权限：`getUserMedia({audio:{echoCancellation:true,noiseSuppression:true,autoGainControl:true}})`；非安全上下文（非 HTTPS/localhost）或拒绝授权时给出明确引导文案。视觉规范沿用现有 Tailwind 设计令牌体系。

## 9. 预置语料文件格式与种子内容

`configs/sos-faq.yaml`（同内容作为默认语料 embed 进 `internal/sos/faq/seed.yaml`，运行时优先读 `sos.faq_file` 指定的外部文件）：

```yaml
faqs:
  - id: klaw-vs-chatbot
    question: "klaw 和市面上的对话机器人有什么不同？"
    answer: |
      三大核心差异：一、客服行业专属大模型——基于客服语料训练，
      提出 SDPO 段级对齐与 EPO 多轮强化学习算法；二、全双工实时交互引擎——
      端到端响应延迟 < 1.8s，支持智能打断与降噪；三、大规模业务实践
    tags: [产品, 定位, 对比]
```

注入格式：系统提示中追加「以下为标准问答语料，命中主题时严格按语料口径回答」+ 逐条 `Q/A` 文本。语料仅用于 prompt 注入，不落库、不做向量检索（本期）。

## 10. 错误处理

| 场景 | 行为 |
|---|---|
| 未配置/无效 DashScope Key | `/sos/status` 返回 `ready=false`；通话页展示配置引导（指明 `sos.dashscope.*` 配置项与环境变量） |
| DashScope 建连失败/中途断开 | 后端自动重连 1 次；失败则下发 `error` 事件并结束会话，前端提示"服务中断，请重试" |
| 浏览器 WS 断开 | 后端关闭上游会话、释放资源（无泄漏） |
| 工具执行失败/超时 | 错误信息经 `function_call_output` 回传，模型口头告知；`run_diagnosis` 超 30s 返回"诊断超时，建议到诊断页查看" |
| 麦克风不可用 | 前端检测安全上下文与权限，展示引导，不建连 |
| 会话空闲超时 | 后端 5 分钟无音频活动自动结束会话并释放上游连接 |

## 11. 安全

- DashScope API Key 只存在于后端（config/环境变量），浏览器全程不接触。
- 会话 WS 走现有 Bearer Token 鉴权中间件；音频与字幕不持久化、不落审计库（审计仅记录"会话开始/结束/工具调用"元数据，不含音频与转写内容）。
- 工具只读 + 诊断，无写操作；`get_pod_logs` 结果截断防大包。

## 12. 测试计划

后端（Go test）：

- 语料加载：合法/非法 YAML、embed 默认、外部覆盖、instructions 拼装快照。
- tools：schema 校验、fake k8s client 下各工具执行与错误路径、日志截断。
- session：事件翻译（上行/下行帧映射）、断线重连一次、空闲超时清理（用假 DashScope WS server）。

前端（Vitest + Testing Library，沿用 `__tests__` 结构与 msw mock 模式）：

- `SosCallPage` 状态机（连接中/通话中/错误/结束）与字幕渲染。
- `useSosSession`：mock WS 下采集帧发送、`speech_started` 停播、挂断清理。
- 悬浮按钮与路由/导航入口渲染。

验收（手动）：kind 集群 + 真实 DashScope Key，语音问三个层次问题各一：语料命中（产品对比）、集群实时（"现在有没有异常 Pod"）、通用知识（"Pod Pending 常见原因"），验证回答层次与打断体验。

## 13. 非目标（本期不做）

- 语料在线管理界面（SQLite + CRUD，二期）
- 通话录音与回放、多音色切换 UI
- WebRTC 接入模式
- 跨页面会话保持（离开通话页即挂断）
- ChatOps（钉钉/飞书）侧 SOS 文字秒回
- 语料向量检索/RAG

## 13.1 变更记录

- 2026-08-25（实施后扩展）：新增实时上游 provider 抽象（`sos.provider: dashscope | glm`）。
  上游地址、鉴权、`session.update` 字段差异收敛在 `internal/sos/dashscope.go` 的
  `BuildUpstreamURL / DialRealtime / BuildSessionUpdateFor`；GLM-Realtime 使用
  `server_vad`、`input_audio_format/output_audio_format` 平铺字段、`Authorization: Bearer`
  服务端直连鉴权。各厂商事件名与现有翻译层一致，心跳事件静默忽略。

## 14. 验收标准

1. 悬浮按钮与导航菜单均可进入 `/sos` 全屏通话页，未配置 Key 时展示配置引导。
2. 配置正确时完成一次全双工语音对话：可打断、有双向字幕、可静音、挂断即释放会话。
3. 三层兜底按优先级生效：语料问题按标准口径回答；集群问题经工具返回真实数据；通用问题声明非实测数据。
4. `go test ./...` 与 `web` 测试全绿；未启用 `sos` 时现有功能零回归。
