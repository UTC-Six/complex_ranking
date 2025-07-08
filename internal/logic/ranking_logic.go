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

	// 只查询当前用户的统计信息
	item, err := l.RankingModel.GetUserStatByUserID(ctx, liveID, userID)
	if err != nil {
		return fmt.Errorf("query user stats failed: %w", err)
	}
	item.Score = model.CalcRankingScore(item)
	zsetKey := fmt.Sprintf("live:ranking:%d", liveID)
	// 只更新当前用户的分数
	err = l.RedisClient.ZAdd(ctx, zsetKey, &redis.Z{
		Score:  item.Score,
		Member: item.UserID,
	}).Err()
	if err != nil {
		return fmt.Errorf("update redis zset failed: %w", err)
	}
	return nil
}
