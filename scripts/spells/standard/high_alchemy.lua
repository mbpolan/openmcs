--- Casts the High Alchemy spell on an inventory item.
-- @param player The player casting the spell
-- @param item The item the spell is being cast on
-- @param slot_id The ID of the inventory slot containing the item
function spell_high_alchemy(player, item, slot_id)
    -- require 5 fire runes and 1 nature rune
    ok = player:consume_runes(554, 5, 561, 1)
    if not ok then
        player:server_message("You do not have enough runes to cast this spell.")
        return
    end

    -- animate the player
    player:animate(713, 4)

    -- consume the target item
    ok = player:consume_item(slot_id)
    if not ok then
        return
    end

    -- add the necessary amount of gold to the player's inventory
    coins = math.floor(item:value() * 0.6)
    player:add_item(995, coins)

    -- switch back to the spell book
    player:sidebar_tab(CLIENT_TAB_SPELLS)

    -- grant 65 magic exp after a 5 tick delay
    player:grant_experience(SKILL_MAGIC, 65, 5)
end