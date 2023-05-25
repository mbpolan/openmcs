-------------------------------------
-- Interface: standard spell book
-------------------------------------
function interface_1151_on_action(player, interface)
    spell_id = interface:id()

    if spell_id == 1174 then
        spell_teleport_camelot(player)
    end
end
