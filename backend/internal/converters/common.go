package converters

import (
	"github.com/jackc/pgx/v5/pgtype"
)

func StringToPgUUID(idStr string) (pgtype.UUID, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(idStr); err != nil {
		return pgtype.UUID{}, err
	}

	return pgUUID, nil
}
