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
	Projects         ProjectModel
	Resources        ResourceModel
	ResourceRequests ResourceRequestModel
}

func NewModels(db *sql.DB) *Models {
	return &Models{
		Projects:         ProjectModel{DB: db},
		Resources:        ResourceModel{DB: db},
		ResourceRequests: ResourceRequestModel{DB: db},
	}
}
