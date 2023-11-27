package repositories

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"studyum/internal/i18n/entities"
)

type Repository interface {
	GetByCode(ctx context.Context, lang string, code string) (entities.I18n, error)
}

type repository struct {
	session *pgxpool.Pool
}

func NewI18nRepository(session *pgxpool.Pool) Repository {
	return &repository{session: session}
}

func (r *repository) GetByCode(ctx context.Context, lang string, group string) (entities.I18n, error) {
	return r.query(ctx, lang, "SELECT key, \"%s\" as value FROM public WHERE \"group\"=$1", group)
}

func (r *repository) query(ctx context.Context, lang string, stmt string, values ...interface{}) (entities.I18n, error) {
	scanner, err := r.session.Query(ctx, fmt.Sprintf(stmt, lang), values...)
	if err != nil {
		return nil, err
	}
	defer scanner.Close()

	return r.scanMap(scanner)
}

func (r *repository) scanMap(scanner pgx.Rows) (dict entities.I18n, err error) {
	dict = make(entities.I18n)
	for scanner.Next() {
		var (
			key   string
			value string
		)

		if err = scanner.Scan(&key, &value); err != nil {
			return
		}

		dict[key] = value
	}

	return dict, nil
}
