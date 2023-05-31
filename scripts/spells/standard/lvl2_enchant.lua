--- Casts the Level-2 Enchant spell on an inventory item.
-- @param player The player casting the spell
-- @param item The item the spell is being cast on
-- @param slot_id The ID of the inventory slot containing the item
-- @return true if the spell is complete and has no pending actions, false if not
function spell_level2_enchant(player, item, slot_id)
    local ok = skill_level_minimum(player, SKILL_MAGIC, 27, "You need magic level 27 to cast this spell.")
    if not ok then
        return true
    end

    -- determine the source item and the item it produces
    local source_item_id = item:id()
    local target_item_id = -1
    if source_item_id == ITEM_EMERALD_RING then
        target_item_id = ITEM_RING_OF_DUELING8
    elseif source_item_id == ITEM_EMERALD_NECKLACE then
        target_item_id = ITEM_BINDING_NECKLACE
    elseif source_item_id == ITEM_EMERALD_AMULET then
        target_item_id = ITEM_AMULET_OF_DEFENCE
    elseif source_item_id == ITEM_PRE_NATURE_AMULET then
        target_item_id = ITEM_AMULET_OF_NATURE
    else
        player:server_message("You cannot cast this spell on this item.")
        return false
    end

    -- require 1 water rune and 1 cosmic rune
    ok = player:consume_items(ITEM_WATER_RUNE, 1, ITEM_COSMIC_RUNE, 1)
    if not ok then
        player:server_message("You do not have enough runes to cast this spell.")
        return true
    end

    -- animate the player with a graphic
    player:animate(719, 3)
    player:graphic(154, 75, 0, 2)

    -- consume the target item
    ok = player:consume_item_in_slot(slot_id, 1)
    if not ok then
        return true
    end

    -- add the resulting item to the player's inventory
    player:add_item(target_item_id, 1)

    -- switch back to the spell book
    player:sidebar_tab(CLIENT_TAB_SPELLS)

    -- grant magic exp after a 3 tick delay
    player:grant_experience(SKILL_MAGIC, 37)
    player:delay(3)
    return false

end