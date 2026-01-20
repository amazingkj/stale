package repository

import (
	"context"
	"time"

	"github.com/jiin/stale/internal/domain"
	"github.com/jiin/stale/internal/util"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type SourceRepository struct {
	db *sqlx.DB
}

func NewSourceRepository(db *sqlx.DB) *SourceRepository {
	return &SourceRepository{db: db}
}

func (r *SourceRepository) Create(ctx context.Context, input domain.SourceInput) (*domain.Source, error) {
	// Encrypt token before storing
	encryptedToken, err := util.Encrypt(input.Token)
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO sources (name, type, token, organization, url, repositories, scan_branch, insecure_skip_verify, membership_only, owner_only, created_at, updated_at)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
              RETURNING id, name, type, token, organization, url, repositories, scan_branch, insecure_skip_verify, membership_only, owner_only, created_at, updated_at, last_scan_at`

	now := time.Now()
	var source domain.Source
	err = r.db.GetContext(ctx, &source, query, input.Name, input.Type, encryptedToken, input.Organization, input.URL, input.Repositories, input.ScanBranch, input.InsecureSkipVerify, input.MembershipOnly, input.OwnerOnly, now, now)
	if err != nil {
		return nil, err
	}

	// Decrypt token for return value
	decrypted, err := util.Decrypt(source.Token)
	if err != nil {
		log.Warn().Err(err).Int64("source_id", source.ID).Msg("failed to decrypt token, using as-is")
	}
	source.Token = decrypted
	return &source, nil
}

func (r *SourceRepository) GetAll(ctx context.Context) ([]domain.Source, error) {
	var sources []domain.Source
	err := r.db.SelectContext(ctx, &sources, "SELECT * FROM sources ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}

	// Decrypt tokens
	for i := range sources {
		decrypted, err := util.Decrypt(sources[i].Token)
		if err != nil {
			log.Warn().Err(err).Int64("source_id", sources[i].ID).Msg("failed to decrypt token, using as-is")
		}
		sources[i].Token = decrypted
	}
	return sources, nil
}

func (r *SourceRepository) GetByID(ctx context.Context, id int64) (*domain.Source, error) {
	var source domain.Source
	err := r.db.GetContext(ctx, &source, "SELECT * FROM sources WHERE id = ?", id)
	if err != nil {
		return nil, err
	}

	// Decrypt token
	decrypted, err := util.Decrypt(source.Token)
	if err != nil {
		log.Warn().Err(err).Int64("source_id", source.ID).Msg("failed to decrypt token, using as-is")
	}
	source.Token = decrypted
	return &source, nil
}

func (r *SourceRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sources WHERE id = ?", id)
	return err
}

func (r *SourceRepository) UpdateLastScan(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "UPDATE sources SET last_scan_at = ?, updated_at = ? WHERE id = ?",
		time.Now(), time.Now(), id)
	return err
}

func (r *SourceRepository) Update(ctx context.Context, id int64, input domain.SourceInput) (*domain.Source, error) {
	// Encrypt token before storing
	encryptedToken, err := util.Encrypt(input.Token)
	if err != nil {
		return nil, err
	}

	query := `UPDATE sources SET name = ?, type = ?, token = ?, organization = ?, url = ?, repositories = ?, scan_branch = ?, insecure_skip_verify = ?, membership_only = ?, owner_only = ?, updated_at = ?
              WHERE id = ?
              RETURNING id, name, type, token, organization, url, repositories, scan_branch, insecure_skip_verify, membership_only, owner_only, created_at, updated_at, last_scan_at`

	var source domain.Source
	err = r.db.GetContext(ctx, &source, query, input.Name, input.Type, encryptedToken, input.Organization, input.URL, input.Repositories, input.ScanBranch, input.InsecureSkipVerify, input.MembershipOnly, input.OwnerOnly, time.Now(), id)
	if err != nil {
		return nil, err
	}

	// Decrypt token for return value
	decrypted, err := util.Decrypt(source.Token)
	if err != nil {
		log.Warn().Err(err).Int64("source_id", source.ID).Msg("failed to decrypt token, using as-is")
	}
	source.Token = decrypted
	return &source, nil
}
