# 运维中枢

运维中枢是一个面向运维、SRE、平台工程和基础设施团队的统一运维平台。项目目标不是只做监控展示，而是逐步建设一套覆盖数据接入、可观测查询、告警治理、拨测监控、自动化执行、证书生命周期管理、故障诊断和权限审计的完整运维工作台。

## 目标

- 看得见：指标、日志、链路、事件、资产、证书、变更都能统一查询和关联。
- 管得住：告警、巡检、脚本、证书、任务、工作流都有权限、审计、记录和失败追踪。
- 能自动：常见故障自动定位，重复操作自动执行，高风险动作有审批和回滚保护。

## 技术栈

- 前端：Vue 3、Vite、Vue Router、Pinia、Element Plus、ECharts
- 后端：Go、Gin、GORM
- 数据库：MySQL
- 主要集成：Prometheus、阿里云 CDN/DCDN，后续扩展日志、链路、多云和 Kubernetes

## 本地启动

首次使用需先准备空的 MySQL 数据库，并仅执行一次数据库初始化：

```bash
cd backend
go run ./cmd/dbinit
```

正常启动不会自动建表、修改菜单或写入初始数据。已有数据库切换到版本化迁移时，请按 [数据库迁移说明](backend/database/migrations/README.md) 登记一次基线版本。

后端智能客服支持 Anthropic Messages API 兼容服务，通过环境变量配置：

```bash
export ANTHROPIC_BASE_URL="https://dreame-claude.dreame.tech"
export ANTHROPIC_API_KEY="你的 API Key"
export ANTHROPIC_MODEL="deepseek-v4-pro"
```

前端 Vite 已绑定 `0.0.0.0:5173`。其他电脑访问时使用运行机器的局域网 IP，例如 `http://192.168.x.x:5173`。

## 当前已具备能力

### 账号与权限

- 用户登录
- JWT 鉴权
- 用户管理
- 角色管理
- 权限管理
- 菜单管理
- 动态菜单
- 动态路由
- 初始化管理员账号

默认账号：

```text
admin / admin123
```

### 接入中心

当前支持维护多种接入类型：

- Prometheus
- Elasticsearch
- InfluxDB
- MySQL
- HTTP
- SLS
- Jaeger
- Loki
- CLS

规划能力：

- 连接测试
- 认证方式配置
- Basic Auth
- Bearer Token
- 自定义配置 JSON
- 接入状态管理
- 隧道节点关联
- 工作空间隔离
- 敏感认证信息脱敏展示
- 敏感字段加密存储

### Prometheus 即时查询

当前能力：

- Prometheus 数据源选择
- PromQL 输入
- 指标名自动补全
- Label 自动补全
- Label value 自动补全
- Instant Query
- Range Query
- Table / Graph 切换
- Table 切换自动执行即时查询
- Graph 切换自动执行范围查询
- Evaluation time
- Range / End time / Step 参数
- 图表统计信息
- 查询错误提示

规划能力：

- 查询历史
- 查询收藏
- 常用 PromQL 模板
- 多查询表达式对比
- 查询结果导出
- Prometheus metadata 缓存
- 自动识别 counter / gauge / histogram
- PromQL 语法高亮
- PromQL 语法校验
- PromQL AI 生成和解释

### 仪表盘管理

当前能力：

- 仪表盘名称
- 仪表盘 URL
- 描述
- 排序
- 状态

规划能力：

- Grafana Dashboard 嵌入
- Dashboard 分组
- 工作空间级 Dashboard
- 常用大盘收藏
- 服务视角大盘
- 告警视角大盘
- 证书视角大盘

### 巡检中心

拨测任务支持规则组管理：从规则组列表进入组内任务，创建或编辑任务时选择所属组。支持规则组创建、编辑和删除空组，默认规则组保留；任务换组后历史记录不变。YAML 导入导出按当前组操作，同名任务仅在组内匹配。
已有部署需先执行 `V2026090701__inspection_rule_groups.sql`，再启动新版后端。

当前能力：

- 即时拨测
- 巡检项管理
- 巡检执行
- 执行记录
- Prometheus 指标检查
- HTTP 健康检查

规划能力：

- TCP 检查
- DNS 检查
- TLS 证书检查
- ICMP 检查
- MySQL 连通性检查
- Redis 连通性检查
- Kubernetes 资源状态检查
- 巡检模板
- 巡检报告
- 巡检失败自动生成工单
- 巡检结果趋势分析

### 脚本管理

当前能力：

- Shell 脚本
- Python 脚本
- 脚本启停
- 脚本执行
- 执行输出
- 执行耗时
- 执行状态

规划能力：

- 参数化脚本
- 脚本审批
- 脚本版本管理
- 脚本执行目标管理
- 并发控制
- 超时控制
- 高危命令拦截
- 输出脱敏
- 回滚脚本

### 工作流管理

当前能力：

- 工作流定义
- 工作流步骤
- 工作流执行
- 执行记录

规划能力：

- 可视化编排
- 条件分支
- 人工确认节点
- 失败重试
- 超时终止
- 通知节点
- 巡检节点
- 脚本节点
- HTTP 节点
- 证书更新节点
- 变更关联节点

### 定时任务

当前能力：

- Cron 表达式
- 定时任务列表
- 定时任务执行
- 执行记录
- Cron 帮助

规划能力：

- Cron 表达式校验
- 下一次执行时间预览
- 任务锁
- 分布式调度
- 失败重试策略
- 通知策略
- 执行窗口
- 禁止时段

### 工作空间

当前能力：

- 工作空间列表
- 工作空间创建
- 工作空间编辑
- 工作空间删除
- 工作空间绑定数据源

规划能力：

- 工作空间成员
- 工作空间角色
- 工作空间资源隔离
- 工作空间视角 Dashboard
- 工作空间告警规则
- 工作空间巡检规则
- 工作空间操作审计

### 智能运维助手

当前能力：

- 会话列表
- 会话消息
- 工作空间上下文
- Chat API

规划能力：

- PromQL 生成
- PromQL 解释
- 告警摘要
- 根因分析
- 巡检报告总结
- 日志摘要
- 变更影响分析
- 知识库问答
- Runbook 推荐
- 自动生成排障步骤

### SSL 证书检测

证书管理模块已移除。拨测任务保留 HTTP、ICMP、TCP 和证书到期四种类型，独立的 MQTT 与 SSL 握手已移除。网站证书有效期监控使用「证书到期」类型，默认提前 30 天告警、每 6 小时检查一次。到期异常和检测失败均沿用任务的连续失败告警策略，并分别显示原因。
共享节点管理继续供拨测和脚本任务使用。已有部署需执行 `V2026090601__remove_certificate_management.sql`，移除旧菜单并停用旧证书监控；历史数据保留。

## 后续完整模块规划

### 1. 数据接入中心

数据接入是 AIOps 的底座。平台需要统一接入不同来源的数据，并形成标准化的数据源模型。

能力范围：

- 指标数据源接入
- 日志数据源接入
- 链路数据源接入
- 事件数据源接入
- CMDB / 资产数据源接入
- 云厂商 API 接入
- Kubernetes 集群接入
- 数据源健康检查
- 数据源权限隔离
- 数据源密钥托管
- 数据源调用审计

优先级：

1. Prometheus
2. Loki / Elasticsearch
3. Jaeger / OpenTelemetry
4. 阿里云 / 腾讯云 / 华为云
5. Kubernetes
6. CMDB

### 2. 可观测查询中心

查询中心用于快速定位问题，是日常运维入口。

能力范围：

- PromQL 查询
- LogQL 查询
- Elasticsearch DSL 查询
- Trace 查询
- SQL 查询
- 多数据源联合查询
- 查询历史
- 查询收藏
- 查询模板
- 查询结果导出
- 查询结果图表化
- 查询上下文保存
- AI 生成查询语句
- AI 解释查询结果

典型场景：

- 查询某服务 QPS
- 查询某接口错误率
- 查询某 Pod CPU 使用率
- 查询某时间段日志错误
- 查询某 Trace 慢调用
- 查询发布前后指标变化

### 3. 告警治理中心

告警治理目标是减少噪音、提高定位效率。

能力范围：

- 告警接入
- 告警列表
- 告警详情
- 告警状态流转
- 告警认领
- 告警关闭
- 告警静默
- 告警抑制
- 告警聚合
- 告警去重
- 告警升级
- 告警通知
- 告警规则管理
- 告警规则测试
- 告警影响面分析
- 告警关联服务
- 告警关联主机
- 告警关联 Pod
- 告警关联变更
- 告警关联日志
- 告警关联 Trace
- 告警关联 Runbook

告警来源：

- Prometheus Alertmanager
- Grafana Alerting
- 云监控
- 自定义 Webhook
- 巡检任务
- 脚本任务
- 证书扫描

### 4. 巡检与健康检查

巡检用于提前发现风险，减少故障发生。

能力范围：

- 巡检模板
- 巡检分组
- 巡检计划
- 即时巡检
- 定时巡检
- 巡检报告
- 巡检评分
- 巡检基线
- 巡检趋势
- 巡检失败通知
- 巡检失败生成工单
- 巡检失败触发工作流

巡检类型：

- HTTP
- TCP
- DNS
- ICMP
- TLS
- Prometheus
- MySQL
- Redis
- Kafka
- Kubernetes
- 云资源
- 证书

### 5. 自动化执行中心

自动化执行中心负责把人工重复操作平台化。

能力范围：

- 脚本管理
- 脚本执行
- 参数输入
- 批量执行
- 执行审批
- 执行日志
- 执行结果归档
- 命令风险识别
- 高危命令拦截
- 执行超时控制
- 并发控制
- 失败重试
- 回滚动作

执行目标：

- 本机
- SSH 主机
- Kubernetes Pod
- Kubernetes Job
- 云厂商 API
- HTTP API
- 数据库

### 6. 工作流编排中心

工作流用于把巡检、脚本、通知、审批、证书更新等能力串起来。

规划节点：

- 开始节点
- 结束节点
- 人工审批节点
- HTTP 请求节点
- Shell 脚本节点
- Python 脚本节点
- Prometheus 查询节点
- 巡检节点
- 证书更新节点
- 通知节点
- 条件判断节点
- 等待节点
- 回滚节点

典型流程：

- 发布前巡检
- 发布后验证
- 故障自愈
- 证书到期更新
- 数据源健康检查
- 主机初始化
- 服务扩容
- 服务重启

### 7. 证书生命周期管理

证书生命周期管理要覆盖发现、评估、更新、验证、审计。

能力范围：

- 域名管理
- 证书库存
- 到期时间扫描
- 过期风险分级
- 证书申请
- 证书上传
- 证书部署
- 证书替换
- 证书回滚
- 证书在线验证
- 证书链验证
- 私钥匹配验证
- 多云同步
- 操作审计
- 到期通知

目标平台：

- 阿里云 CDN
- 阿里云 DCDN
- 阿里云 CLB
- 阿里云 ALB
- 阿里云 NLB
- 腾讯云 CDN
- 腾讯云 CLB
- 华为云 CDN
- Kubernetes Secret
- Nginx
- Ingress Controller

### 8. 资产与拓扑中心

资产是故障关联分析的基础。

能力范围：

- 主机资产
- 云资源资产
- Kubernetes 资源
- 服务资产
- 域名资产
- 证书资产
- 数据库资产
- 中间件资产
- 资产标签
- 资产分组
- 资产负责人
- 资产和工作空间绑定
- 服务拓扑
- 调用关系
- 依赖关系
- 影响面分析

### 9. 变更管理

大量故障都和变更有关，AIOps 必须把变更纳入分析。

能力范围：

- 发布记录
- 配置变更
- 数据库变更
- 云资源变更
- Kubernetes 变更
- 证书变更
- 变更审批
- 变更窗口
- 变更风险评估
- 变更后指标对比
- 变更后自动巡检
- 变更失败回滚建议
- 变更关联告警
- 变更关联故障

### 10. 故障诊断与 AI 分析

AI 能力不应该先做成聊天玩具，而应该嵌入具体运维场景。

能力范围：

- 告警摘要
- 根因候选
- 影响面分析
- 异常指标解释
- 日志摘要
- Trace 慢调用解释
- PromQL 生成
- LogQL 生成
- SQL 生成
- Runbook 推荐
- 故障复盘初稿
- 处理步骤建议
- 历史相似故障匹配
- 知识库问答

输入上下文：

- 当前告警
- 最近指标
- 最近日志
- 最近 Trace
- 最近发布
- 最近配置变更
- 资产拓扑
- 历史工单
- 历史故障复盘
- Runbook

输出要求：

- 给出证据
- 给出置信度
- 给出下一步排查命令
- 给出风险提示
- 不直接执行高危动作
- 高危动作必须审批

### 11. 通知与协同

平台需要把事件推到人和流程。

能力范围：

- 通知渠道管理
- 飞书
- 钉钉
- 企业微信
- 邮件
- Webhook
- 告警通知
- 巡检通知
- 证书到期通知
- 工作流通知
- 审批通知
- 值班人通知
- 通知升级

### 12. 审计与安全

运维平台涉及高权限操作，必须默认重视安全。

能力范围：

- 登录审计
- 操作审计
- API 调用审计
- 敏感字段脱敏
- 密钥加密存储
- 密钥轮换
- 最小权限
- 高危操作二次确认
- 高危操作审批
- IP 白名单
- Token 过期控制
- 会话管理
- 数据库备份
- 审计日志导出

## 技术架构

### 前端结构

```text
frontend/src
├── api             API 封装
├── assets          静态资源
├── directives      权限指令
├── router          动态路由
├── stores          Pinia 状态
├── utils           通用工具
└── views           页面模块
```

页面模块：

```text
frontend/src/views
├── system          系统管理
├── monitor         可观测与接入中心
├── inspection      巡检与自动化
├── certificate     证书管理
├── Chat.vue        智能运维助手
└── Dashboard.vue   首页概览
```

### 后端结构

```text
backend
├── config          配置
├── database        数据库初始化和种子数据
├── handlers        API 处理器
├── middleware      鉴权和跨域
├── models          数据模型
├── routes          路由注册
└── utils           通用工具
```

### 数据库表规划

已有或近期表：

- users
- roles
- permissions
- menus
- datasources
- dashboards
- tunnels
- workspaces
- inspection_items
- inspection_records
- scripts
- script_executions
- workflows
- workflow_executions
- scheduled_tasks
- scheduled_task_executions
- chat_conversations
- chat_messages
- certificate_rotation_records

后续新增表：

- alerts
- alert_rules
- alert_events
- alert_silences
- assets
- services
- service_dependencies
- certificate_assets
- certificate_scan_records
- change_records
- audit_logs
- notification_channels
- runbooks
- knowledge_documents

## 本地运行

### 启动后端

```bash
cd backend
go run main.go
```

默认后端地址：

```text
http://127.0.0.1:8080
```

数据库默认使用 MySQL，可通过环境变量覆盖：

```bash
export DB_DSN='user:password@tcp(host:3306)/aiops?charset=utf8mb4&parseTime=True&loc=Local'
go run main.go
```

### 启动前端

```bash
cd frontend
npm run dev
```

默认前端地址：

```text
http://localhost:5173
```

当前 Vite 已配置 `host: 0.0.0.0`，远程环境可通过机器 IP 访问 5173。

### 构建前端

```bash
cd frontend
npm run build
```

### 后端测试

```bash
cd backend
go test ./...
```

## 实施路线图

### 第一阶段：基础平台稳定

- 完善登录和权限
- 完善菜单和动态路由
- MySQL 稳定运行
- 接入中心稳定
- Prometheus 查询稳定
- 基础巡检稳定
- 操作记录和错误提示完善

### 第二阶段：可观测能力增强

- 查询历史
- 查询收藏
- PromQL 模板
- 日志查询
- Trace 查询
- Dashboard 分组
- 工作空间视角

### 第三阶段：证书轮转完整闭环

- 阿里云 CDN/DCDN 更新证书
- 证书库存
- 证书到期扫描
- 更新前证书校验
- 更新后线上验证
- 失败重试
- 轮转审批
- 到期通知
- 扩展腾讯云、华为云、Kubernetes、Nginx

### 第四阶段：告警治理

- 告警接入
- 告警列表
- 告警聚合
- 告警去重
- 告警静默
- 告警通知
- 告警关联服务和变更
- 告警处理记录

### 第五阶段：自动化工作流

- 工作流可视化
- 节点编排
- 条件分支
- 审批节点
- 回滚节点
- 定时触发
- 告警触发
- 证书轮转触发

### 第六阶段：AI 诊断

- PromQL 生成
- 日志摘要
- 告警摘要
- 根因候选
- Runbook 推荐
- 故障复盘初稿
- 历史故障匹配
- 自动生成排障步骤

### 第七阶段：资产、拓扑和变更

- 资产接入
- 服务拓扑
- 依赖关系
- 变更记录
- 变更影响分析
- 发布后自动巡检
- 故障影响面分析

## 安全原则

- AccessKey、Token、私钥默认不落库
- 必须落库的密钥需要加密存储
- 敏感字段展示必须脱敏
- 高危操作必须二次确认
- 高危操作需要审批能力
- 所有写操作需要审计
- 自动化动作必须有超时和回滚方案
- AI 只能给建议，高危动作不能直接自动执行

## 设计原则

- 先把运维链路跑通，再做智能化
- 先做单云单产品，再抽象多云
- 先做记录和审计，再做自动执行
- 先做明确规则，再引入 AI 判断
- 所有外部 API 调用都要有超时、错误提示和执行记录
- 所有自动化能力都必须支持人工确认和失败追踪

## 当前优先级

1. 证书轮转：完善阿里云 CDN/DCDN 更新证书
2. 证书管理：证书解析、私钥匹配、到期扫描
3. Prometheus 查询：查询历史、收藏、模板
4. 告警治理：先接入 Alertmanager
5. 巡检中心：增加 TLS、TCP、DNS 检查
6. 工作流：把证书更新、巡检、通知串起来
7. AI 助手：先做 PromQL 生成和告警摘要
