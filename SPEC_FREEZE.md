# FOUNDATION 规格冻结表

- 档位：`compact_10`；形态：backend；目标容量：10 道独立题；主题经过禁止选题清单筛查，不属于电商、库存、OA、博客、问诊、预约或报表核心产品。
- 业务边界：教师申请、学院初审、企业实践、技术服务与授课证据、同行/企业评价、人事认证、续认/失效复核；教师、学院、企业导师、人事和管理员权限分离。
- 持久化：SQLite 真实 SQL；版本迁移 001；users、sessions、quotas、applications、practices、evidences、reviews、audit_events、idempotency_keys、worker_jobs 十张关联表，外键、唯一约束和业务索引齐全。
- 事务：申请提交在一个事务内锁定并更新 quota、写 applications 与 audit；失败回滚；重启重新打开同一数据库恢复状态。
- 状态机：draft → submitted → college_approved → evidence_review → peer_review → certified → renewal；拒绝和失效路径明确，非法转换拒绝。
- 并发：quota version 条件更新、practice version 乐观锁、SQLite busy timeout；并发测试使用固定输入并运行 race。
- context/worker：HTTP 到 service/repository 传递 context；worker 支持取消、领取、重试退避、永久失败和优雅停止。
- 错误传播：领域错误分类、HTTP 统一 JSON 错误码、请求 ID、panic recovery；审计记录操作者、对象、动作、结果和请求关联。
- HTTP/身份：登录、会话过期、退出撤销、教师与人事角色；`/healthz` 与 `/readyz`；入口为 `cmd/server`。
- Docker：Go 版本来自 go.mod（1.25.0），构建 `./cmd/server`，默认入口 `/app/dual-teacher`，已验证 linux/amd64 与 linux/arm64。
- 测试：领域、service、真实数据库迁移/事务/恢复、HTTP 契约、并发、worker、幂等、分页、时间边界和鉴权测试；目标测试 Go 代码不少于 1500 行。
- 运行时出题边界（只规划正确能力，不预埋缺陷）：状态转换与截止时间、quota 事务、practice 乐观锁、证据核验、评价门槛、审计原子性、幂等键生命周期、session 撤销/过期、worker 重试恢复、组合查询与分页。
- 规模门禁：非测试生产 Go ≥2000 行、≥20 文件、≥10 package；测试 Go ≥1500 行；禁止空壳、重复结构和注释凑数。
