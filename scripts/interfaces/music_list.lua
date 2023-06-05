-- map of music identifiers for interface text
MUSIC_IDS_TO_INTERFACES = {
    [151] = 13972
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

    -- play a selected song
    local song_id = MUSIC_INTERFACES_TO_IDS[id]
    if song_id == nil then
        player:server_message("This music track is not yet available!")
        return
    end

    -- play the song
    -- TODO: check if the player has this track unlocked, if it's not a default track
    player:play_music(song_id)
end

--- Handles updating the music list interface.
-- @param player The player
function interface_962_on_update(player)
    -- TODO
end