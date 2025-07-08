package logic

import (
	"context"
	"fmt"
	"sync"

	"complex_ranking/internal/model"

	"github.com/go-redis/redis/v8"
)

// RankingLogic 负责排行榜的业务逻辑
// 支持并发安全
type RankingLogic struct {
	RedisClient  *redis.Client
	RankingModel *model.RankingModel
	mu           sync.Mutex // 简单并发控制，生产可用分布式锁
}

// OnUserRaiseHand 用户举手时触发，维护排行榜
func (l *RankingLogic) OnUserRaiseHand(ctx context.Context, liveID, userID int64) error {
	l.mu.Lock() // 简单并发控制，生产建议用分布式锁
	defer l.mu.Unlock()

	// 1. 查询该直播下所有用户的统计信息
	stats, err := l.RankingModel.GetUserStats(ctx, liveID)
	if err != nil {
		return fmt.Errorf("query user stats failed: %w", err)
	}

	// 2. 计算分数并批量写入 Redis zSet
	zsetKey := fmt.Sprintf("live:ranking:%d", liveID)
	var zMembers []*redis.Z
	for _, item := range stats {
		item.Score = model.CalcRankingScore(item)
		zMembers = append(zMembers, &redis.Z{
			Score:  item.Score,
			Member: item.UserID,
		})
	}
	// 清空并重建排行榜（可优化为只更新变动用户）
	pipe := l.RedisClient.TxPipeline()
	pipe.Del(ctx, zsetKey)
	if len(zMembers) > 0 {
		pipe.ZAdd(ctx, zsetKey, zMembers...)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("update redis zset failed: %w", err)
	}
	return nil
}
