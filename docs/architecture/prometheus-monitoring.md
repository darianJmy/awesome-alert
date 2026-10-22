# 接入中心与指标数据源

接入中心当前支持 Prometheus 和 VictoriaMetrics。两者复用 PromQL 查询接口，可用于即时查询、历史曲线、指标/标签发现、平台告警规则评估和 AI 指标查询。实例分组允许混合使用两种指标数据源。

## 配置入口

「接入中心 → 数据源实例 → 新增实例」选择产品类型，再配置名称、查询地址、认证、标签和状态。列表显示产品类型，支持按类型和状态筛选。已有实例不允许直接变更类型，避免改变规则引用的含义。

- Prometheus：`http://prometheus:9090`。
- VictoriaMetrics 单机：`http://victoriametrics:8428`。
- VictoriaMetrics 集群：`http://vmselect:8481/select/0/prometheus`，将 `0` 替换为实际租户 ID；不要使用 vminsert 写入地址。
- 反向代理地址可以保留路径前缀；平台在配置地址后追加 `/api/v1/...`。
- 认证支持无认证、Basic Auth、Bearer Token；高级配置支持自定义 HTTP 头。连接测试、即时查询、历史查询和规则评估复用同一 HTTP 客户端及 SSH 隧道能力。
- 编辑时留空保留同一认证方式的已有凭据，切换为无认证会清除旧凭据。
- 分组和规则引用的实例必须解除引用后才能删除。禁用实例不能查询或参与规则评估。

Prometheus 连接测试检查 `/api/v1/status/buildinfo`。VictoriaMetrics 连接测试通过实际查询地址执行 `vector(1)`，校验成功状态、向量结构和返回值；无需实现 Prometheus 的版本信息接口。该测试验证查询兼容性，不验证服务的产品身份。

## 扩展结构

`internal/handlers/datasource_registry.go` 是支持类型的唯一目录，定义：

- `type`：产品类型，例如 `prometheus`、`victoriametrics`。
- `capabilities`：数据能力，当前为 `metrics`，后续可扩展 `logs`、`traces`。
- `protocol`：查询协议，当前为 `promql`；不同产品可以复用协议实现。
- 连接测试适配器，以及前端需要的显示名称、地址示例和说明。

`GET /api/datasource-types` 返回目录。前端从目录生成类型选择和地址提示；创建/修改时后端从目录写入分类和能力，不信任客户端提交的能力标记。

接入管理使用 `supportedDatasourceTypes()`，PromQL 查询和规则评估使用 `metricDatasourceTypes()` / `isMetricDatasourceType()`。新增日志或链路类型时，需要注册真实适配器、实现对应查询接口及权限/工具入口；不能仅增加下拉选项或复用指标查询接口。现有 `/api/prometheus/*` 路径保留兼容，同时接受两种指标产品。

尚未支持的历史数据源记录继续保留，但不出现在可用接入列表中，也不能经接入 API 重新启用或执行连接测试。菜单统一显示「接入中心」，无需清空数据库。首页统计全部已支持的数据源，并保留旧 Prometheus 统计字段供旧客户端使用。

## 告警范围

本次支持从 VictoriaMetrics 查询指标并由平台执行告警规则。读取外部 vmalert 的当前告警属于独立的告警状态接口，不包含在本次指标接入中。

## 验证

后端测试覆盖两种类型的 CRUD、认证保留/清除、请求头、禁用状态、历史类型隔离、混合批量操作保护、分组与告警引用、AI 数据源选择，以及 VictoriaMetrics 单机和集群路径的连接测试、即时查询、历史查询、指标发现和规则查询。前端进行 Vue 编译检查。

接口参考：https://docs.victoriametrics.com/url-examples/index.html
