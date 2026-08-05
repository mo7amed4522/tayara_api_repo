package mongo

import (
	"context"
	"errors"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	colEvents = "payment_events"
	colStates = "payment_states"
)

// LedgerRepo persists the payment ledger (events + state) in MongoDB via the qmgo client.
type LedgerRepo struct{}

// db returns the configured payment database (guarded: callers check GVA_MONGO first).
func (LedgerRepo) db() *qmgo.Database {
	return global.GVA_MONGO.Database
}

// AppendEvent writes an append-only lifecycle event (idempotent by EventID).
func (r LedgerRepo) AppendEvent(ctx context.Context, ev *LedgerEvent) error {
	if global.GVA_MONGO == nil {
		return errors.New("mongodb not initialized")
	}
	_, err := r.db().Collection(colEvents).InsertOne(ctx, ev)
	return err
}

// UpsertState writes the current lifecycle state (idempotent by PaymentID).
func (r LedgerRepo) UpsertState(ctx context.Context, st *PaymentState) error {
	if global.GVA_MONGO == nil {
		return errors.New("mongodb not initialized")
	}
	_, err := r.db().Collection(colStates).UpsertId(ctx, st.PaymentID, st)
	return err
}

// GetState returns the current lifecycle state for a payment.
func (r LedgerRepo) GetState(ctx context.Context, paymentID string) (*PaymentState, error) {
	if global.GVA_MONGO == nil {
		return nil, errors.New("mongodb not initialized")
	}
	var st PaymentState
	if err := r.db().Collection(colStates).Find(ctx, bson.M{"_id": paymentID}).One(&st); err != nil {
		return nil, err
	}
	return &st, nil
}
