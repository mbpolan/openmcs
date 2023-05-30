--- Casts the Low Alchemy spell on an inventory item.
-- @param player The player casting the spell
-- @param item The item the spell is being cast on
-- @param slot_id The ID of the inventory slot containing the item
-- @return true if the spell is complete and has no pending actions, false if not
function spell_low_alchemy(player, item, slot_id)
    local ok = skill_level_minimum(player, SKILL_MAGIC, 21, "You need magic level 21 to cast this spell.")
    if not ok then
        return true
    end

    -- require 3 fire runes and 1 nature rune
    ok = player:consume_items(ITEM_FIRE_RUNE, 3, ITEM_NATURE_RUNE, 1)
    if not ok then
        player:server_message("You do not have enough runes to cast this spell.")
        return true
    end

    -- animate the player with a graphic
    player:animate(712, 2)
    player:graphic(112, 75, 0, 2)

    -- consume the target item
    ok = player:consume_item_in_slot(slot_id, 1)
    if not ok then
        return true
    end

    -- add the necessary amount of gold to the player's inventory
    local coins = math.floor(item:value() * 0.4)
    player:add_item(ITEM_COINS2, coins)

    -- switch back to the spell book
    player:sidebar_tab(CLIENT_TAB_SPELLS)

    -- grant 31 magic exp after a 5 tick delay
    player:grant_experience(SKILL_MAGIC, 31)
    player:delay(5)
    return false
end