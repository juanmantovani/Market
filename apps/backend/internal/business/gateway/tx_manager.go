package gateway

import "context"

// TxManager runs fn as one unit of work: every repository call made with the
// ctx fn receives either commits together or not at all. A use case that
// changes more than one aggregate (a sale, its stock and its movements) runs
// inside WithinTx; repositories never open transactions on their own.
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
