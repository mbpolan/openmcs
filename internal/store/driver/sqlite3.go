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
	// prepare a player model for populating
	p := model.NewPlayer(username)

	// load their basic information first
	err := s.loadPlayerInfo(username, p)
	if err != nil {
		return nil, err
	}

	// load their equipped items
	err = s.loadPlayerEquipment(p.ID, p)
	if err != nil {
		return nil, err
	}

	// load their appearance
	err = s.loadPlayerAppearance(p.ID, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Close cleans up resources used by the SQLite3 driver.
func (s *SQLite3Driver) Close() error {
	return s.db.Close()
}

// loadPlayerInfo loads a player's basic information.
func (s *SQLite3Driver) loadPlayerInfo(username string, p *model.Player) error {
	// query the player's basic information
	stmt, err := s.db.Prepare(`
		SELECT
		    ID,
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
		return err
	}

	defer stmt.Close()

	// expect exactly zero or one row
	row := stmt.QueryRow(username)

	// extract their data into their model
	var lastLoginDttm string
	err = row.Scan(
		&p.ID,
		&p.Username,
		&p.PasswordHash,
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
		return err
	}

	return nil
}

// loadPlayerEquipment loads a player's equipped items.
func (s *SQLite3Driver) loadPlayerEquipment(id int, p *model.Player) error {
	// query for each slot the player has an equipped item
	stmt, err := s.db.Prepare(`
		SELECT
		    SLOT_ID,
		    ITEM_ID
		FROM
		    PLAYER_EQUIPMENT
		WHERE
		    PLAYER_ID = ?
	`)
	if err != nil {
		return err
	}

	rows, err := stmt.Query(id)
	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		var slotID, itemID int
		err := rows.Scan(&slotID, &itemID)
		if err != nil {
			return err
		}

		if slotID < 0 || slotID >= len(p.Appearance.Equipment) {
			return fmt.Errorf("slot ID out of bounds: %d", slotID)
		}

		p.Appearance.Equipment[slotID] = itemID
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}

// loadPlayerAppearance loads a player's body appearance.
func (s *SQLite3Driver) loadPlayerAppearance(id int, p *model.Player) error {
	// query for each body the player has an appearance attribute
	stmt, err := s.db.Prepare(`
		SELECT
		    BODY_ID,
		    APPEARANCE_ID
		FROM
		    PLAYER_APPEARANCE
		WHERE
		    PLAYER_ID = ?
	`)
	if err != nil {
		return err
	}

	rows, err := stmt.Query(id)
	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		var bodyID, itemID int
		err := rows.Scan(&bodyID, &itemID)
		if err != nil {
			return err
		}

		if bodyID < 0 || bodyID >= len(p.Appearance.Body) {
			return fmt.Errorf("body ID out of bounds: %d", bodyID)
		}

		p.Appearance.Body[bodyID] = itemID
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}
