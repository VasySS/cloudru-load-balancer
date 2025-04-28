// Package postgres provides methods for working with Postgres.
package postgres

import (
	"context"

	"github.com/VasySS/cloudru-load-balancer/internal/infrastructure/repository"
)

// Repository is a Postgres repository wrapper.
type Repository struct {
	txManager *TxManager
}

// New creates a new Postgres repository.
func New(txManager *TxManager) *Repository {
	return &Repository{
		txManager: txManager,
	}
}

// RunTx begins a new transaction with default isolation level.
func (r *Repository) RunTx(ctx context.Context, fn repository.TxFunc) error {
	return r.txManager.RunTx(ctx, fn)
}

// RunReadTx begins a new transaction with ReadOnly access mode.
func (r *Repository) RunReadTx(ctx context.Context, fn repository.TxFunc) error {
	return r.txManager.RunReadTx(ctx, fn)
}

// ReadUncommitted begins a new transaction with ReadUncommitted isolation level.
func (r *Repository) ReadUncommitted(ctx context.Context, fn repository.TxFunc) error {
	return r.txManager.ReadUncommitted(ctx, fn)
}

// RunReadCommitted begins a new transaction with ReadCommitted isolation level.
func (r *Repository) RunReadCommitted(ctx context.Context, fn repository.TxFunc) error {
	return r.txManager.RunReadCommitted(ctx, fn)
}

// RunRepeatableRead begins a new transaction with RepeatableRead isolation level.
func (r *Repository) RunRepeatableRead(ctx context.Context, fn repository.TxFunc) error {
	return r.txManager.RunRepeatableRead(ctx, fn)
}

// RunSerializable begins a new transaction with Serializable isolation level.
func (r *Repository) RunSerializable(ctx context.Context, fn repository.TxFunc) error {
	return r.txManager.RunSerializable(ctx, fn)
}
