# AIOps Function Calling 重构准绳

## 定位

本文件是后续智能运维重构的架构准绳。目标是建设个人 AIOps 平台的智能运维核心：全只读数据源、强智能自由度、长期可扩展、可演示、可逐步产品化。

技术底座：

- Go 后端
- Vue 前端
- LLM Function Calling
- Markdown Skill 知识体系
- Go 数据源驱动

本架构彻底放弃 LLM 裸拼 URL、裸拼 HTTP header、裸访问数据源的方案。LLM 只能通过后端注册的标准工具工作。

## 顶层原则

LLM 负责：

- 理解用户问题和告警上下文
- 决定排查路径
- 选择需要调用的工具
- 生成工具业务参数
- 根据工具结果继续多轮探查
- 汇总最终结论、风险和建议

Go 后端负责：

- 工具注册和参数 schema
- 数据源连接和凭证保管
- 鉴权、限流、超时、重试
- 执行真实查询
- 返回标准化、安全、可审计的结果
- 控制最大工具调用轮次

LLM 永远不能：

- 看到 AK、SK、token、kubeconfig、密码
- 看到真实连接地址，除非该地址本身是非敏感展示信息
- 生成 HTTP URL、header、认证信息
- 直接访问网络
- 直接执行 SQL、Shell、kubectl、curl
- 绕过后端工具访问数据源

## 四条红线

1. 所有鉴权信息只存在 Go 后端，绝不传给 LLM。
2. LLM 只能输出工具名和业务参数，禁止输出 URL、header、请求体、token。
3. 单次智能分析最大工具调用轮次为 8 轮，达到上限必须停止并总结已有证据。
4. 单次日志或指标查询最大时间窗口为 24 小时，超过必须拒绝或要求用户缩小范围。

## 分层架构

### 1. Vue 交互层

两个入口：

- 自然语言运维入口：用户直接提问，AI 自主规划查询。
- 告警自动分析入口：AlertManager 或平台告警触发 AI 根因分析任务。

前端职责：

- 展示对话、工具调用轨迹和最终结论
- 展示指标图表、日志片段、K8s 资源摘要
- 对写操作展示确认卡片
- 不保存敏感凭证

### 2. AI 调度核心层

Go 后端实现 `AI Orchestrator`，是智能运维大脑。

职责：

- 收集用户问题、告警上下文、会话历史、启用 Skill、可用工具清单
- 调用 LLM
- 解析 LLM function call
- 校验工具、参数、权限、风险和查询范围
- 路由到工具 handler
- 将工具结果回填给 LLM
- 循环最多 8 轮
- 保存任务记录、工具调用记录、错误信息和最终结论

核心循环：

```text
input context
  -> LLM
  -> tool call
  -> validate
  -> execute driver
  -> append tool result
  -> LLM
  -> ...
  -> final answer
```

终止条件：

- LLM 输出最终答案
- 达到最大调用轮次
- 工具连续失败
- 参数缺失，需要用户补充
- 查询范围或权限不允许

### 3. Tool Registry

所有工具必须注册到后端 `Tool Registry`。

工具定义包含：

```text
name
description
datasource_type
risk_level
params_schema
timeout
max_time_range
max_result_bytes
handler
```

工具风险等级：

- `readonly_low`：普通只读查询，可以直接执行。
- `readonly_expensive`：只读但可能较重，需要限制时间范围和返回大小。
- `write_requires_confirmation`：写操作，必须用户确认。
- `disabled`：当前禁止。

当前阶段默认只开放只读工具。写工具后续单独设计确认协议。

### 4. 数据源驱动层

每类数据源独立 Go driver：

```text
drivers/prometheus
drivers/kubernetes
drivers/tencentcls
drivers/aliyunsls
drivers/mysqlslowlog
```

驱动职责：

- 从数据库加载数据源配置
- 解密敏感配置
- 封装认证
- 调用原生 API
- 处理超时、重试、错误
- 将原始响应转成标准结果
- 脱敏和压缩大结果

新增数据源时，只新增：

- 数据源类型
- driver
- tool handler
- Skill 说明

AI Orchestrator 不因新增数据源而改主流程。

### 5. 存储层

业务库 SQLite 保存：

- 用户、角色、菜单、权限
- 数据源配置
- 告警规则
- 对话历史
- AI 分析任务
- 工具调用轨迹
- Skill 知识
- 长期记忆

后续可选：

- 高频查询轻量缓存
- 工具结果摘要缓存
- 告警分析报告归档

## 标准 Tool Call 协议

LLM 只能输出工具调用 JSON。

示例：

```json
{
  "tool": "query_prometheus",
  "params": {
    "datasource_id": 1,
    "promql": "sum(rate(container_cpu_usage_seconds_total{namespace=\"iot\"}[5m]))",
    "time_range": "1h",
    "step": "1m"
  }
}
```

禁止：

```json
{
  "url": "https://10.10.38.108:6443/api/v1/pods",
  "headers": {
    "Authorization": "Bearer xxx"
  }
}
```

工具返回协议：

```json
{
  "tool": "query_prometheus",
  "ok": true,
  "datasource_id": 1,
  "summary": "查询成功，返回 12 条时序",
  "data": {},
  "meta": {
    "duration_ms": 120,
    "truncated": false
  }
}
```

错误返回协议：

```json
{
  "tool": "query_prometheus",
  "ok": false,
  "error_code": "TIME_RANGE_TOO_LARGE",
  "message": "单次指标查询最大时间范围为 24h，请缩小范围",
  "retryable": true
}
```

## 首批标准工具

### query_prometheus

用途：查询时序指标。

参数：

```json
{
  "datasource_id": 1,
  "promql": "string",
  "time_range": "1h",
  "step": "1m"
}
```

限制：

- 最大时间范围 24h
- PromQL 不允许远程写入
- 返回序列数限制

### query_prometheus_instant

用途：执行即时 PromQL。

参数：

```json
{
  "datasource_id": 1,
  "promql": "string",
  "time": "optional RFC3339"
}
```

### get_k8s_pod_events

用途：获取 Pod 事件，例如 OOM、调度失败、拉镜像失败、容器启停。

参数：

```json
{
  "datasource_id": 4,
  "namespace": "iot",
  "pod_name": "demo-xxx",
  "time_range": "1h"
}
```

### get_k8s_pod_logs

用途：获取 Pod 日志。

参数：

```json
{
  "datasource_id": 4,
  "namespace": "iot",
  "pod_name": "demo-xxx",
  "container": "optional",
  "tail": 200,
  "time_range": "1h",
  "filter": "optional keyword"
}
```

限制：

- 最大 tail 1000
- 最大时间范围 24h
- 默认返回 200 行

### list_k8s_resources

用途：列出 Kubernetes 资源，后端根据 resource 映射到 API。

参数：

```json
{
  "datasource_id": 4,
  "resource": "pods|services|ingresses|deployments|servicemonitors|resourcequotas",
  "namespace": "optional",
  "label_selector": "optional",
  "field_selector": "optional",
  "view": "compact"
}
```

说明：

- LLM 不传 URL。
- 后端决定 API group、version、path。
- `view=compact` 是默认值。

### find_k8s_pod_anomalies

用途：查找异常 Pod。

参数：

```json
{
  "datasource_id": 4,
  "namespace": "optional",
  "restart_threshold": 3,
  "time_range": "optional"
}
```

后端异常判断：

- `phase` 非 `Running/Succeeded`
- Ready 数小于容器总数
- `waiting.reason` 非空
- `terminated.reason` 异常
- `restartCount` 超过阈值
- Pod 删除卡住

### query_tencent_cls

用途：检索腾讯云 CLS 日志。

参数：

```json
{
  "datasource_id": 3,
  "query": "error OR ERROR",
  "time_range": "30m",
  "limit": 50
}
```

### query_aliyun_sls

用途：检索阿里云 SLS 日志。

参数：

```json
{
  "datasource_id": 2,
  "query": "error",
  "logstore": "optional",
  "time_range": "30m",
  "limit": 50
}
```

### query_mysql_slow_sql

用途：查询慢 SQL。

参数：

```json
{
  "datasource_id": 5,
  "database": "optional",
  "time_range": "1h",
  "filter": "optional"
}
```

## OOM 告警分析标准链路

输入：

```json
{
  "alert_name": "PodOOMKilled",
  "namespace": "iot-middleware",
  "pod": "zookeeper-0",
  "starts_at": "2026-07-31T10:00:00+08:00"
}
```

运行链路：

1. Orchestrator 创建分析任务。
2. 注入告警上下文、可用工具、Skill。
3. LLM 调用 `query_prometheus` 查询内存趋势。
4. LLM 调用 `get_k8s_pod_events` 查询 OOM 事件。
5. LLM 调用 `get_k8s_pod_logs` 查询异常前日志。
6. LLM 视情况查询 JVM、容器重启、节点资源。
7. 达到信息充足或 8 轮上限。
8. 输出结论。

最终输出结构：

```text
故障现象
关键证据
可能根因
影响范围
建议动作
后续观察指标
```

## Skill 体系边界

Skill 是 Markdown 知识，不是代码。

Skill 可以描述：

- 平台有哪些工具
- 每个工具怎么用
- 某类告警排查 SOP
- PromQL 模板
- 日志检索模板
- 业务命名规则

Skill 不能：

- 保存密钥
- 保存 token
- 执行代码
- 直接访问数据库
- 直接访问网络
- 覆盖后端安全策略

## 当前代码迁移策略

现状中已有一些分散逻辑：

- `internal/handlers/chat.go` 中的关键词意图识别
- Prometheus 硬编码查询
- SLS/CLS/K8s 的局部工具函数
- 临时的 `k8s.api_get` path 白名单探索

迁移目标：

1. 保留现有可用能力，避免一次性推倒。
2. 新增 `ai` 包：
   - `orchestrator.go`
   - `tool_registry.go`
   - `protocol.go`
   - `memory.go`
3. 新增 `drivers` 包：
   - `prometheus`
   - `kubernetes`
   - `tencentcls`
   - `aliyunsls`
4. 将 `internal/handlers/chat.go` 逐步瘦身为 HTTP handler，只负责接收请求和返回结果。
5. 将所有工具调用记录写入消息 metadata 或单独 `ai_tool_calls` 表。
6. 废弃让 LLM 输出 API path 的长期路线，改为标准业务工具参数。

## 实施顺序

### 阶段 1：协议和注册中心

- 定义 Tool Call JSON 协议
- 定义 Tool Result JSON 协议
- 实现 Tool Registry
- 把现有 Prometheus 查询包装成 `query_prometheus`

### 阶段 2：Orchestrator

- 实现 LLM 多轮工具调用循环
- 支持最大 8 轮
- 支持工具错误回填
- 支持最终回答
- 保存调用轨迹

### 阶段 3：Kubernetes Driver

- 实现 `list_k8s_resources`
- 实现 `find_k8s_pod_anomalies`
- 实现 `get_k8s_pod_events`
- 实现 `get_k8s_pod_logs`
- 默认 compact 输出

### 阶段 4：日志和指标 Driver

- Prometheus query/query_range/label_values
- Tencent CLS query
- Aliyun SLS query/list_logstores

### 阶段 5：告警自动分析

- AlertManager webhook
- 告警分析任务表
- 自动启动 Orchestrator
- 前端告警分析报告页面

### 阶段 6：前端体验

- 工具调用时间线
- 指标图表
- 日志片段
- K8s 资源摘要
- 可复制排查报告

## 验收标准

智能运维达到以下效果才算符合本架构：

- 用户问普通知识问题，直接由 LLM 回答。
- 用户问真实运维数据，LLM 自主选择工具。
- LLM 不知道也不需要知道数据源真实 URL 和密钥。
- 新增一种数据源时，主调度逻辑不用改。
- 单次任务完整记录每次工具调用。
- 任意工具失败时，AI 能解释失败并调整下一步。
- 不再靠无限追加关键词 if/else 扩展能力。
