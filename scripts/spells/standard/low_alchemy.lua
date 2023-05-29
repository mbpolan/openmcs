--- Casts the Low Alchemy spell on an inventory item.
-- @param player The player casting the spell
-- @param item The item the spell is being cast on
-- @param slot_id The ID of the inventory slot containing the item
function spell_low_alchemy(player, item, slot_id)
    -- require 3 fire runes and 1 nature rune
    ok = player:consume_runes(554, 3, 561, 1)
    if not ok then
        player:server_message("You do not have enough runes to cast this spell.")
        return
    end

    -- animate the player with a graphic
    player:animate(712, 2)
    player:graphic(112, 75, 2)

    -- consume the target item
    ok = player:consume_item(slot_id)
    if not ok then
        return
    end

    -- add the necessary amount of gold to the player's inventory
    coins = math.floor(item:value() * 0.4)
    player:add_item(995, coins)

    -- switch back to the spell book
    player:sidebar_tab(CLIENT_TAB_SPELLS)

    -- grant 31 magic exp after a 5 tick delay
    player:grant_experience(SKILL_MAGIC, 31, 5)
end