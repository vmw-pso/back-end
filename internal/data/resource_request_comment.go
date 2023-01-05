package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/vmw-pso/back-end/internal/validator"
)

type ResourceRequestComment struct {
	ID                int64     `json:"id"`
	ResourceRequestID int64     `json:"requestID,omitempty"`
	Comment           string    `json:"comment"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt,omitempty"`
	Version           int64     `json:"version,omitempty"`
}

func ValidateComment(v *validator.Validator, comment string) {
	v.Check(comment != "", "comment", "must be provided")
}

type ResourceRequestCommentModel struct {
	DB *sql.DB
}

func (m *ResourceRequestCommentModel) Insert(c ResourceRequestComment) error {
	query := `
		INSERT INTO resource_request_comment
		(request_id, comment, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING comment_id, version`

	args := []interface{}{
		&c.ResourceRequestID,
		&c.Comment,
		&c.CreatedAt,
		&c.UpdatedAt,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&c.ID, &c.Version)
}

func (m *ResourceRequestCommentModel) Get(id int64) (*ResourceRequestComment, error) {
	query := `
		SELECT request_id, comment, created_at, updated_at, version
		FROM resource_request_comment
		WHERE comment_id=$1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var c ResourceRequestComment
	c.ID = id

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&c.ResourceRequestID,
		&c.Comment,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &c, nil
}

func (m *ResourceRequestCommentModel) Update(c *ResourceRequestComment) error {
	query := `
		UPDATE resource_request_comment
		SET request_id=$1, comment=$2, updated_at=$3, version=$4
		WHERE comment_id=$5
		RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	args := []interface{}{
		c.ResourceRequestID,
		c.Comment,
		c.UpdatedAt,
		c.Version + 1,
		c.ID,
	}

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&c.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m *ResourceRequestCommentModel) GetForRequest(reqId int64) ([]*ResourceRequestComment, error) {
	query := `
		SELECT comment_id, comment, created_at
		FROM resource_request_comment
		WHERE request_id=$1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, reqId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []*ResourceRequestComment{}

	for rows.Next() {
		var comment ResourceRequestComment
		err := rows.Scan(
			&comment.ID,
			&comment.Comment,
			&comment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, &comment)
	}

	return comments, nil
}
