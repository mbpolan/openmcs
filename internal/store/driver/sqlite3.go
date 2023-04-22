package driver

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/mbpolan/openmcs/internal/config"
	"github.com/mbpolan/openmcs/internal/model"
	_ "modernc.org/sqlite"
)

// SQLite3Driver is a driver that interfaces with a SQLite3 database.
type SQLite3Driver struct {
	db *sql.DB
}

// NewSQLite3Driver creates a new SQLite3 database driver.
func NewSQLite3Driver(cfg *config.SQLite3DatabaseConfig) (Driver, error) {
	// enable foreign keys
	query := "_fk=true"
	dsn := fmt.Sprintf("%s?%s", cfg.URI, query)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	return &SQLite3Driver{
		db: db,
	}, nil
}

// Migration returns a handle to the underlying store for use with SQLite3 migrations.
func (s *SQLite3Driver) Migration() (database.Driver, error) {
	return sqlite3.WithInstance(s.db, &sqlite3.Config{})
}

// LoadPlayer loads information about a player from a SQLite3 database.
func (s *SQLite3Driver) LoadPlayer(username string) (*model.Player, error) {
	stmt, err := s.db.Prepare(`
		SELECT
		    USERNAME,
		    PASSWORD_HASH,
		    GLOBAL_X,
		    GLOBAL_Y,
		    GLOBAL_Z,
		    GENDER,
		    FLAGGED,
		    MUTED,
		    PUBLIC_CHAT_MODE,
		    PRIVATE_CHAT_MODE,
		    INTERACTION_MODE,
		    TYPE,
		    LAST_LOGIN_DTTM
		FROM
		    PLAYER
		WHERE
		    USERNAME = ? COLLATE NOCASE
	`)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	p := &model.Player{
		Appearance: &model.EntityAppearance{},
	}

	row := stmt.QueryRow(username)

	var lastLoginDttm string
	err = row.Scan(
		&p.Username,
		&p.Password,
		&p.GlobalPos.X,
		&p.GlobalPos.Y,
		&p.GlobalPos.Z,
		&p.Appearance.Gender,
		&p.Flagged,
		&p.Muted,
		&p.Modes.PublicChat,
		&p.Modes.PrivateChat,
		&p.Modes.Interaction,
		&p.Type,
		&lastLoginDttm)
	if err != nil {
		return nil, err
	}

	return p, nil
	//
	//// TODO: maintain player position
	//globalPos := model.Vector3D{
	//	X: 3116 + x,
	//	Y: 3116 + y,
	//	Z: 0,
	//}

	//return model.NewPlayer(int(username[0]), username, "", model.PlayerNormal, false, globalPos), nil
}

// Close cleans up resources used by the SQLite3 driver.
func (s *SQLite3Driver) Close() error {
	return s.db.Close()
}
