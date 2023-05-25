-------------------------------------
-- Interface: standard spell book
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
