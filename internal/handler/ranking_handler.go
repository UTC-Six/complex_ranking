package handler

import (
	"net/http"
	"strconv"

	"complex_ranking/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type RaiseHandReq struct {
	LiveId int64 `json:"liveId"`
	UserId int64 `json:"userId"`
}

type RaiseHandResp struct {
	Success bool `json:"success"`
}

type GetRankingReq struct {
	LiveId int64 `json:"liveId"`
}

type RankingItem struct {
	UserId        int64 `json:"userId"`
	StageCount    int   `json:"stageCount"`
	RaiseCount    int   `json:"raiseCount"`
	LastRaiseTime int64 `json:"lastRaiseTime"`
}

type GetRankingResp struct {
	List []RankingItem `json:"list"`
}

func RaiseHandHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RaiseHandReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}
		ctx := r.Context()
		err := svcCtx.RankingLogic.OnUserRaiseHand(ctx, req.LiveId, req.UserId)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		httpx.OkJson(w, &RaiseHandResp{Success: true})
	}
}

func GetRankingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		liveIdStr := r.URL.Query().Get("liveId")
		_, err := strconv.ParseInt(liveIdStr, 10, 64)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		ctx := r.Context()
		zsetKey := "live:ranking:" + liveIdStr
		// 获取前 100 名
		items, err := svcCtx.RedisClient.ZRangeWithScores(ctx, zsetKey, 0, 99).Result()
		if err != nil {
			httpx.Error(w, err)
			return
		}
		var resp GetRankingResp
		for _, z := range items {
			userId, _ := z.Member.(int64)
			// 这里可扩展为从 DB 查询详细信息
			resp.List = append(resp.List, RankingItem{
				UserId: userId,
				// 这里只返回 userId，详细信息可扩展
			})
		}
		httpx.OkJson(w, &resp)
	}
}
