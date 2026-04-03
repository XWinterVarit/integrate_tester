package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/model"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/repository"
)

const (
	featureLock = "LOCK"
	lockTTL     = 5 * time.Minute
)

type lockPayload struct {
	LockedBy     string `json:"locked_by"`
	LockedAt     string `json:"locked_at"`
	ExpiresAt    string `json:"expires_at"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
}

type LockService struct {
	adminRepo *repository.AppDataRepository
}

func NewLockService(adminRepo *repository.AppDataRepository) *LockService {
	return &LockService{adminRepo: adminRepo}
}

func lockKey(resourceType, resourceID string) string {
	return resourceType + ":" + resourceID
}

func (s *LockService) Acquire(ctx context.Context, req model.AcquireLockRequest) (*model.LockResponse, error) {
	key := lockKey(req.ResourceType, req.ResourceID)
	scopeClient := req.ScopeClient

	// Check existing lock
	existing, err := s.adminRepo.Get(ctx, featureLock, scopeClient, "", key)
	if err == nil {
		var lp lockPayload
		if json.Unmarshal([]byte(existing.Data), &lp) == nil {
			expires, _ := time.Parse(time.RFC3339, lp.ExpiresAt)
			if time.Now().Before(expires) && lp.LockedBy != req.SessionID {
				return nil, fmt.Errorf("LOCKED:This resource is currently being edited by another user. Please try again later.")
			}
		}
	}

	now := time.Now()
	lp := lockPayload{
		LockedBy:     req.SessionID,
		LockedAt:     now.Format(time.RFC3339),
		ExpiresAt:    now.Add(lockTTL).Format(time.RFC3339),
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
	}
	data, _ := json.Marshal(lp)
	if err := s.adminRepo.Upsert(ctx, featureLock, scopeClient, "", key, string(data)); err != nil {
		return nil, fmt.Errorf("acquire lock: %w", err)
	}

	return &model.LockResponse{
		Key:          key,
		LockedBy:     lp.LockedBy,
		LockedAt:     lp.LockedAt,
		ExpiresAt:    lp.ExpiresAt,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
	}, nil
}

func (s *LockService) Renew(ctx context.Context, key, scopeClient, sessionID string) (*model.LockResponse, error) {
	existing, err := s.adminRepo.Get(ctx, featureLock, scopeClient, "", key)
	if err != nil {
		return nil, fmt.Errorf("lock not found")
	}
	var lp lockPayload
	if err := json.Unmarshal([]byte(existing.Data), &lp); err != nil {
		return nil, fmt.Errorf("invalid lock data")
	}
	if lp.LockedBy != sessionID {
		return nil, fmt.Errorf("LOCKED:lock held by another session")
	}

	now := time.Now()
	lp.ExpiresAt = now.Add(lockTTL).Format(time.RFC3339)
	data, _ := json.Marshal(lp)
	if err := s.adminRepo.Upsert(ctx, featureLock, scopeClient, "", key, string(data)); err != nil {
		return nil, fmt.Errorf("renew lock: %w", err)
	}

	return &model.LockResponse{
		Key:          key,
		LockedBy:     lp.LockedBy,
		LockedAt:     lp.LockedAt,
		ExpiresAt:    lp.ExpiresAt,
		ResourceType: lp.ResourceType,
		ResourceID:   lp.ResourceID,
	}, nil
}

func (s *LockService) Release(ctx context.Context, key, scopeClient, sessionID string) error {
	existing, err := s.adminRepo.Get(ctx, featureLock, scopeClient, "", key)
	if err != nil {
		return nil // already released
	}
	var lp lockPayload
	if json.Unmarshal([]byte(existing.Data), &lp) == nil {
		if lp.LockedBy != sessionID {
			return fmt.Errorf("LOCKED:cannot release lock held by another session")
		}
	}
	_ = s.adminRepo.Delete(ctx, featureLock, scopeClient, "", key)
	return nil
}

func (s *LockService) GetLock(ctx context.Context, key, scopeClient string) (*model.LockResponse, error) {
	existing, err := s.adminRepo.Get(ctx, featureLock, scopeClient, "", key)
	if err != nil {
		return nil, fmt.Errorf("lock not found")
	}
	var lp lockPayload
	if err := json.Unmarshal([]byte(existing.Data), &lp); err != nil {
		return nil, fmt.Errorf("invalid lock data")
	}
	expires, _ := time.Parse(time.RFC3339, lp.ExpiresAt)
	if time.Now().After(expires) {
		return nil, fmt.Errorf("lock expired")
	}
	return &model.LockResponse{
		Key:          key,
		LockedBy:     lp.LockedBy,
		LockedAt:     lp.LockedAt,
		ExpiresAt:    lp.ExpiresAt,
		ResourceType: lp.ResourceType,
		ResourceID:   lp.ResourceID,
	}, nil
}
