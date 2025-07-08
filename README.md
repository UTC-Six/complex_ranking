# complex_ranking
一个比较复杂的排行榜维护方法

## 方案设计

- 使用 gozero 作为 API 框架，MySQL 存储举手/上台记录，Redis zSet 维护排行榜。
- 排名规则：
  1. 上台次数越少越靠前
  2. 举手次数越多越靠前
  3. 本次举手时间戳越早越靠前
- 触发时机：用户举手时，实时维护排行榜。
- 并发控制：支持并发安全，生产建议用分布式锁。
- 主要表结构：
  - kc_interaction_mic_stage（上台发言记录）
  - kc_interaction_mic_raise（举手记录）
- Redis zSet 的 score 计算方式：`score = stageCount * 1e10 - raiseCount * 1e5 + lastRaiseTime.Unix()`，分数越小排名越靠前。

## 目录结构
- internal/types：结构体定义
- internal/model：MySQL 查询与分数计算
- internal/logic：排行榜业务逻辑
- internal/handler：HTTP handler
- internal/svc：依赖注入
- api/ranking.api：API 定义

## 主要接口
- POST `/api/ranking/raiseHand` 用户举手，维护排行榜
- GET `/api/ranking/list?liveId=xxx` 获取排行榜

## 启动方法
1. 配置 MySQL、Redis 连接
2. 编译并运行服务
3. 按 API 文档调用接口
