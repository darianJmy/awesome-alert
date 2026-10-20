# 项目目录

本项目参照 `golang-standards/project-layout` 中的常用目录约定，保留实际使用的目录。

- `.air.toml`：本地开发热重载配置，通过 `make dev` 使用。
- `cmd/guanlan`：服务端入口；发布后运行 `./guanlan`。
- `cmd/aiops-agent`：独立 Agent 入口。
- `internal/app`：应用生命周期与 Wire 依赖装配。生成代码同目录保存。
- `internal/conf`：配置解析及 `conf/config.yaml` 内嵌默认值。
- `pkg/logger`：通用结构化日志，使用独立 `logger.Config` 配置。
- `pkg/database`：SQLite 连接、AutoMigrate、事务初始化、版本校验和 GORM 日志适配。通过模型列表和回调接入业务。
- `pkg/middleware`：通用 Gin 访问日志、异常恢复和可配置 CORS。
- `internal/database`：业务表清单、版本号和种子数据，调用公共数据库组件。
- `internal/middleware`：依赖用户、角色和审计模型的业务鉴权与审计。
- 其他 `internal` 包：API、业务模型、鉴权、AI、Agent 等私有实现。
- `web`：Vue 项目；`embed.go` 将 `dist` 嵌入 Go 二进制。
- `configs`：外部配置示例，默认运行无需复制该目录。
- `build/package`：容器等打包定义；Docker 构建上下文为仓库根目录。
- `deployments/migrations`：数据库部署说明和只供旧版使用的 MySQL SQL。
- `assets`：项目图片和图标。
- `docs`：设计与使用文档。
- `test/artifacts`：浏览器验证截图与辅助测试产物。

包测试保留在对应包中。公共包不引用 `internal`；配置与业务模型由应用层提供。没有脚本或部署清单的目录不创建空占位文件夹。

构建与验证在仓库根目录执行：

```bash
make build
./guanlan
# 本地开发：Air 热重载，缺少前端构建资源时自动准备
make dev
make wire
make test
```

最终发布只需要 `guanlan` 二进制。默认配置、前端资源和首次数据库初始化代码均包含在内；SQLite 数据和可选日志文件在运行时创建。
