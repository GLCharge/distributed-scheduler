package postgres

import (
	"database/sql"
	"github.com/GLCharge/otelzap"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func rollback(tx *sqlx.Tx, log *otelzap.Logger) {
	err := tx.Rollback()
	if err != sql.ErrTxDone && err != nil {
		log.Error("Failed to rollback transaction", zap.Error(err))
	}
}
