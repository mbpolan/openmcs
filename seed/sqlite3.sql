-- add a player with username "mike" and password "mike"
INSERT INTO PLAYER (
    USERNAME, PASSWORD_HASH, EMAIL, GLOBAL_X, GLOBAL_Y, GLOBAL_Z, GENDER, FLAGGED, MUTED, PUBLIC_CHAT_MODE, PRIVATE_CHAT_MODE, INTERACTION_MODE, TYPE
) VALUES
    ('Mike', 'ab2341a2a5ec2b5ebd0ba195499408ac4ff54e63b52fa25b0b506d9f0a67cd35', 'mike@mbpolan.com', 3116, 3116, 0, 0, 0, 0, 0, 0, 0, 0),
    ('Hurz', 'cdb162dd4fa9ae4245525a9ec9c2868f19578e00c500cb9f192ea12c9330191d', 'mike@mbpolan.com', 3116, 3116, 0, 0, 0, 0, 0, 0, 0, 0)
;

-- add appearance data for seed players
INSERT INTO PLAYER_APPEARANCE (
    PLAYER_ID, BODY_ID, APPEARANCE_ID
) VALUES
    -- mike
    (1, 0, 0),
    (1, 1, 0),
    (1, 2, 0),
    (1, 3, 0),
    (1, 4, 0),
    -- hurz
    (2, 0, 0),
    (2, 1, 0),
    (2, 2, 0),
    (2, 3, 0),
    (2, 4, 0)
;

-- add equipped item data for seed players
INSERT INTO PLAYER_EQUIPMENT (
    PLAYER_ID, SLOT_ID, ITEM_ID
) VALUES
    -- mike
    (1, 0, 256),
    (1, 1, 266),
    (1, 2, 274),
    (1, 3, 282),
    (1, 4, 292),
    (1, 5, 298),
    (1, 6, 289),
    (1, 7, 1564),
    (1, 8, 1552),
    (1, 9, 1699),
    (1, 10, 1817),
    (1, 11, 2216),
    -- hurz
    (2, 0, 256),
    (2, 1, 266),
    (2, 2, 274),
    (2, 3, 282),
    (2, 4, 292),
    (2, 5, 298),
    (2, 6, 289),
    (2, 7, 1564),
    (2, 8, 1552),
    (2, 9, 1699),
    (2, 10, 1817),
    (2, 11, 2216)
;

-- initialize skills for players
INSERT INTO PLAYER_SKILL (
    PLAYER_ID, SKILL_ID, LEVEL, EXPERIENCE
) VALUES
    -- mike
    (1, 0, 1, 0),
    (1, 1, 1, 0),
    (1, 2, 1, 0),
    (1, 3, 1, 0),
    (1, 4, 1, 0),
    (1, 5, 1, 0),
    (1, 6, 1, 0),
    (1, 7, 1, 0),
    (1, 8, 1, 0),
    (1, 9, 1, 0),
    (1, 10, 1, 0),
    (1, 11, 1, 0),
    (1, 12, 1, 0),
    (1, 13, 1, 0),
    (1, 14, 1, 0),
    (1, 15, 1, 0),
    (1, 16, 1, 0),
    (1, 17, 1, 0),
    (1, 18, 1, 0),
    (1, 19, 1, 0),
    (1, 20, 1, 0),
    -- mike
    (2, 0, 1, 0),
    (2, 1, 1, 0),
    (2, 2, 1, 0),
    (2, 3, 1, 0),
    (2, 4, 1, 0),
    (2, 5, 1, 0),
    (2, 6, 1, 0),
    (2, 7, 1, 0),
    (2, 8, 1, 0),
    (2, 9, 1, 0),
    (2, 10, 1, 0),
    (2, 11, 1, 0),
    (2, 12, 1, 0),
    (2, 13, 1, 0),
    (2, 14, 1, 0),
    (2, 15, 1, 0),
    (2, 16, 1, 0),
    (2, 17, 1, 0),
    (2, 18, 1, 0),
    (2, 19, 1, 0),
    (2, 20, 1, 0)
;