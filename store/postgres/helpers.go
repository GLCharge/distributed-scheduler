package postgres

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func rollback(tx *sqlx.Tx, log *zap.SugaredLogger) {
	err := tx.Rollback()
	if err != sql.ErrTxDone && err != nil {
		log.Errorw("Failed to rollback transaction", "error", err)
	}
}
