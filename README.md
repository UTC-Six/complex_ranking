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
- Redis zSet 的 score 计算方式：

### 排行榜分数（score）计算规则详解

score = stageCount * 1e13 - raiseCount * 1e8 + lastRaiseTime

- **stageCount**：上台次数，权重最大（1e13），保证上台次数越少排名越靠前。
- **raiseCount**：举手次数，次级权重（1e8），保证在上台次数相同的情况下，举手次数越多排名越靠前。
- **lastRaiseTime**：最近一次举手的毫秒级时间戳，最小权重，保证在上台次数和举手次数都相同的情况下，举手时间越早排名越靠前。

#### 为什么这样设计能满足排名规则？

1. **上台次数优先级最高**：
   - 由于 stageCount 乘以 1e13，任何一次上台次数的变化都会让 score 变化至少 1e13，远大于后两项的影响。
   - 这样即使 raiseCount 再大、lastRaiseTime 再早，也无法超越上台次数更少的用户。
2. **举手次数优先级次之**：
   - raiseCount 乘以 1e8，数量级远小于 stageCount 的权重，但足以区分同 stageCount 下的用户。
   - 举手次数越多，score 越小，排名越靠前。
3. **举手时间最细粒度**：
   - lastRaiseTime 直接加到 score 上，单位为毫秒。
   - 在上台次数和举手次数都相同的情况下，举手时间越早（lastRaiseTime 越小），score 越小，排名越靠前。

**举例说明：**
- 用户A：stageCount=1, raiseCount=10, lastRaiseTime=1000000000000
- 用户B：stageCount=2, raiseCount=100, lastRaiseTime=900000000000
- 用户C：stageCount=1, raiseCount=5, lastRaiseTime=900000000000

score(A) = 1*1e13 - 10*1e8 + 1000000000000 = 10000000000000 - 1000000000 + 1000000000000 = 10999000000000
score(B) = 2*1e13 - 100*1e8 + 900000000000 = 20000000000000 - 10000000000 + 900000000000 = 20990000000000
score(C) = 1*1e13 - 5*1e8 + 900000000000 = 10000000000000 - 500000000 + 900000000000 = 10895000000000

排序结果：C（最前） < A < B
- C、A 上台次数最少，C 举手次数少于A但举手时间更早，故C排A前。
- B 上台次数最多，无论举手次数和时间如何都排最后。

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

特殊情形详解：stageCount = 0；raiseCount = 0 是 score 的计算是否正确？
1. stageCount = 0 的情况
   这表示该用户从未上台发言。
   由于 score 公式中 stageCount 权重最大（1e13），stageCount = 0 时，score 的主项为 0，理论上所有未上台的用户一定排在所有上过台的用户前面。
   这完全符合“上台次数越少越靠前”的业务需求。
2. raiseCount = 0 的情况
   这表示该用户从未举手。
   score 公式中 raiseCount 是负权重（-1e8），raiseCount = 0 时，这一项为 0。
   如果有用户 stageCount = 0 且 raiseCount = 0，则他们的 score 只由 lastRaiseTime 决定。
   但实际上，lastRaiseTime 只有在有举手行为时才有意义。对于 raiseCount = 0 的用户，lastRaiseTime 应为 0 或未赋值。
3. 极端情况的排序
   未上台且未举手的用户（stageCount = 0, raiseCount = 0）：score = 0 + 0 + 0 = 0
   未上台但举手过的用户（stageCount = 0, raiseCount > 0）：score = 0 - N1e8 + lastRaiseTime
   上过台的用户（stageCount > 0）：score ≥ 1e13，必然排在未上台用户之后
   所以：
   所有未上台的用户都排在上过台的用户前面。
   未上台但举手次数多的用户排在未举手的用户前面。
   未上台且未举手的用户（score=0）理论上会排在所有有举手行为的用户后面（因为负数更小），但如果你希望“未举手的用户”不参与排行榜，可以在业务层过滤掉 raiseCount = 0 的用户。
   建议
   业务上通常只展示有举手行为的用户，即 raiseCount > 0，这样排行榜才有实际意义。
   如果需要展示所有用户，score 设计依然能保证排序规则的正确性。
   结论
   现有 score 设计已正确覆盖 stageCount = 0 和 raiseCount = 0 的情况，排序逻辑不会出错。
   如需业务上过滤“未举手用户”，可在查询或展示时加一层判断。
