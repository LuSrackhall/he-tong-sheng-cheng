## ADDED Requirements

### Requirement: 收据创建事务原子性
打印收据时，序号分配（`AllocateSequence`）和收据记录创建（`Create`）SHALL 在同一个数据库事务中执行。事务失败时 SHALL 自动回滚，不产生序号空洞。

#### Scenario: 正常创建收据
- **WHEN** 为一个未打印的收款记录创建收据
- **THEN** 系统分配序号并创建收据记录，两者在同一事务中完成

#### Scenario: 收据记录创建失败
- **WHEN** 序号分配成功但收据记录写入失败
- **THEN** 事务回滚，序号不被消耗，不产生空洞
