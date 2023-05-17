package driver

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/mbpolan/openmcs/internal/config"
	"github.com/mbpolan/openmcs/internal/model"
	_ "modernc.org/sqlite"
	"strings"
)

const playerListTypeFriend int = 0
const playerListTypeIgnored int = 1

// slotIDsToEquipmentSlots maps numeric slot IDs from the database to model.EquipmentSlotType values.
var slotIDsToEquipmentSlots = map[int]model.EquipmentSlotType{
	0:  model.EquipmentSlotTypeHead,
	1:  model.EquipmentSlotTypeCape,
	2:  model.EquipmentSlotTypeNecklace,
	3:  model.EquipmentSlotTypeWeapon,
	4:  model.EquipmentSlotTypeBody,
	5:  model.EquipmentSlotTypeShield,
	7:  model.EquipmentSlotTypeLegs,
	9:  model.EquipmentSlotTypeHands,
	10: model.EquipmentSlotTypeFeet,
	12: model.EquipmentSlotTypeRing,
	13: model.EquipmentSlotTypeAmmo,
}

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

// LoadItemAttributes loads information about all item attributes from a SQLite3 database.
func (s *SQLite3Driver) LoadItemAttributes() ([]*model.ItemAttributes, error) {
	stmt, err := s.db.Prepare(`
		SELECT
		    ITEM_ID,
		    EQUIP_SLOT_ID,
		    SPEED,
		    WEIGHT,
		    TWO_HANDED,
		    ATTACK_STAB,
			ATTACK_SLASH,
			ATTACK_CRUSH,
			ATTACK_MAGIC,
			ATTACK_RANGE,
			DEFENSE_STAB,
			DEFENSE_SLASH,
			DEFENSE_CRUSH,
			DEFENSE_MAGIC,
			DEFENSE_RANGE,
			STRENGTH_BONUS,
			PRAYER_BONUS
		FROM
		    ITEM_ATTRIBUTES
	`)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	var attributes []*model.ItemAttributes

	defer rows.Close()
	for rows.Next() {
		var itemID int
		var weight float64
		var slotID, speed, twoHanded sql.NullInt32
		var atkStab, atkSlash, atkCrush, atkMagic, atkRange sql.NullInt32
		var defStab, defSlash, defCrush, defMagic, defRange sql.NullInt32
		var strength, prayer sql.NullInt32

		err := rows.Scan(&itemID, &slotID, &speed, &weight, &twoHanded,
			&atkStab, &atkSlash, &atkCrush, &atkMagic, &atkRange,
			&defStab, &defSlash, &defCrush, &defMagic, &defRange,
			&strength, &prayer)
		if err != nil {
			return nil, err
		}

		nature := model.ItemNatureNotUsable
		if slotID.Valid {
			if twoHanded.Valid && twoHanded.Int32 == 1 {
				nature |= model.ItemNatureEquipmentTwoHanded
			} else {
				nature |= model.ItemNatureEquipmentOneHanded
			}
		}

		equipSlotID := model.EquipmentSlotTypeHead
		if slotID.Valid {
			equipSlotID = slotIDsToEquipmentSlots[int(slotID.Int32)]
		}

		itemSpeed := 0
		if speed.Valid {
			itemSpeed = int(speed.Int32)
		}

		attributes = append(attributes, &model.ItemAttributes{
			ItemID:        itemID,
			Nature:        nature,
			EquipSlotType: equipSlotID,
			Speed:         itemSpeed,
			Weight:        weight,
			Attack: model.ItemCombatAttributes{
				Stab:  safeNullInt32(atkStab, 0),
				Slash: safeNullInt32(atkSlash, 0),
				Crush: safeNullInt32(atkCrush, 0),
				Magic: safeNullInt32(atkMagic, 0),
				Range: safeNullInt32(atkRange, 0),
			},
			Defense: model.ItemCombatAttributes{
				Stab:  safeNullInt32(defStab, 0),
				Slash: safeNullInt32(defSlash, 0),
				Crush: safeNullInt32(defCrush, 0),
				Magic: safeNullInt32(defMagic, 0),
				Range: safeNullInt32(defRange, 0),
			},
			StrengthBonus: safeNullInt32(strength, 0),
			PrayerBonus:   safeNullInt32(prayer, 0),
		})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return attributes, nil
}

// SavePlayer updates a player's information in a SQLite3 database.
func (s *SQLite3Driver) SavePlayer(p *model.Player) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// save the player's basic information
	err = s.savePlayerInfo(p)
	if err != nil {
		return err
	}

	// save their equipment
	err = s.savePlayerEquipment(p)
	if err != nil {
		return err
	}

	// save their appearance
	err = s.savePlayerAppearance(p)
	if err != nil {
		return err
	}

	// save their friends and ignored lists
	err = s.savePlayerLists(p)
	if err != nil {
		return err
	}

	// save their skills
	err = s.savePlayerSkills(p)
	if err != nil {
		return err
	}

	// save their inventory
	err = s.savePlayerInventory(p)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// LoadPlayer loads information about a player from a SQLite3 database.
func (s *SQLite3Driver) LoadPlayer(username string) (*model.Player, error) {
	// prepare a player model for populating
	p := model.NewPlayer(username)

	// load their basic information first
	err := s.loadPlayerInfo(username, p)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

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

	// load their friends and ignored lists
	err = s.loadPlayerLists(p)
	if err != nil {
		return nil, err
	}

	// load their skills
	err = s.loadPlayerSkills(p)
	if err != nil {
		return nil, err
	}

	// load their inventory
	err = s.loadPlayerInventory(p)
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
		    UPDATE_DESIGN,
		    FLAGGED,
		    MUTED,
		    PUBLIC_CHAT_MODE,
		    PRIVATE_CHAT_MODE,
		    INTERACTION_MODE,
		    TYPE
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
	err = row.Scan(
		&p.ID,
		&p.Username,
		&p.PasswordHash,
		&p.GlobalPos.X,
		&p.GlobalPos.Y,
		&p.GlobalPos.Z,
		&p.Appearance.Gender,
		&p.UpdateDesign,
		&p.Flagged,
		&p.Muted,
		&p.Modes.PublicChat,
		&p.Modes.PrivateChat,
		&p.Modes.Interaction,
		&p.Type)
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
		    ITEM_ID,
		    AMOUNT
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
		var slotID, itemID, amount int
		err := rows.Scan(&slotID, &itemID, &amount)
		if err != nil {
			return err
		}

		slot, ok := slotIDsToEquipmentSlots[slotID]
		if !ok {
			return fmt.Errorf("slot ID out of bounds: %d", slotID)
		}

		item := &model.Item{
			ID: itemID,
		}

		p.SetEquippedItem(item, amount, slot)
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
		    HEAD_ID,
		    FACE_ID,
		    BODY_ID,
		    ARMS_ID,
		    HANDS_ID,
		    LEGS_ID,
		    FEET_ID
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
		var headID, faceID, bodyID, armsID, handsID, legsID, feedID int
		err := rows.Scan(&headID, &faceID, &bodyID, &armsID, &handsID, &legsID, &feedID)
		if err != nil {
			return err
		}

		p.Appearance.Base = model.EntityBase{
			Head:  headID,
			Face:  faceID,
			Body:  bodyID,
			Arms:  armsID,
			Hands: handsID,
			Legs:  legsID,
			Feet:  feedID,
		}
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}

// loadPlayerLists loads a player's friends and ignored lists.
func (s *SQLite3Driver) loadPlayerLists(p *model.Player) error {
	stmt, err := s.db.Prepare(`
		SELECT
		    p.USERNAME, l.TYPE
		FROM
		    PLAYER_LIST l
		JOIN
			PLAYER p ON p.ID = l.OTHER_ID
		WHERE
		    l.PLAYER_ID = ?
	`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	rows, err := stmt.Query(p.ID)
	if err != nil {
		return err
	}

	for rows.Next() {
		var username string
		var entryType int
		err := rows.Scan(&username, &entryType)
		if err != nil {
			return err
		}

		if entryType == playerListTypeFriend {
			p.Friends = append(p.Friends, username)
		} else if entryType == playerListTypeIgnored {
			p.Ignored = append(p.Ignored, username)
		}
	}

	return nil
}

// loadPlayerSkills loads a player's skills.
func (s *SQLite3Driver) loadPlayerSkills(p *model.Player) error {
	stmt, err := s.db.Prepare(`
		SELECT
		    SKILL_ID,
		    LEVEL,
		    EXPERIENCE
		FROM
		    PLAYER_SKILL
		WHERE
		    PLAYER_ID = ?
	`)
	if err != nil {
		return err
	}

	rows, err := stmt.Query(p.ID)
	if err != nil {
		return err
	}

	for rows.Next() {
		var skillID, level, experience int
		err := rows.Scan(&skillID, &level, &experience)
		if err != nil {
			return err
		}

		// map the skill id to a skill type
		skillType := model.SkillType(skillID)
		p.SetSkill(&model.Skill{
			Type:       skillType,
			Level:      level,
			Experience: experience,
		})
	}

	return nil
}

// loadPlayerInventory loads a player's inventory items.
func (s *SQLite3Driver) loadPlayerInventory(p *model.Player) error {
	stmt, err := s.db.Prepare(`
		SELECT
		    SLOT_ID,
		    ITEM_ID,
		    AMOUNT
		FROM
		    PLAYER_INVENTORY
		WHERE
		    PLAYER_ID = ?
	`)
	if err != nil {
		return err
	}

	rows, err := stmt.Query(p.ID)
	if err != nil {
		return err
	}

	for rows.Next() {
		var slotID, itemID, amount int
		err := rows.Scan(&slotID, &itemID, &amount)
		if err != nil {
			return err
		}

		// create a placeholder item for this item id
		item := &model.Item{
			ID: itemID,
		}

		// set the item into the player's inventory at the specified slot
		p.SetInventoryItem(item, amount, slotID)
	}

	return nil
}

// savePlayerInfo updates a player's basic information.
func (s *SQLite3Driver) savePlayerInfo(p *model.Player) error {
	stmt, err := s.db.Prepare(`
		UPDATE
			PLAYER
		SET
			GLOBAL_X = ?,
			GLOBAL_Y = ?,
			GLOBAL_Z = ?,
			GENDER = ?,
			FLAGGED = ?,
			MUTED = ?,
			PUBLIC_CHAT_MODE =  ?,
			PRIVATE_CHAT_MODE = ?,
			INTERACTION_MODE = ?,
			LAST_LOGIN_DTTM = DATETIME('NOW')
		WHERE
		    ID = ?
	`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	rs, err := stmt.Exec(
		p.GlobalPos.X,
		p.GlobalPos.Y,
		p.GlobalPos.Z,
		p.Appearance.Gender,
		p.Flagged,
		p.Muted,
		p.Modes.PublicChat,
		p.Modes.PrivateChat,
		p.Modes.Interaction,
		p.ID)
	if err != nil {
		return err
	}

	count, err := rs.RowsAffected()
	if err != nil {
		return err
	}

	if count != 1 {
		return fmt.Errorf("expected 1 row update, got %d", count)
	}

	return nil
}

// savePlayerEquipment saves a player's equipment.
func (s *SQLite3Driver) savePlayerEquipment(p *model.Player) error {
	// prepare a delete to clear out the player's equipment
	delStmt, err := s.db.Prepare(`
		DELETE FROM
		    PLAYER_EQUIPMENT
		WHERE
		    PLAYER_ID = ?
	`)
	if err != nil {
		return err
	}

	defer delStmt.Close()

	// delete all records for the player's equipment
	_, err = delStmt.Exec(p.ID)
	if err != nil {
		return err
	}

	insertTemplate := `
		INSERT INTO
			PLAYER_EQUIPMENT (
			    PLAYER_ID,
				SLOT_ID,
				ITEM_ID,
				AMOUNT
			)
		VALUES %s
	`

	valueTemplate := "(?, ?, ?, ?)"

	var bulk []string
	var values []any

	// collect the items in the player's equipment into tuples
	for slotID, slot := range p.Appearance.Equipment {
		bulk = append(bulk, valueTemplate)
		values = append(values, p.ID)
		values = append(values, int(slotID))
		values = append(values, slot.Item.ID)
		values = append(values, slot.Amount)
	}

	// bail out if there are no equipment items
	if len(bulk) == 0 {
		return nil
	}

	// prepare the final insert query
	insert := fmt.Sprintf(insertTemplate, strings.Join(bulk, ","))
	stmt, err := s.db.Prepare(insert)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(values...)
	if err != nil {
		return err
	}

	return nil
}

// savePlayerAppearance saves a player's appearance information.
func (s *SQLite3Driver) savePlayerAppearance(p *model.Player) error {
	stmt, err := s.db.Prepare(`
		UPDATE
			PLAYER_APPEARANCE
		SET
		    HEAD_ID = ?,
		    FACE_ID = ?,
		    BODY_ID = ?,
		    ARMS_ID = ?,
		    HANDS_ID = ?,
		    LEGS_ID = ?,
		    FEET_ID = ?
		WHERE
		    PLAYER_ID = ?
	`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	rs, err := stmt.Exec(p.Appearance.Base.Head,
		p.Appearance.Base.Face,
		p.Appearance.Base.Body,
		p.Appearance.Base.Arms,
		p.Appearance.Base.Hands,
		p.Appearance.Base.Legs,
		p.Appearance.Base.Feet,
		p.ID)
	if err != nil {
		return err
	}

	count, err := rs.RowsAffected()
	if err != nil {
		return err
	}

	if count != 1 {
		return fmt.Errorf("expected 1 row for appearance player ID %d, got %d", p.ID, count)
	}

	return nil
}

// savePlayerLists saves a player's friends and ignored lists.
func (s *SQLite3Driver) savePlayerLists(p *model.Player) error {
	// prepare a delete to clear out the player's lists
	delStmt, err := s.db.Prepare(`
		DELETE FROM
		    PLAYER_LIST
		WHERE
		    PLAYER_ID = ?
	`)
	if err != nil {
		return err
	}

	defer delStmt.Close()

	// delete all entries from the player's lists
	_, err = delStmt.Exec(p.ID)
	if err != nil {
		return err
	}

	// if there's nothing to insert, bail out early
	if len(p.Friends)+len(p.Ignored) == 0 {
		return nil
	}

	// prepare a template to bulk insert all entries
	insertTemplate := `
		INSERT OR IGNORE INTO
			PLAYER_LIST (
				 PLAYER_ID,
				 OTHER_ID,
				 TYPE
			)
		VALUES %s
	`

	valueTemplate := "(?, (SELECT ID FROM PLAYER WHERE USERNAME = ? COLLATE NOCASE), ?)"

	// prepare placeholders for each entry in the friends and ignored list
	// the player list has a maximum size that is less than sqlite's parameter restrictions, so we don't need to
	// explicitly limit it here
	var bulk []string
	var values []any
	for _, username := range p.Friends {
		bulk = append(bulk, valueTemplate)
		values = append(values, p.ID)
		values = append(values, username)
		values = append(values, playerListTypeFriend)
	}

	for _, username := range p.Ignored {
		bulk = append(bulk, valueTemplate)
		values = append(values, p.ID)
		values = append(values, username)
		values = append(values, playerListTypeIgnored)
	}

	// prepare the final insert query
	insert := fmt.Sprintf(insertTemplate, strings.Join(bulk, ","))
	stmt, err := s.db.Prepare(insert)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(values...)
	if err != nil {
		return err
	}

	return nil
}

func (s *SQLite3Driver) savePlayerSkills(p *model.Player) error {
	// prepare a delete to clear out the player's skills
	delStmt, err := s.db.Prepare(`
		DELETE FROM
		    PLAYER_SKILL
		WHERE
		    PLAYER_ID = ?
	`)
	if err != nil {
		return err
	}

	defer delStmt.Close()

	// delete all entries from the player's skills
	_, err = delStmt.Exec(p.ID)
	if err != nil {
		return err
	}

	insertTemplate := `
		INSERT INTO
			PLAYER_SKILL (
			    PLAYER_ID,
			    SKILL_ID,
				LEVEL,
				EXPERIENCE
			)
		VALUES %s
	`

	valueTemplate := "(?, ?, ?, ?)"

	var bulk []string
	var values []any

	// collect all of the player's skills into tuples
	for _, v := range p.Skills {
		bulk = append(bulk, valueTemplate)
		values = append(values, p.ID)
		values = append(values, int(v.Type))
		values = append(values, v.Level)
		values = append(values, v.Experience)
	}

	// prepare the final insert query
	insert := fmt.Sprintf(insertTemplate, strings.Join(bulk, ","))
	stmt, err := s.db.Prepare(insert)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(values...)
	if err != nil {
		return err
	}

	return nil
}

// savePlayerInventory saves a player's inventory.
func (s *SQLite3Driver) savePlayerInventory(p *model.Player) error {
	// prepare a delete to clear out the player's inventory
	delStmt, err := s.db.Prepare(`
		DELETE FROM
		    PLAYER_INVENTORY
		WHERE
		    PLAYER_ID = ?
	`)
	if err != nil {
		return err
	}

	defer delStmt.Close()

	// delete all entries from the player's inventory
	_, err = delStmt.Exec(p.ID)
	if err != nil {
		return err
	}

	insertTemplate := `
		INSERT INTO
			PLAYER_INVENTORY (
			    PLAYER_ID,
			    SLOT_ID,
				ITEM_ID,
				AMOUNT
			)
		VALUES %s
	`

	valueTemplate := "(?, ?, ?, ?)"

	var bulk []string
	var values []any

	// collect the items in the player's inventory into tuples
	for _, v := range p.Inventory {
		if v == nil {
			continue
		}

		bulk = append(bulk, valueTemplate)
		values = append(values, p.ID)
		values = append(values, v.ID)
		values = append(values, v.Item.ID)
		values = append(values, v.Amount)
	}

	// bail out if there are no inventory items
	if len(bulk) == 0 {
		return nil
	}

	// prepare the final insert query
	insert := fmt.Sprintf(insertTemplate, strings.Join(bulk, ","))
	stmt, err := s.db.Prepare(insert)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(values...)
	if err != nil {
		return err
	}

	return nil
}
