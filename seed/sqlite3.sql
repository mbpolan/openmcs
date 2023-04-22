-- add a player with username "mike" and password "mike"
INSERT INTO PLAYER (
    ID, USERNAME, PASSWORD_HASH, EMAIL, GLOBAL_X, GLOBAL_Y, GLOBAL_Z, GENDER, FLAGGED, MUTED, PUBLIC_CHAT_MODE, PRIVATE_CHAT_MODE, INTERACTION_MODE, TYPE, LAST_LOGIN_DTTM
) VALUES (
    0, 'Mike', 'ab2341a2a5ec2b5ebd0ba195499408ac4ff54e63b52fa25b0b506d9f0a67cd35', 'mike@mbpolan.com', 3116, 3116, 0, 0, 0, 0, 0, 0, 0, 0, DATETIME('NOW')
);

-- add appearance data for seed players
INSERT INTO PLAYER_APPEARANCE (
    PLAYER_ID, BODY_ID, APPEARANCE_ID
) VALUES
    (0, 0, 0),
    (0, 1, 0),
    (0, 2, 0),
    (0, 3, 0),
    (0, 4, 0)
;

-- add equipped item data for seed players
INSERT INTO PLAYER_EQUIPMENT (
    PLAYER_ID, SLOT_ID, ITEM_ID
) VALUES
    (0, 0, 256),
    (0, 1, 266),
    (0, 2, 274),
    (0, 3, 282),
    (0, 4, 292),
    (0, 5, 298),
    (0, 6, 289),
    (0, 7, 1564),
    (0, 8, 1552),
    (0, 9, 1699),
    (0, 10, 1817),
    (0, 11, 2216)
;
