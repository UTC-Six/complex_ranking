package model

import (
	"context"
	"database/sql"

	"complex_ranking/internal/types"
)

type RankingModel struct {
	DB *sql.DB
}

// GetUserStatByUserID 获取某直播下指定用户的上台次数、举手次数、最近举手时间
func (m *RankingModel) GetUserStatByUserID(ctx context.Context, liveID, userID int64) (*types.RankingItem, error) {
	var raiseCount int
	var firstRaiseTime sql.NullTime
	err := m.DB.QueryRowContext(ctx, `
		SELECT COUNT(*), MIN(created_at)
		FROM kc_interaction_mic_raise
		WHERE live_id = ? AND user_id = ?
	`, liveID, userID).Scan(&raiseCount, &firstRaiseTime)
	if err != nil {
		return nil, err
	}
	var stageCount int
	err = m.DB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM kc_interaction_mic_stage
		WHERE live_id = ? AND user_id = ?
	`, liveID, userID).Scan(&stageCount)
	if err != nil {
		return nil, err
	}
	item := &types.RankingItem{
		UserID:     userID,
		StageCount: stageCount,
		RaiseCount: raiseCount,
	}
	if firstRaiseTime.Valid {
		item.LastRaiseTime = firstRaiseTime.Time.UnixMilli()
	}
	return item, nil
}
