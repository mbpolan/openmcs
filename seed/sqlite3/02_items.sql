-- Migration: 02_items.up.sql
-- Description: builds server-side data for items

INSERT OR IGNORE INTO ITEM_ATTRIBUTES (
    ITEM_ID, EQUIP_SLOT_ID, WEIGHT, WEAPON_STYLE, ATTACK_MAGIC, DEFENSE_STAB
) VALUES
    (88, 10, 0.340, NULL, 0, 0),
    (775, 9, 0.226, NULL, 0, 0),
    (882, 13, 0.000, NULL, 0, 0),
    (1038, 0, 0.056, NULL, 0, 0),
    (1079, 7, 9.071, NULL, 0, 0),
    (1113, 4, 6.803, NULL, 0, 0),
    (1187, 5, 3.175, NULL, -6, 50),
    (1201, 5, 5.443, NULL, 0, 0),
    (1333, 3, 1.814, 'SLASH_SWORD', 0, 0),
    (1704, 2, 0.010, NULL, 0, 0),
    (2572, 12, 0.006, NULL, 0, 0),
    (4315, 1, 0.0453, NULL, 0, 0)
;