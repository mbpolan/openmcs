-------------------------------------
-- Spell: teleport to Camelot
-------------------------------------
function spell_teleport_camelot(player)
    -- require 1 law rune and 10 air runes
    ok = player:consume_runes(563, 1, 556, 10)
    if not ok then
        player:server_message("You do not have enough runes to cast this spell.")
        return
    end
end