package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"studyum/internal/i18n/entities"
)

type Repository interface {
	GetByCode(ctx context.Context, lang string, code string) (entities.I18n, error)
}

type repository struct {
	session *sql.DB
}

func NewI18nRepository(session *sql.DB) Repository {
	return &repository{session: session}
}

func (r *repository) GetByCode(ctx context.Context, lang string, group string) (entities.I18n, error) {
	return r.query(ctx, lang, "SELECT key, %s as value FROM i18n.public WHERE group=$1", group)
}

func (r *repository) query(ctx context.Context, lang string, stmt string, values ...interface{}) (entities.I18n, error) {
	scanner, err := r.session.QueryContext(ctx, fmt.Sprintf(stmt, lang), values...)
	if err != nil {
		return nil, err
	}

	return r.scanMap(scanner)
}

func (r *repository) scanMap(scanner *sql.Rows) (dict entities.I18n, err error) {
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
