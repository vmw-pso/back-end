package data

import (
	"database/sql"
	"errors"
)

var (
	ErrNotFound     = errors.New("record not found")
	ErrEditConflict = errors.New("edit conflict")
)

type Models struct {
	Resources ResourceModel
	Projects  ProjectModel
}

func NewModels(db *sql.DB) *Models {
	return &Models{
		Resources: ResourceModel{DB: db},
		Projects:  ProjectModel{DB: db},
	}
}
