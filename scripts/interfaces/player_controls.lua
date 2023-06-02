--- Handles an action performed on the player controls interface.
-- @param player The player performing the action
-- @param interface The interface that received the action
function interface_147_on_action(player, interface)
    local id = interface:id()

    -- movement speed
    if id == 152 then
        -- walk
        player:movement_speed(MOVE_SPEED_WALK)
    elseif id == 153 then
        -- run
        player:movement_speed(MOVE_SPEED_RUN)
    end

    -- auto retaliate mode
    if id == 150 then
        -- on
        player:auto_retaliate(true)
    elseif id == 151 then
        -- off
        player:auto_retaliate(false)
    end

    -- emotes
    if id == 168 then
        -- yes
        player:animate(855, 2)
    elseif id == 169 then
        -- no
        player:animate(856, 2)
    elseif id == 162 then
        -- think
        player:animate(857, 4)
    elseif id == 164 then
        -- bow
        player:animate(858, 3)
    elseif id == 165 then
        -- angry
        player:animate(859, 4)
    elseif id == 161 then
        -- cry
        player:animate(860, 4)
    elseif id == 170 then
        -- laugh
        player:animate(861, 4)
    elseif id == 171 then
        -- cheer
        player:animate(862, 5)
    elseif id == 163 then
        -- wave
        player:animate(863, 4)
    elseif id == 167 then
        -- beckon
        player:animate(864, 3)
    elseif id == 172 then
        -- clap
        player:animate(865, 4)
    elseif id == 166 then
        -- dance
        player:animate(866, 7)
    elseif id == 13362 then
        -- panic
        player:animate(2105, 3)
    elseif id == 13363 then
        -- jig
        player:animate(2106, 3)
    elseif id == 13364 then
        -- spin
        player:animate(2107, 2)
    elseif id == 13365 then
        -- head bang
        player:animate(2108, 3)
    elseif id == 13366 then
        -- joy jump
        player:animate(2109, 2)
    elseif id == 13367 then
        -- raspberry
        player:animate(2110, 4)
    elseif id == 13368 then
        -- yawn
        player:animate(2111, 4)
    elseif id == 13369 then
        -- salute
        player:animate(2112, 2)
    elseif id == 13370 then
        -- shrug
        player:animate(2113, 2)
    elseif id == 11100 then
        -- blow kiss
        player:animate(1368, 2)
    elseif id == 667 then
        -- glass box
        player:animate(1131, 4)
    elseif id == 6503 then
        -- climb rope
        player:animate(1130, 4)
    elseif id == 6506 then
        -- lean
        player:animate(1129, 5)
    elseif id == 666 then
        -- glass wall
        player:animate(1128, 8)
    elseif id == 13383 then
        -- goblin bow
        player:animate(2127, 3)
    elseif id == 13384 then
        -- goblin dance
        player:animate(2128, 4)
    elseif id == 15166 then
        -- scared
        player:animate(2836, 6)
    elseif id == 18464 then
        -- zombie walk
        player:animate(3544, 6)
    elseif id == 18465 then
        -- zombie dance
        player:animate(3543, 6)
    end
end

--- Handles updating the player options interface.
-- @param player The player
function interface_147_on_update(player)
    -- movement speed: op code 173
    local speed = player:movement_speed()
    if speed == MOVE_SPEED_WALK then
        player:interface_setting(173, 0)
    elseif speed == MOVE_SPEED_RUN then
        player:interface_setting(173, 1)
    end

    -- auto retaliate: op code 172
    local auto_retaliate = player:auto_retaliate()
    if auto_retaliate then
        player:interface_setting(172, 0)
    else
        player:interface_setting(172, 1)
    end
end
