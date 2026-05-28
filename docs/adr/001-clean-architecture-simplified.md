# ADR-001: 采用 Clean Architecture 简化版

## 状态

已接受 (accepted)

## 背景

项目需要一套清晰的分层架构，既要保证业务逻辑不泄漏到框架层，又要避免过度设计。标准的 Clean Architecture 四层（Entities / Use Cases / Interface Adapters / Frameworks）对中小型项目来说分层过细，增加了不必要的抽象成本。

## 决策

采用 Clean Architecture 简化版三层结构：Handler -> Service -> Repository，Domain 作为最内层实体层。

具体规则：
- Domain 层不依赖任何外部框架
- Service 层依赖 Repository 接口而非具体实现
- Handler 层只负责 HTTP 协议转换，不包含业务逻辑
- 统一错误类型 `pkg/errors` 用于跨层传递业务错误

## 影响

### 正面

- 分层清晰，新人能在 30 分钟内理解代码流向
- Service 层可独立测试，通过 mock Repository 实现单元测试
- 更换数据库或 Web 框架时，只需要重写 Repository 或 Handler

### 负面

- 比直接写 Handler + SQL 多了两层抽象，简单 CRUD 场景下显得啰嗦
- 接口定义增加了初期代码量

## 备选方案

| 方案 | 不选原因 |
|------|----------|
| 标准四层 Clean Architecture | Use Case 层和 Interface Adapter 层在项目初期价值不大，增加了不必要的接口和目录 |
| MVC 模式 | Controller 直接操作数据库，业务逻辑和持久化耦合，难以测试 |
| 函数式直接风格 | 没有分层边界，代码量增大后维护困难 |
