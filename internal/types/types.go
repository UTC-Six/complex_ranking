package types

import "time"

// MicStageRecord 表示上台发言互动记录
type MicStageRecord struct {
	ID        int64     `db:"id"`
	LiveID    int64     `db:"live_id"`
	UserID    int64     `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

// MicRaiseRecord 表示举手互动记录
type MicRaiseRecord struct {
	ID        int64     `db:"id"`
	LiveID    int64     `db:"live_id"`
	UserID    int64     `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

// RankingItem 表示排行榜中的一项
// Score 由上台次数、举手次数、举手时间综合计算
// 用于 Redis zSet 的排序
// 具体排序规则见业务逻辑
type RankingItem struct {
	UserID        int64   `json:"user_id"`
	StageCount    int     `json:"stage_count"`
	RaiseCount    int     `json:"raise_count"`
	LastRaiseTime int64   `json:"last_raise_time"`
	Score         float64 `json:"score"`
}
