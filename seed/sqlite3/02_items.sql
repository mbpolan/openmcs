-- Migration: 02_items.up.sql
-- Description: builds server-side data for items

INSERT OR IGNORE INTO ITEM_ATTRIBUTES (
    ITEM_ID, EQUIP_SLOT_ID, WEIGHT, TWO_HANDED, ATTACK_MAGIC, DEFENSE_STAB
) VALUES
    (88, 10, 0.340, 0, 0, 0),
    (775, 9, 0.226, 0, 0, 0),
    (882, 13, 0.000, 0, 0, 0),
    (1038, 0, 0.056, 0, 0, 0),
    (1079, 7, 9.071, 0, 0, 0),
    (1113, 4, 6.803, 0, 0, 0),
    (1187, 5, 3.175, 0, -6, 50),
    (1201, 5, 5.443, 0, 0, 0),
    (1333, 3, 1.814, 0, 0, 0),
    (1704, 2, 0.010, 0, 0, 0),
    (2572, 12, 0.006, 0, 0, 0),
    (4315, 1, 0.0453, 0, 0, 0)
;