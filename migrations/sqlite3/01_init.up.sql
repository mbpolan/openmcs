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
    -- mode for player's INTEGEReractions
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

-- create table for storing player friends lists
CREATE TABLE FRIEND (
    -- primary key
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    -- owning player
    PLAYER_ID INTEGER NOT NULL REFERENCES USERS(ID) ON DELETE CASCADE,
    -- friended player
    FRIEND_ID INTEGER NOT NULL REFERENCES USERS(ID) ON DELETE CASCADE,
    -- date time when the row was inserted
    CREATED_DTTM TEXT NOT NULL DEFAULT CURRENT_DATE,
    -- date time when the row was updated
    UPDATED_DTTM TEXT NULL
);

-- create an index on friend.player_id since it will be queried on
CREATE INDEX IDX_FRIEND_PLAYER_ID ON FRIEND(PLAYER_ID);

-- create a trigger on friend to manage the CREATED_DTTM column
CREATE TRIGGER
    FRIEND_CREATED_DTTM
AFTER INSERT ON
    FRIEND
BEGIN
    UPDATE
        FRIEND
    SET
        CREATED_DTTM = DATETIME('NOW')
    WHERE
        ID = NEW.ID;
END;

-- create a trigger on friend to manage the UPDATED_DTTM column
CREATE TRIGGER
    FRIEND_UPDATED_DTTM
AFTER UPDATE ON
    FRIEND
BEGIN
    UPDATE
        FRIEND
    SET
        UPDATED_DTTM = DATETIME('NOW')
    WHERE
        ID = NEW.ID;
END;

-- create table for storing player ignore lists
CREATE TABLE IGNORED (
    -- primary key
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    -- owning player
    PLAYER_ID INTEGER NOT NULL REFERENCES USERS(ID) ON DELETE CASCADE,
    -- ignored player
    IGNORED_ID INTEGER NOT NULL REFERENCES USERS(ID) ON DELETE CASCADE,
    -- date time when the row was inserted
    CREATED_DTTM TEXT NOT NULL DEFAULT CURRENT_DATE,
    -- date time when the row was updated
    UPDATED_DTTM TEXT NULL
);

-- create an index on ignored.player_id since it will be queried on
CREATE INDEX IDX_IGNORED_PLAYER_ID ON IGNORED(PLAYER_ID);

-- create a trigger on ignored to manage the CREATED_DTTM column
CREATE TRIGGER
    IGNORED_CREATED_DTTM
AFTER INSERT ON
    IGNORED
BEGIN
    UPDATE
        IGNORED
    SET
        CREATED_DTTM = DATETIME('NOW')
    WHERE
        ID = NEW.ID;
END;

-- create a trigger on ignored to manage the UPDATED_DTTM column
CREATE TRIGGER
    IGNORED_UPDATED_DTTM
AFTER UPDATE ON
    IGNORED
BEGIN
    UPDATE
        IGNORED
    SET
        UPDATED_DTTM = DATETIME('NOW')
    WHERE
        ID = NEW.ID;
END;
