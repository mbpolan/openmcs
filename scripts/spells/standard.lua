-------------------------------------
-- Standard spell book
-------------------------------------

--- Handles an action performed on the standard spell book parent interface
-- @param player The player that performed the action
-- @param interface The spell interface the action was performed on
function interface_1151_on_action(player, interface)
    local spell_id = interface:id()

    if spell_id == 1164 then
        spell_teleport_varrock(player)
    elseif spell_id == 1167 then
        spell_teleport_lumbridge(player)
    elseif spell_id == 1170 then
        spell_teleport_falador(player)
    elseif spell_id == 1174 then
        spell_teleport_camelot(player)
    elseif spell_id == 1540 then
        spell_teleport_ardougne(player)
    elseif spell_id == 1541 then
        spell_teleport_watchtower(player)
    elseif spell_id == 7455 then
        spell_teleport_trollheim(player)
    elseif spell_id == 18470 then
        spell_teleport_ape_atoll(player)
    end
end

--- Handles a spell that was cast on an inventory item.
-- @param player The player casting the spell
-- @param item The item the spell was cast on
-- @param slot_id The ID of the inventory slot containing the item
-- @param inv_interface The inventory interface
-- @param spell_book_interface The interface where the spell was chosen from
-- @param spell_interface The interface for the spell
-- @return true if the spell is complete and has no pending actions, false if not
function on_cast_spell_on_item(player, item, slot_id, inv_interface, spell_book_interface, spell_interface)
    local spell_id = spell_interface:id()

    if spell_id == 1162 then
        return spell_low_alchemy(player, item, slot_id)
    elseif spell_id == 1178 then
        return spell_high_alchemy(player, item, slot_id)
    elseif spell_id == 1173 then
        return spell_superheat_item(player, item, slot_id)
    end

    print('unknown spell: ', spell_id)
    player:server_message("This spell is not yet available!")
    return true
end