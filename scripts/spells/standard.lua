-------------------------------------
-- Standard spell book
-------------------------------------


function interface_1151_on_action(player, interface)
    spell_id = interface:id()

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
    end
end

--- Handles a spell that was cast on an inventory item.
-- @param player The player casting the spell
-- @param item The item the spell was cast on
-- @param slot_id The ID of the inventory slot containing the item
-- @param inv_interface The inventory interface
-- @param spell_book_interface The interface where the spell was chosen from
-- @param spell_interface The interface for the spell
function on_cast_spell_on_item(player, item, slot_id, inv_interface, spell_book_interface, spell_interface)
    spell_id = spell_interface:id()

    if spell_id == 1162 then
        spell_low_alchemy(player, item, slot_id)
    elseif spell_id == 1178 then
        spell_high_alchemy(player, item, slot_id)
    else
        print('unknown spell: ', spell_id)
    end
end