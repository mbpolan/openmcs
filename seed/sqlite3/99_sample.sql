-- Migration: 99_sample.up.sql
-- Description: sample data to bootstrap test players

-- add a player with username "mike" and password "mike"
INSERT INTO PLAYER (
    USERNAME, PASSWORD_HASH, EMAIL, GLOBAL_X, GLOBAL_Y, GLOBAL_Z, GENDER, UPDATE_DESIGN, FLAGGED, MUTED, PUBLIC_CHAT_MODE, PRIVATE_CHAT_MODE, INTERACTION_MODE, TYPE
) VALUES
    ('Mike', 'ab2341a2a5ec2b5ebd0ba195499408ac4ff54e63b52fa25b0b506d9f0a67cd35', 'mike@example.com', 3209, 3429, 0, 0, 1, 0, 0, 0, 0, 0, 0),
    ('Hurz', 'cdb162dd4fa9ae4245525a9ec9c2868f19578e00c500cb9f192ea12c9330191d', 'mike@example.com', 3222, 3428, 0, 0, 1, 0, 0, 0, 0, 0, 0)
;

-- add appearance data for seed players
INSERT INTO PLAYER_APPEARANCE (
    PLAYER_ID, HEAD_ID, FACE_ID, BODY_ID, ARMS_ID, HANDS_ID, LEGS_ID, FEET_ID
) VALUES
    -- mike
    (1, 256, 266, 274, 282, 289, 292, 298),
    -- hurz
    (2, 256, 266, 274, 282, 289, 292, 298)
;

-- initialize skills for players
INSERT INTO PLAYER_SKILL (PLAYER_ID, SKILL_ID, LEVEL, EXPERIENCE)
SELECT 1, ID, 1, 0
FROM SKILL_LOOKUP;

INSERT INTO PLAYER_SKILL (PLAYER_ID, SKILL_ID, LEVEL, EXPERIENCE)
SELECT 2, ID, 1, 0
FROM SKILL_LOOKUP;
