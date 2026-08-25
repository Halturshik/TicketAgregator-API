package repository

import (
	"context"
	"database/sql"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

func (r *Repository) FindRule(ctx context.Context, transport string, isInternational bool, age int, isRussian bool) (*documents.DocumentRule, error) {
	var rule documents.DocumentRule
	err := r.DB.QueryRowContext(ctx, `
		SELECT transport_type, is_international, min_age, max_age, is_russian,
			allow_internal_passport, allow_international_passport,
			allow_birth_certificate, allow_foreign_passport
		FROM document_rules
		WHERE transport_type IN ('any', $1)
			AND is_international = $2
			AND is_russian = $3
			AND min_age <= $4 AND max_age >= $4
		ORDER BY CASE WHEN transport_type = $1 THEN 0 ELSE 1 END
		LIMIT 1
	`, transport, isInternational, isRussian, age).Scan(
		&rule.Transport, &rule.IsInternational, &rule.MinAge, &rule.MaxAge, &rule.IsRussian,
		&rule.AllowInternalPassport, &rule.AllowInternationalPassport,
		&rule.AllowBirthCertificate, &rule.AllowForeignPassport,
	)
	if err == sql.ErrNoRows {
		return nil, documents.ErrRuleNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rule, nil
}
