-- map of song identifiers for interface text
SONG_IDS_TO_INTERFACES = {
    [151] = 13972
}

-- map of song text interfaces to song identifiers
SONG_INTERFACES_TO_IDS = {}
for k, v in pairs(SONG_IDS_TO_INTERFACES) do
    SONG_INTERFACES_TO_IDS[v] = k
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
    local song_id = SONG_INTERFACES_TO_IDS[id]
    if song_id == nil then
        player:server_message("This song is not yet available!")
        return
    end

    -- TODO: play the song
end

--- Handles updating the music list interface.
-- @param player The player
function interface_962_on_update(player)
    -- TODO
end