package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"gitlab16.skiftrade.kz/templates/go/pkg/logger"
	"gitlab16.skiftrade.kz/templates/go/internal/repository/models"
)

func (r *Repository) ReadUser(ctx context.Context, id int64, dbTx pgx.Tx) (user models.User, err error) {
	const q = `
SELECT id, name, surname
FROM users
WHERE id = $1
LIMIT 1`

	qry := r.getQueryable(dbTx)
	err = qry.QueryRow(ctx, q, id).Scan(&user.ID, &user.Name, &user.Surname)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, models.ErrUserIsNotFound
		}
		slog.ErrorContext(ctx, "failed to read user from postgres", logger.ErrorAttr(err), logger.Int64Attr("user_id", id))
		return user, err
	}

	return user, nil
}
