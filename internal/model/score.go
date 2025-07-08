package model

import (
	"complex_ranking/internal/types"
)

// CalcRankingScore 计算排行榜分数，分数越小排名越靠前
// 规则：上台次数越少优先，举手次数越多优先，举手时间越早优先
// 采用多级权重编码：score = stageCount * 1e13 - raiseCount * 1e8 + lastRaiseTime
func CalcRankingScore(item *types.RankingItem) float64 {
	// 1e13 保证上台次数优先，1e8 保证举手次数优先，毫秒时间戳保证最细粒度
	return float64(item.StageCount)*1e13 - float64(item.RaiseCount)*1e8 + float64(item.LastRaiseTime)
}
