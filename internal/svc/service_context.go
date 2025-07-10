package svc

import (
	"complex_ranking/internal/logic"
	"complex_ranking/internal/model"
	"database/sql"

	"github.com/go-redis/redis/v8"
)

type ServiceContext struct {
	RedisClient      *redis.Client
	DB               *sql.DB
	RankingModel     *model.RankingModel
	InteractionModel *model.InteractionModel
	RankingLogic     *logic.RankingLogic
}

func NewServiceContext(redisOpt *redis.Options, mysqlDSN string) (*ServiceContext, error) {
	db, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		return nil, err
	}
	redisClient := redis.NewClient(redisOpt)
	rankingModel := &model.RankingModel{DB: db}
	interactionModel := model.NewInteractionModel(db)
	rankingLogic := &logic.RankingLogic{
		RedisClient:  redisClient,
		RankingModel: rankingModel,
	}
	return &ServiceContext{
		RedisClient:      redisClient,
		DB:               db,
		RankingModel:     rankingModel,
		InteractionModel: interactionModel,
		RankingLogic:     rankingLogic,
	}, nil
}
