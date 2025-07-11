package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"complex_ranking/internal/model"
	"complex_ranking/internal/types"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// RankingLogic 负责排行榜的业务逻辑
// 支持并发安全
// 分布式锁实现
type RankingLogic struct {
	RedisClient  *redis.Client
	RankingModel *model.RankingModel
	// mu           sync.Mutex // 移除本地锁，改用分布式锁
}

// OnUserRaiseHand 用户举手时触发，维护排行榜
func (l *RankingLogic) OnUserRaiseHand(ctx context.Context, classRoomId, liveID, userID int64) error {
	lockKey := fmt.Sprintf("lock:classroom:ranking:%d", classRoomId)
	lockValue := uuid.NewString()
	locked, err := l.RedisClient.SetNX(ctx, lockKey, lockValue, 5*time.Second).Result()
	if err != nil {
		return fmt.Errorf("acquire redis lock failed: %w", err)
	}
	if !locked {
		return errors.New("another operation is in progress, please try again later")
	}
	defer func() {
		// 使用 pipeline 一次性执行 Get 和 Del 操作
		pipe := l.RedisClient.Pipeline()
		getCmd := pipe.Get(ctx, lockKey)
		pipe.Del(ctx, lockKey)
		_, err := pipe.Exec(ctx)
		if err == nil {
			val, _ := getCmd.Result()
			if val != lockValue {
				// 如果不是自己的锁，重新删除
				l.RedisClient.Del(ctx, lockKey)
			}
		}
	}()

	// 只查询当前用户的统计信息
	item, err := l.RankingModel.GetUserStatByUserID(ctx, liveID, userID)
	if err != nil {
		return fmt.Errorf("query user stats failed: %w", err)
	}
	item.Score = model.CalcRankingScore(item)
	zsetKey := fmt.Sprintf("classroom:ranking:%d", classRoomId)

	// 使用 pipeline 一次性执行 ZAdd 和 Expire 操作
	pipe := l.RedisClient.Pipeline()
	pipe.ZAdd(ctx, zsetKey, &redis.Z{
		Score:  item.Score,
		Member: item.UserID,
	})
	pipe.Expire(ctx, zsetKey, 24*time.Hour)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("update redis zset and set expiry failed: %w", err)
	}

	return nil
}

// OnUserCancelRaiseHand 用户撤销举手时，移除排行榜中的该用户
func (l *RankingLogic) OnUserCancelRaiseHand(ctx context.Context, classRoomId, liveID, userID int64) error {
	lockKey := fmt.Sprintf("lock:classroom:ranking:%d", classRoomId)
	lockValue := uuid.NewString()
	locked, err := l.RedisClient.SetNX(ctx, lockKey, lockValue, 5*time.Second).Result()
	if err != nil {
		return fmt.Errorf("acquire redis lock failed: %w", err)
	}
	if !locked {
		return errors.New("another operation is in progress, please try again later")
	}
	defer func() {
		// 使用 pipeline 一次性执行 Get 和 Del 操作
		pipe := l.RedisClient.Pipeline()
		getCmd := pipe.Get(ctx, lockKey)
		pipe.Del(ctx, lockKey)
		_, err := pipe.Exec(ctx)
		if err == nil {
			val, _ := getCmd.Result()
			if val != lockValue {
				// 如果不是自己的锁，重新删除
				l.RedisClient.Del(ctx, lockKey)
			}
		}
	}()

	zsetKey := fmt.Sprintf("classroom:ranking:%d", classRoomId)
	err = l.RedisClient.ZRem(ctx, zsetKey, userID).Err()
	if err != nil {
		return fmt.Errorf("remove user from redis zset failed: %w", err)
	}
	return nil
}

// GetRanking 获取排行榜
func (l *RankingLogic) GetRanking(ctx context.Context, classRoomId int64) ([]*types.RankingItem, error) {
	zsetKey := fmt.Sprintf("classroom:ranking:%d", classRoomId)

	// 获取排行榜前100名，按分数升序排列（分数越小排名越靠前）
	result, err := l.RedisClient.ZRangeWithScores(ctx, zsetKey, 0, 99).Result()
	if err != nil {
		return nil, fmt.Errorf("get ranking failed: %w", err)
	}

	var items []*types.RankingItem
	for _, z := range result {
		userID := z.Member.(string)
		// 这里需要将string转换为int64，实际使用时需要根据你的userID类型调整
		// 假设userID是数字字符串
		userIDInt, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			continue
		}

		item := &types.RankingItem{
			UserID: userIDInt,
			Score:  z.Score,
		}
		items = append(items, item)
	}

	return items, nil
}

// StartInteractionWithLock 业务逻辑：保证同 classroomId 下只有一个 status=2 的记录
func (l *RankingLogic) StartInteractionWithLock(ctx context.Context, classroomID, liveID int64) error {
	tx, err := l.RankingModel.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// 悲观锁查找进行中
	row := tx.QueryRowContext(ctx, "SELECT id FROM interaction_record WHERE classroom_id=? AND status=2 FOR UPDATE", classroomID)
	var id int64
	err = row.Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil {
		return errors.New("当前教室已有进行中的互动")
	}

	// 插入新记录
	_, err = tx.ExecContext(ctx, "INSERT INTO interaction_record (classroom_id, live_id, status, created_at) VALUES (?, ?, 2, ?)",
		classroomID, liveID, time.Now())
	return err
}
