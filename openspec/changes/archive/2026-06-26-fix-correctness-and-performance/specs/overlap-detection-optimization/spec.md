## ADDED Requirements

### Requirement: 合同重叠检测 SQL 优化
系统 SHALL 在 `ContractRepo` 接口中新增 `CheckOverlap(assetID, tenantID uint, start, end time.Time) (bool, error)` 方法。该方法 SHALL 使用 SQL WHERE 条件（`asset_id = ? AND tenant_id = ? AND status IN ('active','arrears') AND start_date < ? AND end_date > ?`）精确查询，SHALL NOT 将全部活跃合同加载到内存中。sqlite 和 postgres 两套仓库实现 SHALL 都实现该方法。

#### Scenario: 检测到合同重叠
- **WHEN** 存在资产 A 和租户 B 的活跃合同时段 2024-01-01 到 2024-12-31
- **AND** 创建新合同：资产 A、租户 B、时段 2024-06-01 到 2025-06-01
- **THEN** `CheckOverlap` 返回 true，handler 返回 409 冲突

#### Scenario: 无合同重叠
- **WHEN** 存在资产 A 和租户 B 的活跃合同时段 2024-01-01 到 2024-12-31
- **AND** 创建新合同：资产 A、租户 B、时段 2025-01-01 到 2025-12-31
- **THEN** `CheckOverlap` 返回 false，允许创建

#### Scenario: 不同资产无重叠
- **WHEN** 存在资产 A 和租户 B 的活跃合同
- **AND** 创建新合同：资产 C、租户 B、相同时段
- **THEN** `CheckOverlap` 返回 false，允许创建
