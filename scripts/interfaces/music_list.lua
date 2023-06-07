-- map of music identifiers for interface text
MUSIC_IDS_TO_INTERFACES = {
    [MUSIC_MEDIEVAL] = 4368,
    [MUSIC_GRND_SCP] = 13972
}

-- map of music identifiers to flags if they are available without unlock
DEFAULT_MUSIC_IDS = {
    [MUSIC_GRND_SCP] = true
}

-- map of music text interfaces to music identifiers
MUSIC_INTERFACES_TO_IDS = {}
for k, v in pairs(MUSIC_IDS_TO_INTERFACES) do
    MUSIC_INTERFACES_TO_IDS[v] = k
end

--- Handles an action performed on the music list interface.
-- @param player The player performing the action
-- @param interface The interface that received the action
function interface_962_on_action(player, interface)
    local id = interface:id()

    if id == 6269 then
        -- auto mode
    elseif id == 6270 then
        -- manual mode
    elseif id == 9925 then
        -- loop mode
    end

    -- find the music track id that was clicked
    local music_id = MUSIC_INTERFACES_TO_IDS[id]
    if music_id == nil then
        player:server_message("This music track is not yet available!")
        return
    end

    -- play the track if the player has this track unlocked of if it's available by default
    local available = DEFAULT_MUSIC_IDS[music_id]
    if not available then
        available = player:music_track(music_id)
    end

    if available then
        player:play_music(music_id)
    else
        player:server_message("You need to unlock this music track first.")
    end
end

--- Handles updating the music list interface.
-- @param player The player
function interface_962_on_update(player)
    -- if a non-default track is unlocked, update its interface color
    for music_id, inf_id in pairs(MUSIC_IDS_TO_INTERFACES) do
        if DEFAULT_MUSIC_IDS[music_id] ~= true then
            if player:music_track(music_id) then
                player:interface_color(inf_id, 0, 31, 0)
            else
                player:interface_color(inf_id, 31, 0, 0)
            end
        end
    end
end