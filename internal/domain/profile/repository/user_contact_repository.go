package repository

import (
	"context"
	"database/sql"

	"lending-hub-service/internal/domain/profile/entity"
)

type UserContactRepository struct {
	db *sql.DB
}

func NewUserContactRepository(db *sql.DB) *UserContactRepository {
	return &UserContactRepository{db: db}
}

func (r *UserContactRepository) GetByUserID(ctx context.Context, userID string) (*entity.UserContact, error) {
	query := `
		SELECT id, user_id, mobile, email, raw_phone, source, created_at, updated_at
		FROM lender_user_profile
		WHERE user_id = $1
	`
	var uc entity.UserContact
	var email sql.NullString

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&uc.ID, &uc.UserID, &uc.Mobile, &email,
		&uc.RawPhone, &uc.Source, &uc.CreatedAt, &uc.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // not found, not an error
	}
	if err != nil {
		return nil, err
	}

	if email.Valid {
		uc.Email = email.String
	}

	return &uc, nil
}

func (r *UserContactRepository) Upsert(ctx context.Context, uc *entity.UserContact) error {
	query := `
		INSERT INTO lender_user_profile (user_id, mobile, email, raw_phone, source)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id)
		DO UPDATE SET
			mobile = EXCLUDED.mobile,
			email = EXCLUDED.email,
			raw_phone = EXCLUDED.raw_phone,
			source = EXCLUDED.source,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		uc.UserID, uc.Mobile, uc.Email, uc.RawPhone, uc.Source,
	).Scan(&uc.ID, &uc.CreatedAt, &uc.UpdatedAt)
}
