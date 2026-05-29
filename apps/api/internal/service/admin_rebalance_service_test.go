package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	admindomain "github.com/suncrestlabs/nester/apps/api/internal/domain/admin"
	"github.com/suncrestlabs/nester/apps/api/internal/domain/vault"
	"github.com/suncrestlabs/nester/apps/api/internal/repository/postgres"
	"github.com/suncrestlabs/nester/apps/api/internal/service"
)

type rebalanceAdminRepo struct {
	detail      admindomain.VaultDetail
	inFlight    bool
	records     []admindomain.VaultRebalanceRecord
	createFails bool
}

func (r *rebalanceAdminRepo) GetVaultHealthDashboard(context.Context) (admindomain.VaultHealthDashboardData, error) {
	return admindomain.VaultHealthDashboardData{}, nil
}
func (r *rebalanceAdminRepo) ListVaults(context.Context, admindomain.VaultListFilter) ([]admindomain.VaultSummary, int, error) {
	return nil, 0, nil
}
func (r *rebalanceAdminRepo) GetVaultDetail(context.Context, uuid.UUID) (admindomain.VaultDetail, error) {
	return r.detail, nil
}
func (r *rebalanceAdminRepo) UpdateVaultStatus(context.Context, uuid.UUID, vault.VaultStatus) (admindomain.VaultDetail, error) {
	return r.detail, nil
}
func (r *rebalanceAdminRepo) ListSettlements(context.Context, admindomain.SettlementListFilter) ([]admindomain.SettlementSummary, int, error) {
	return nil, 0, nil
}
func (r *rebalanceAdminRepo) ListUsers(context.Context, admindomain.UserListFilter) ([]admindomain.UserSummary, int, error) {
	return nil, 0, nil
}
func (r *rebalanceAdminRepo) GetLastEventIndexedAt(context.Context) (*time.Time, error) {
	return nil, nil
}
func (r *rebalanceAdminRepo) DatabaseHealth(context.Context) (int64, error) {
	return 1, nil
}
func (r *rebalanceAdminRepo) HasInFlightRebalance(context.Context, uuid.UUID) (bool, error) {
	return r.inFlight, nil
}
func (r *rebalanceAdminRepo) CreateVaultRebalance(_ context.Context, record admindomain.VaultRebalanceRecord) (admindomain.VaultRebalanceRecord, error) {
	if r.createFails {
		return admindomain.VaultRebalanceRecord{}, postgres.ErrRebalanceInFlight
	}
	if record.ID == uuid.Nil {
		record.ID = uuid.New()
	}
	r.records = append(r.records, record)
	return record, nil
}
func (r *rebalanceAdminRepo) UpdateVaultRebalance(_ context.Context, record admindomain.VaultRebalanceRecord) (admindomain.VaultRebalanceRecord, error) {
	for i := range r.records {
		if r.records[i].ID == record.ID {
			r.records[i] = record
			return record, nil
		}
	}
	r.records = append(r.records, record)
	return record, nil
}

type rebalanceChainInvoker struct {
	simulateErr error
	submitHash  string
	submitErr   error
}

func (c rebalanceChainInvoker) PauseVault(context.Context, string) error   { return nil }
func (c rebalanceChainInvoker) UnpauseVault(context.Context, string) error { return nil }
func (c rebalanceChainInvoker) RebalanceVault(context.Context, string) (string, error) {
	return c.submitHash, c.submitErr
}
func (c rebalanceChainInvoker) SimulateRebalanceVault(context.Context, string) error {
	return c.simulateErr
}

func TestAdminService_TriggerRebalance_DryRun(t *testing.T) {
	vaultID := uuid.New()
	repo := &rebalanceAdminRepo{
		detail: admindomain.VaultDetail{
			VaultSummary: admindomain.VaultSummary{
				ID:              vaultID,
				ContractAddress: "CVAULT",
				CurrentBalance:  decimal.RequireFromString("1000"),
				Status:          vault.StatusActive,
			},
			Allocations: []vault.Allocation{
				{Protocol: "aave", Amount: decimal.RequireFromString("1000")},
			},
		},
	}
	svc := service.NewAdminService(repo, rebalanceChainInvoker{}, "", "")

	resp, err := svc.TriggerRebalance(context.Background(), vaultID, admindomain.RebalanceRequest{
		Strategy: "auto",
		DryRun:   true,
	})
	if err != nil {
		t.Fatalf("TriggerRebalance() error = %v", err)
	}
	if resp.Status != "dry_run" {
		t.Fatalf("status = %q, want dry_run", resp.Status)
	}
	if resp.RebalanceID == uuid.Nil {
		t.Fatal("expected rebalance_id")
	}
}

func TestAdminService_TriggerRebalance_Submit(t *testing.T) {
	vaultID := uuid.New()
	repo := &rebalanceAdminRepo{
		detail: admindomain.VaultDetail{
			VaultSummary: admindomain.VaultSummary{
				ID:              vaultID,
				ContractAddress: "CVAULT",
				CurrentBalance:  decimal.RequireFromString("1000"),
				Status:          vault.StatusActive,
			},
		},
	}
	svc := service.NewAdminService(repo, rebalanceChainInvoker{submitHash: "abc123"}, "", "")

	resp, err := svc.TriggerRebalance(context.Background(), vaultID, admindomain.RebalanceRequest{
		Strategy: "auto",
		DryRun:   false,
	})
	if err != nil {
		t.Fatalf("TriggerRebalance() error = %v", err)
	}
	if resp.Status != "submitted" || resp.TxHash != "abc123" {
		t.Fatalf("resp = %+v, want submitted with hash abc123", resp)
	}
	if resp.EstimatedCompletionMS != 5000 {
		t.Fatalf("estimated_completion_ms = %d, want 5000", resp.EstimatedCompletionMS)
	}
}

func TestAdminService_TriggerRebalance_InFlight(t *testing.T) {
	vaultID := uuid.New()
	repo := &rebalanceAdminRepo{
		inFlight: true,
		detail: admindomain.VaultDetail{
			VaultSummary: admindomain.VaultSummary{
				ID:     vaultID,
				Status: vault.StatusActive,
			},
		},
	}
	svc := service.NewAdminService(repo, rebalanceChainInvoker{}, "", "")

	_, err := svc.TriggerRebalance(context.Background(), vaultID, admindomain.RebalanceRequest{DryRun: true})
	if !errors.Is(err, service.ErrRebalanceInFlight) {
		t.Fatalf("error = %v, want ErrRebalanceInFlight", err)
	}
}
