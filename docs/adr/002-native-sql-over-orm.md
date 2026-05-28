# ADR-002: 使用原生 SQL 而非 ORM

## 状态

已接受 (accepted)

## 背景

Go 生态中有多个 ORM 可选（GORM、Ent、Bun），但团队对 SQL 都比较熟悉，且项目需要用到 PostgreSQL 的 pgvector 扩展（向量检索），ORM 对这类扩展的支持往往滞后或不完善。

## 决策

使用 `database/sql` + `pgx` 驱动，手写原生 SQL。Repository 层负责 SQL 编写和结果映射，Service 层不接触 SQL。

## 影响

### 正面

- 完全控制 SQL，能充分利用 PostgreSQL 特性（CTE、窗口函数、pgvector）
- 没有 ORM 的"魔法"，性能问题更容易定位
- 团队成员 SQL 能力得到提升

### 负面

- 手写 SQL 增加了样板代码
- 没有 ORM 的自动迁移工具，需要手写迁移文件（golang-migrate）
- 需要手动处理字段映射（Scan）

## 备选方案

| 方案 | 不选原因 |
|------|----------|
| GORM | 对 pgvector 支持需要第三方插件，复杂查询生成效率低 |
| Ent | 强类型代码生成，学习曲线陡峭，与现有 Clean Architecture 分层融合成本高 |
| sqlc | 生成代码的灵活性不足，对动态查询支持有限；团队更偏好手写 SQL |
