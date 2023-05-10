-- Migration: 01_init.up.sql
-- Description: creates baseline tables and database objects

-- ----------------------------------------------------------------------------
-- Table: PLAYER
-- ----------------------------------------------------------------------------

-- create table for storing general player data
CREATE TABLE PLAYER (
    -- primary key
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    -- player username
    USERNAME TEXT NOT NULL,
    -- sha256 hash of password
    PASSWORD_HASH TEXT NOT NULL,
    -- player email address
    EMAIL TEXT NOT NULL,
    -- position along x-axis in global coordinates
    GLOBAL_X INTEGER NOT NULL,
    -- position along y-axis in global coordinates
    GLOBAL_Y INTEGER NOT NULL,
    -- position along z-axis in global coordinates
    GLOBAL_Z INTEGER NOT NULL,
    -- character gender
    GENDER INTEGER NOT NULL,
    -- flag if the player's client should send anti-cheating metadata
    FLAGGED INTEGER NOT NULL,
    -- flag if the player is muted
    MUTED INTEGER NOT NULL,
    -- mode for player's public chat
    PUBLIC_CHAT_MODE INTEGER NOT NULL,
    -- mode for player's private chat
    PRIVATE_CHAT_MODE INTEGER NOT NULL,
    -- mode for player's interactions
    INTERACTION_MODE INTEGER NOT NULL,
    -- access rights of the player (normal, mod, admin, etc.)
    TYPE INTEGER NOT NULL,
    -- date time when the player last logged in
    LAST_LOGIN_DTTM TEXT NULL,
    -- date time when the row was inserted
    CREATED_DTTM TEXT NOT NULL DEFAULT CURRENT_DATE,
    -- date time when the row was updated
    UPDATED_DTTM TEXT NULL
);

-- create an index on player.username since it will be queried on login
CREATE INDEX IDX_PLAYER_USERNAME ON PLAYER (USERNAME);

-- create a trigger on player to manage the CREATED_DTTM column
CREATE TRIGGER
    PLAYER_CREATED_DTTM
AFTER INSERT ON
    PLAYER
BEGIN
    UPDATE
        PLAYER
    SET
        CREATED_DTTM = DATETIME('NOW')
    WHERE
        ID = NEW.ID;
END;

-- create a trigger on player to manage the UPDATED_DTTM column
CREATE TRIGGER
    PLAYER_UPDATED_DTTM
AFTER UPDATE ON
    PLAYER
BEGIN
    UPDATE
        PLAYER
    SET
        UPDATED_DTTM = DATETIME('NOW')
    WHERE
        ID = NEW.ID;
END;

-- ----------------------------------------------------------------------------
-- Table: PLAYER_EQUIPMENT
-- ----------------------------------------------------------------------------

-- create table for storing a player's equipped items
CREATE TABLE PLAYER_EQUIPMENT (
    -- primary key
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    -- owning player
    PLAYER_ID INTEGER NOT NULL REFERENCES USERS(ID),
    -- equipment slot id
    SLOT_ID INTEGER NOT NULL,
    -- equipped item id
    ITEM_ID INTEGER NOT NULL,
    -- date time when the row was inserted
    CREATED_DTTM TEXT NOT NULL DEFAULT CURRENT_DATE,
    -- date time when the row was updated
    UPDATED_DTTM TEXT NULL
);

-- create an index on player_equipment.player_id since it will be queried on
CREATE INDEX IDX_PLAYER_EQUIPMENT_PLAYER_ID ON PLAYER_EQUIPMENT(PLAYER_ID);

-- create a trigger on player_equipment to manage the CREATED_DTTM column
CREATE TRIGGER
    PLAYER_EQUIPMENT_CREATED_DTTM
AFTER INSERT ON
    PLAYER_EQUIPMENT
BEGIN
    UPDATE
        PLAYER_EQUIPMENT
    SET
        CREATED_DTTM = DATETIME('NOW')
    WHERE
        ID = NEW.ID;
END;

-- create a trigger on player_equipment to manage the UPDATED_DTTM column
CREATE TRIGGER
    PLAYER_EQUIPMENT_UPDATED_DTTM
AFTER UPDATE ON
    PLAYER_EQUIPMENT
BEGIN
    UPDATE
        PLAYER_EQUIPMENT
    SET
        UPDATED_DTTM = DATETIME('NOW')
    WHERE
        ID = NEW.ID;
END;

-- ----------------------------------------------------------------------------
-- Table: PLAYER_APPEARANCE
-- ----------------------------------------------------------------------------

-- create table for storing a player's character appearance
CREATE TABLE PLAYER_APPEARANCE (
    -- primary key
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    -- owning player
    PLAYER_ID INTEGER NOT NULL REFERENCES USERS(ID) ON DELETE CASCADE,
    -- body part id
    BODY_ID INTEGER NOT NULL,
    -- appearance modifier id
    APPEARANCE_ID INTEGER NOT NULL,
    -- date time when the row was inserted
    CREATED_DTTM TEXT NOT NULL DEFAULT CURRENT_DATE,
    -- date time when the row was updated
    UPDATED_DTTM TEXT NULL
);

-- create an index on player_appearance.player_id since it will be queried on
CREATE INDEX IDX_PLAYER_APPEARANCE_PLAYER_ID ON PLAYER_APPEARANCE(PLAYER_ID);

-- create a trigger on player_equipment to manage the CREATED_DTTM column
CREATE TRIGGER
    PLAYER_APPEARANCE_CREATED_DTTM
AFTER INSERT ON
    PLAYER_APPEARANCE
BEGIN
    UPDATE
        PLAYER_APPEARANCE
    SET
        CREATED_DTTM = DATETIME('NOW')
    WHERE
        ID = NEW.ID;
END;

-- create a trigger on player_appearance to manage the UPDATED_DTTM column
CREATE TRIGGER
    PLAYER_APPEARANCE_UPDATED_DTTM
AFTER UPDATE ON
    PLAYER_APPEARANCE
BEGIN
    UPDATE
        PLAYER_APPEARANCE
    SET
        UPDATED_DTTM = DATETIME('NOW')
    WHERE
        ID = NEW.ID;
END;

-- ----------------------------------------------------------------------------
-- Table: PLAYER_LIST
-- ----------------------------------------------------------------------------

-- create table for storing player friends and ignored lists
CREATE TABLE PLAYER_LIST (
    -- primary key
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    -- owning player
    PLAYER_ID INTEGER NOT NULL REFERENCES USERS(ID) ON DELETE CASCADE,
    -- referenced player
    OTHER_ID INTEGER NOT NULL REFERENCES USERS(ID) ON DELETE CASCADE,
    -- friend or ignored
    TYPE INTEGER NOT NULL,
    -- date time when the row was inserted
    CREATED_DTTM TEXT NOT NULL DEFAULT CURRENT_DATE,
    -- date time when the row was updated
    UPDATED_DTTM TEXT NULL
);

-- create an index on player_list.player_id since it will be queried on
CREATE INDEX IDX_PLAYER_LIST_PLAYER_ID ON PLAYER_LIST(PLAYER_ID);

-- create a trigger on player_list to manage the CREATED_DTTM column
CREATE TRIGGER
    PLAYER_LIST_CREATED_DTTM
AFTER INSERT ON
    PLAYER_LIST
BEGIN
    UPDATE
        PLAYER_LIST
    SET
        CREATED_DTTM = DATETIME('NOW')
    WHERE
        ID = NEW.ID;
END;

-- create a trigger on player_list to manage the UPDATED_DTTM column
CREATE TRIGGER
    PLAYER_LIST_UPDATED_DTTM
AFTER UPDATE ON
    PLAYER_LIST
BEGIN
    UPDATE
        PLAYER_LIST
    SET
        UPDATED_DTTM = DATETIME('NOW')
    WHERE
        ID = NEW.ID;
END;

-- ----------------------------------------------------------------------------
-- Table: SKILL_LOOKUP
-- ----------------------------------------------------------------------------

-- create a reference table for storing skills
CREATE TABLE SKILL_LOOKUP (
   -- primary key
   ID INTEGER PRIMARY KEY AUTOINCREMENT,
   -- skill name
   NAME TEXT NOT NULL,
    -- date time when the row was inserted
   CREATED_DTTM TEXT NOT NULL DEFAULT CURRENT_DATE,
    -- date time when the row was updated
   UPDATED_DTTM TEXT NULL
);

-- create a trigger on skill_lookup to manage the CREATED_DTTM column
CREATE TRIGGER
    SKILL_LOOKUP_CREATED_DTTM
AFTER INSERT ON
    SKILL_LOOKUP
BEGIN
    UPDATE
        SKILL_LOOKUP
    SET
        CREATED_DTTM = DATETIME('NOW')
    WHERE
            ID = NEW.ID;
END;

-- create a trigger on skill_lookup to manage the UPDATED_DTTM column
CREATE TRIGGER
    SKILL_LOOKUP_UPDATED_DTTM
AFTER UPDATE ON
    SKILL_LOOKUP
BEGIN
    UPDATE
        SKILL_LOOKUP
    SET
        UPDATED_DTTM = DATETIME('NOW')
    WHERE
        ID = NEW.ID;
END;

-- ----------------------------------------------------------------------------
-- Table: PLAYER_SKILL
-- ----------------------------------------------------------------------------

-- create table for storing player skill levels
CREATE TABLE PLAYER_SKILL (
    -- primary key
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    -- owning player
    PLAYER_ID INTEGER NOT NULL REFERENCES PLAYER(ID) ON DELETE CASCADE,
    -- skill id
    SKILL_ID INT NOT NULL REFERENCES SKILL_LOOKUP(ID) ON DELETE CASCADE ,
    -- skill level
    LEVEL INT NOT NULL,
    -- skill experience
    EXPERIENCE INT NOT NULL,
    -- date time when the row was inserted
    CREATED_DTTM TEXT NOT NULL DEFAULT CURRENT_DATE,
    -- date time when the row was updated
    UPDATED_DTTM TEXT NULL
);

-- create an index on player_skill.player_id since it will be queried on
CREATE INDEX IDX_PLAYER_SKILL_PLAYER_ID ON PLAYER_LIST(PLAYER_ID);

-- create a trigger on player_skill to manage the CREATED_DTTM column
CREATE TRIGGER
    PLAYER_SKILL_CREATED_DTTM
AFTER INSERT ON
    PLAYER_SKILL
BEGIN
    UPDATE
        PLAYER_SKILL
    SET
        CREATED_DTTM = DATETIME('NOW')
    WHERE
        ID = NEW.ID;
END;

-- create a trigger on player_skill to manage the UPDATED_DTTM column
CREATE TRIGGER
    PLAYER_SKILL_UPDATED_DTTM
AFTER UPDATE ON
    PLAYER_SKILL
BEGIN
    UPDATE
        PLAYER_SKILL
    SET
        UPDATED_DTTM = DATETIME('NOW')
    WHERE
        ID = NEW.ID;
END;

-- ----------------------------------------------------------------------------
-- Table: PLAYER_INVENTORY
-- ----------------------------------------------------------------------------

-- create table for storing player inventory items
CREATE TABLE PLAYER_INVENTORY (
    -- primary key
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    -- owning player
    PLAYER_ID INTEGER NOT NULL REFERENCES PLAYER(ID) ON DELETE CASCADE,
    -- slot id
    SLOT_ID INT NOT NULL CHECK (SLOT_ID >= 0 AND SLOT_ID < 28),
    -- item id
    ITEM_ID INT NOT NULL,
    -- stack size (amount)
    AMOUNT INT NOT NULL,
    -- date time when the row was inserted
    CREATED_DTTM TEXT NOT NULL DEFAULT CURRENT_DATE,
    -- date time when the row was updated
    UPDATED_DTTM TEXT NULL,
    -- enforce uniqueness on the player_id and slot_id
    UNIQUE (PLAYER_ID, SLOT_ID)
);

-- create an index on player_inventory.player_id since it will be queried on
CREATE INDEX IDX_PLAYER_INVENTORY_PLAYER_ID ON PLAYER_LIST(PLAYER_ID);

-- create a trigger on player_inventory to manage the CREATED_DTTM column
CREATE TRIGGER
    PLAYER_INVENTORY_CREATED_DTTM
AFTER INSERT ON
    PLAYER_INVENTORY
BEGIN
    UPDATE
        PLAYER_INVENTORY
    SET
        CREATED_DTTM = DATETIME('NOW')
    WHERE
        ID = NEW.ID;
END;

-- create a trigger on player_inventory to manage the UPDATED_DTTM column
CREATE TRIGGER
    PLAYER_INVENTORY_UPDATED_DTTM
AFTER UPDATE ON
    PLAYER_INVENTORY
BEGIN
    UPDATE
        PLAYER_INVENTORY
    SET
        UPDATED_DTTM = DATETIME('NOW')
    WHERE
        ID = NEW.ID;
END;

-- ----------------------------------------------------------------------------
-- Table: ITEM_ATTRIBUTES
-- ----------------------------------------------------------------------------

-- create table for storing item attributes
CREATE TABLE ITEM_ATTRIBUTES (
    -- primary key
    ITEM_ID INTEGER PRIMARY KEY,
    -- the slot this item is equipped into
    EQUIP_SLOT_ID INTEGER NULL,
    -- the speed at which this item deals damage
    SPEED INTEGER NULL,
    -- flag if the item requires two hand slots to wield
    TWO_HANDED INT NULL,
    -- the weight of the item
    WEIGHT REAL NOT NULL,
    -- date time when the row was inserted
    CREATED_DTTM TEXT NOT NULL DEFAULT CURRENT_DATE,
    -- date time when the row was updated
    UPDATED_DTTM TEXT NULL
);

-- create a trigger on item_attributes to manage the CREATED_DTTM column
CREATE TRIGGER
    ITEM_ATTRIBUTES_CREATED_DTTM
AFTER INSERT ON
    ITEM_ATTRIBUTES
BEGIN
    UPDATE
        ITEM_ATTRIBUTES
    SET
        CREATED_DTTM = DATETIME('NOW')
    WHERE
        ITEM_ID = NEW.ITEM_ID;
END;

-- create a trigger on item_attributes to manage the UPDATED_DTTM column
CREATE TRIGGER
    ITEM_ATTRIBUTES_UPDATED_DTTM
AFTER UPDATE ON
    ITEM_ATTRIBUTES
BEGIN
    UPDATE
        ITEM_ATTRIBUTES
    SET
        UPDATED_DTTM = DATETIME('NOW')
    WHERE
        ITEM_ID = NEW.ITEM_ID;
END;
