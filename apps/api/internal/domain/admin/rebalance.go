package admin

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	RebalanceStrategyAuto = "auto"

	RebalanceStatusPending   = "pending"
	RebalanceStatusSubmitted = "submitted"
	RebalanceStatusCompleted = "completed"
	RebalanceStatusFailed    = "failed"
	RebalanceStatusDryRun    = "dry_run"
)

// AllocationDeltaProjection is a single source-level change previewed by dry_run.
type AllocationDeltaProjection struct {
	SourceID string `json:"source_id"`
	Delta    string `json:"delta"`
	Current  string `json:"current,omitempty"`
	Target   string `json:"target,omitempty"`
}

// RebalanceRequest is the JSON body for POST /api/v1/admin/vaults/{id}/rebalance.
type RebalanceRequest struct {
	Strategy string `json:"strategy"`
	DryRun   bool   `json:"dry_run"`
}

// RebalanceResponse is returned for both dry_run and live submissions.
type RebalanceResponse struct {
	Status                  string                      `json:"status"`
	TxHash                  string                      `json:"tx_hash,omitempty"`
	RebalanceID             uuid.UUID                   `json:"rebalance_id"`
	EstimatedCompletionMS   int64                       `json:"estimated_completion_ms,omitempty"`
	ProjectedDeltas         []AllocationDeltaProjection `json:"projected_deltas,omitempty"`
}

// VaultRebalanceRecord is the persisted audit row for a rebalance attempt.
type VaultRebalanceRecord struct {
	ID               uuid.UUID
	VaultID          uuid.UUID
	Strategy         string
	DryRun           bool
	Status           string
	TxHash           *string
	ProjectedDeltas  json.RawMessage
	ErrorMessage     *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
