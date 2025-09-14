package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aouxes/uptime-monitor/internal/models"
)

func (s *Storage) CreateSite(ctx context.Context, site *models.Site) error {
	query := `
        INSERT INTO sites (url, user_id, last_status, last_checked)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at
    `

	err := s.db.QueryRow(ctx, query,
		site.URL,
		site.UserID,
		"UNKNOWN",
		time.Now(),
	).Scan(&site.ID, &site.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create site: %w", err)
	}

	log.Printf("Site created successfully: ID=%d, URL=%s, UserID=%d",
		site.ID, site.URL, site.UserID)
	return nil
}

func (s *Storage) GetUserSites(ctx context.Context, userID int) ([]models.Site, error) {
	query := `
        SELECT id, url, user_id, last_status, last_checked, created_at
        FROM sites 
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sites: %w", err)
	}
	defer rows.Close()

	var sites []models.Site
	for rows.Next() {
		var site models.Site
		err := rows.Scan(
			&site.ID,
			&site.URL,
			&site.UserID,
			&site.LastStatus,
			&site.LastChecked,
			&site.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan site: %w", err)
		}
		sites = append(sites, site)
	}

	return sites, nil
}

func (s *Storage) UpdateSiteStatus(ctx context.Context, siteID int, status string) error {
	query := `
        UPDATE sites 
        SET last_status = $1, last_checked = $2
        WHERE id = $3
    `

	_, err := s.db.Exec(ctx, query, status, time.Now(), siteID)
	if err != nil {
		return fmt.Errorf("failed to update site status: %w", err)
	}

	return nil
}

func (s *Storage) DeleteSite(ctx context.Context, siteID, userID int) error {
	query := `DELETE FROM sites WHERE id = $1 AND user_id = $2`

	result, err := s.db.Exec(ctx, query, siteID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete site: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("site not found or access denied")
	}

	log.Printf("Site deleted: ID=%d, UserID=%d", siteID, userID)
	return nil
}
