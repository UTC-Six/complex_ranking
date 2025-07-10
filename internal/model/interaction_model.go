package model

import (
	"complex_ranking/internal/types"
	"database/sql"
)

type InteractionModel struct {
	DB *sql.DB
}

func NewInteractionModel(db *sql.DB) *InteractionModel {
	return &InteractionModel{DB: db}
}

// Insert 插入一条互动记录
func (m *InteractionModel) Insert(record *types.InteractionRecord) error {
	_, err := m.DB.Exec(
		"INSERT INTO interaction_record (classroom_id, live_id, status, created_at) VALUES (?, ?, ?, ?)",
		record.ClassroomID, record.LiveID, record.Status, record.CreatedAt,
	)
	return err
}

// GetOngoingByClassroomID 查询某 classroomId 下 status=2 的进行中互动
func (m *InteractionModel) GetOngoingByClassroomID(classroomID int64) (*types.InteractionRecord, error) {
	var record types.InteractionRecord
	err := m.DB.QueryRow(
		"SELECT id, classroom_id, live_id, status, created_at FROM interaction_record WHERE classroom_id=? AND status=2 LIMIT 1",
		classroomID,
	).Scan(&record.ID, &record.ClassroomID, &record.LiveID, &record.Status, &record.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}
