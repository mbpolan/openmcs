-- Migration: 02_items.up.sql
-- Description: builds server-side data for items

INSERT INTO ITEM_ATTRIBUTES (
    ITEM_ID, EQUIP_SLOT_ID, WEIGHT, TWO_HANDED
) VALUES
    (1187, 5, 3.175, 0),
    (1333, 3, 1.814, 0),
    (1201, 5, 5.443, 0),
    (1038, 0, 0.056, 0)
;