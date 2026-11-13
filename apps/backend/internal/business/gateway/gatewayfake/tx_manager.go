package gatewayfake

import "context"

// TxManager is a fake gateway.TxManager: it runs fn right away and counts
// the units of work. With Err set it fails before running fn, like a
// database that can't open a transaction.
type TxManager struct {
	Err   error
	Calls int
}

func (fake *TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	fake.Calls++
	if fake.Err != nil {
		return fake.Err
	}

	return fn(ctx)
}
