# SOP Hub 服务端

SOP（标准作业指导书）系统的 Go 后端服务。基于 **Gin + GORM**，整合三类数据库，为 Web 后台与 APP 端提供 REST 接口与静态资源服务。

## 技术栈

| 类别 | 选型 |
| --- | --- |
| 语言 / HTTP | Go 1.25 / Gin v1.12 |
| ORM | GORM v1.31（MySQL / SQL Server 双驱动） |
| 鉴权 | JWT（golang-jwt/jwt/v5）Bearer Token |
| 配置 | Viper + fsnotify（支持热重载） |
| 日志 | Zap + Lumberjack（按日切割、压缩归档） |
| 其他 | excelize（Excel 导入导出）、decimal（金额精度） |

## 目录结构

```
sop-2.4/
├── main.go                  # 入口：初始化 → 路由 → 优雅关闭
├── config.yaml              # 运行配置（服务、三库、日志、JWT）
├── api.html                 # Web 端 API 接口文档（63 接口，含明暗主题）
├── app.html                 # APP 端 API 接口文档（31 接口，含明暗主题）
├── global/                  # 全局配置快照与数据库连接
├── internal/
│   ├── api/                 # 请求 DTO 与 Handler（路由 → service 的适配层）
│   ├── service/             # 业务逻辑与数据查询
│   ├── model/               # GORM 模型
│   ├── router/              # 路由注册（按模块拆分）
│   ├── middleware/          # JWT / CORS / 日志 / 恢复 / 超时豁免等
│   ├── config/              # 配置结构
│   ├── database/            # 数据库连接选择器（Onebe / SOPDB / Sparepart）
│   └── initialize/          # 配置、数据库、日志、Viper 初始化
├── pkg/
│   ├── errors/              # 业务错误码与错误包装
│   ├── jwt/                 # JWT 签发 / 解析
│   ├── logger/              # 日志与 GORM 日志适配
│   └── response/            # 统一响应信封与分页封装
├── uploads/                 # 上传文件（媒体、现场照片）
└── logs/                    # 运行日志
```

## 数据库

服务整合三个数据库，通过 `internal/database` 的选择器访问：

| 数据库 | 驱动 | 读写 | 说明 |
| --- | --- | --- | --- |
| **SOPHub** | SQL Server | 读写 | SOP 业务主库，业务表全部在此 |
| **Onebe** | MySQL | 只读 | 老库，查询维修位置 / 内容 / 问题类型等字典数据 |
| **SPAREPART** | SQL Server | 只读 | 备件老库，查询备件字典数据 |

> 强制约定：service 层只能通过 `database.Onebe()` / `SOPDB()` / `Sparepart()` 访问数据库，禁止直接引用全局连接；Onebe / SPAREPART 为只读库，代码层禁止写操作，数据库侧也需以最小权限只读账号保障。

## 快速开始

### 环境要求

- Go 1.25+
- 可访问的 SQL Server（SOPHub、SPAREPART）与 MySQL（Onebe）实例

### 配置

复制并修改 `config.yaml`（含三库连接串、端口、JWT 密钥等）：

> ⚠️ `config.yaml` 内敏感信息为占位/示例；生产环境必须更换 JWT secret、appTokenKey 及数据库口令，且 `server.mode` 置为 `release`、`server.skipPasswordCheck` 置为 `false`。

```bash
cp config.yaml config.local.yaml
# 按需编辑后启动时指定
```

### 启动

```bash
go run main.go
# 或构建后运行
go build -o sop.exe ./
./sop.exe
```

启动成功后可访问：

- 健康检查：`GET /health`（对 SOPHub / Onebe 双库 Ping，任一不可用返回 503）
- 静态资源：`/uploads/*`（视频、PDF、现场照片等，豁免服务器级读写超时）
- API：`/api/v1/*`

默认监听 `:8000`（由 `server.port` 控制，需与前端代理目标一致）。

## 主要特性

- **认证体系**：Web 登录签发 JWT（有效期 `jwt.expiresDays`）；APP 端通过 `appLogin` 借用独立密钥 `server.appTokenKey` 换取永久 AppToken（`Claims.Type="app"`）。
- **统一响应信封**：`{code, data, msg}`，`code=0` 成功 / `7` 业务失败（HTTP 200）；鉴权失败 401、无权限 403、未找到 404、系统错误 500。
- **统一分页信封**：`{list, total, page, pageSize, pagination{page, limit, total}}`；参数自动归一化（`page<=0→1`、`limit<=0→10`、超上限截断）。
- **三态筛选约定**：部分字典接口（如维修位置 / 内容 / 问题类型）支持按 `processId` 过滤，语义为三态——不传或 `null` 不过滤、`0` 查 `process_id IS NULL` 的记录、`>0` 精确匹配。
- **优雅关闭**：监听 SIGINT / SIGTERM，先排空在途请求再关闭三个数据库连接池。
- **超时防护**：全局读写超时兜底，媒体上传与 `/uploads` 静态资源由 `middleware.NoTimeout()` 豁免。
- **错误脱敏**：非业务错误原文只写服务端日志，客户端仅返回通用提示，不透出 SQL / 路径等敏感信息。

## 接口文档

- [api.html](api.html) — Web 端全量接口文档（63 接口 / 17 模块）
- [app.html](app.html) — APP 端接口文档（31 接口 / 15 模块）

两份文档均含请求/响应 JSON 样例与字段说明表格，支持暗 / 亮主题切换。
