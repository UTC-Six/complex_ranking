package handler

import (
	"net/http"
	"strconv"

	"complex_ranking/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type RaiseHandReq struct {
	ClassRoomId int64 `json:"classRoomId"`
	LiveId      int64 `json:"liveId"`
	UserId      int64 `json:"userId"`
}

type RaiseHandResp struct {
	Success bool `json:"success"`
}

type GetRankingReq struct {
	ClassRoomId int64 `json:"classRoomId"`
}

type RankingItem struct {
	UserId        int64   `json:"userId"`
	StageCount    int     `json:"stageCount"`
	RaiseCount    int     `json:"raiseCount"`
	LastRaiseTime int64   `json:"lastRaiseTime"`
	Score         float64 `json:"score"`
}

type GetRankingResp struct {
	List []RankingItem `json:"list"`
}

type CancelRaiseHandReq struct {
	ClassRoomId int64 `json:"classRoomId"`
	LiveId      int64 `json:"liveId"`
	UserId      int64 `json:"userId"`
}

type CancelRaiseHandResp struct {
	Success bool `json:"success"`
}

func RaiseHandHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RaiseHandReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}
		ctx := r.Context()
		err := svcCtx.RankingLogic.OnUserRaiseHand(ctx, req.ClassRoomId, req.LiveId, req.UserId)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		httpx.OkJson(w, &RaiseHandResp{Success: true})
	}
}

func GetRankingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		classRoomIdStr := r.URL.Query().Get("classRoomId")
		classRoomId, err := strconv.ParseInt(classRoomIdStr, 10, 64)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		ctx := r.Context()

		// 使用新的GetRanking方法
		items, err := svcCtx.RankingLogic.GetRanking(ctx, classRoomId)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		var resp GetRankingResp
		for _, item := range items {
			resp.List = append(resp.List, RankingItem{
				UserId:        item.UserID,
				StageCount:    item.StageCount,
				RaiseCount:    item.RaiseCount,
				LastRaiseTime: item.LastRaiseTime,
				Score:         item.Score,
			})
		}
		httpx.OkJson(w, &resp)
	}
}

func CancelRaiseHandHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CancelRaiseHandReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}
		ctx := r.Context()
		err := svcCtx.RankingLogic.OnUserCancelRaiseHand(ctx, req.ClassRoomId, req.LiveId, req.UserId)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		httpx.OkJson(w, &CancelRaiseHandResp{Success: true})
	}
}
