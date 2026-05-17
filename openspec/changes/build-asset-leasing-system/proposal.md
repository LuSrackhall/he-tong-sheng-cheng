## Why

社区/村委商铺租赁管理目前依赖纸质记录或 Excel，工作人员需手动跟踪缴费状态、判断催缴时机，易出错、不可追溯。需要一个工具在后台自动完成记录、计算和提醒，不改变人的工作习惯。本系统以合同为核心锚点，实现签合同→收租金→催欠款的完整闭环，第一切入场景是商铺租赁，架构预留多资产类型扩展。

## What Changes

**新建系统：通用资产租赁与催收管理系统**

- 从零搭建完整的前后端系统
- 三大日常操作入口：签新合同（选资产→录租户→定合同→打印）、收租金（搜索→收款→打印收据）、看催缴清单（五级 Tab 自动分级）
- 四大后台模块：资产管理、租户管理、合同管理（含模板映射）、催缴管理
- 后端 PDF 生成（合同、身份证复印件、三联收据）
- 双部署模式：SQLite 本地单机 + PostgreSQL 服务端多人

## Capabilities

### New Capabilities

- `asset-management`: 资产台账管理，以名称和月租金为核心字段，支持 ExtraFields 扩展和分组/标签
- `tenant-management`: 租户档案管理，身份证信息录入，历史租赁记录可追溯，支持 ExtraFields 扩展
- `contract-management`: 合同全生命周期管理（执行中/已缴全/欠费中/已到期），模板上传与字段映射，作为系统唯一锚点
- `rent-collection`: 收款记账，每笔收款关联合同，自动更新已收总额和钱用到哪天，收据打印
- `arrears-collection`: 五级递进催缴清单自动生成，按紧急程度分级附带建议动作，催缴历史可追溯
- `system-printing`: 后端 PDF 生成（合同替换模板占位符、身份证复印件、三联收据连号管理）
- `user-auth`: JWT 认证与角色分权（admin/operator），用户管理

### Modified Capabilities

- 无。全新系统，不涉及现有能力变更。

## Impact

- 新建仓库 `/Users/srackhalllu/Desktop/资源管理器/safe/he-tong-sheng-cheng/`
- Go 后端分层架构：`cmd/server`、`internal/transport`、`internal/service`、`internal/domain`、`internal/repository`
- Vue 3 前端：`web/` 目录，10 个页面视图
- 数据库：SQLite 本地（自动创建 `data/` 目录）+ PostgreSQL 服务端
- API 约 30+ 端点，统一响应格式 `{code, data, message}`
- 部署：Go 内嵌 Vue 静态文件，单一二进制分发
- 无外部服务依赖（OCR 预留接口，本次不实现）
