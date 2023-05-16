package driver

import "database/sql"

// safeNullInt32 returns the value of a possibly null int32 column or a default value.
func safeNullInt32(n sql.NullInt32, i int) int {
	if n.Valid {
		return int(n.Int32)
	}

	return i
}
