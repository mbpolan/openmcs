-- initialize skill lookup
INSERT INTO SKILL_LOOKUP (
    ID, NAME
) VALUES
      (0, 'Attack'),
      (1, 'Defense'),
      (2, 'Strength'),
      (3, 'Hitpoints'),
      (4, 'Ranged'),
      (5, 'Prayer'),
      (6, 'Magic'),
      (7, 'Cooking'),
      (8, 'Woodcutting'),
      (9, 'Fletching'),
      (10, 'Fishing'),
      (11, 'Firemaking'),
      (12, 'Crafting'),
      (13, 'Smithing'),
      (14, 'Mining'),
      (15, 'Herblore'),
      (16, 'Agility'),
      (17, 'Thieving'),
      (18, 'Slayer'),
      (19, 'Farming'),
      (20, 'Runecraft')
;

-- add a player with username "mike" and password "mike"
INSERT INTO PLAYER (
    USERNAME, PASSWORD_HASH, EMAIL, GLOBAL_X, GLOBAL_Y, GLOBAL_Z, GENDER, FLAGGED, MUTED, PUBLIC_CHAT_MODE, PRIVATE_CHAT_MODE, INTERACTION_MODE, TYPE
) VALUES
    ('Mike', 'ab2341a2a5ec2b5ebd0ba195499408ac4ff54e63b52fa25b0b506d9f0a67cd35', 'mike@example.com', 3209, 3429, 0, 0, 0, 0, 0, 0, 0, 0),
    ('Hurz', 'cdb162dd4fa9ae4245525a9ec9c2868f19578e00c500cb9f192ea12c9330191d', 'mike@example.com', 3222, 3428, 0, 0, 0, 0, 0, 0, 0, 0)
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
INSERT INTO PLAYER_SKILL (PLAYER_ID, SKILL_ID, LEVEL, EXPERIENCE)
SELECT 1, ID, 1, 0
FROM SKILL_LOOKUP;

INSERT INTO PLAYER_SKILL (PLAYER_ID, SKILL_ID, LEVEL, EXPERIENCE)
SELECT 2, ID, 1, 0
FROM SKILL_LOOKUP;

INSERT INTO ITEM_ATTRIBUTES (
    ITEM_ID, EQUIP_SLOT_ID, WEIGHT, TWO_HANDED
) VALUES
    (1187, 5, 3.175, 0),
    (1333, 3, 1.814, 0)
;