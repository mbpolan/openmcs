-------------------------------------
-- Teleport spell library functions
-------------------------------------

--- Teleports a player using a standard spell book spell with an animation.
-- The vararg should be an even-sized list of pairs of IDs, one for the rune ID followed
-- by the amount of that rune consume. If the player does not have enough runes to cast
-- the spell, a server message will be sent to them instead.
-- @param player The player to teleport
-- @param x The destination x-coordinate, in global coordinates
-- @param y The destination y-coordinate, in global coordinates
-- @param z The destination z-coordinate
function teleport_standard(player, x, y, z, ...)
    ok = player:consume_items(unpack(arg))
    if not ok then
        player:server_message("You do not have enough runes to cast this spell.")
        return
    end

    -- animate the teleportation spell with a graphic
    player:animate(714, 5)
    player:graphic(308, 75, 45, 5)

    -- teleport the player to their new position
    player:teleport(x, y, z)
end
