--- Handles an action performed on the game options interface.
-- @param player The player performing the action
-- @param interface The interface that received the action
function interface_904_on_action(player, interface)
    local id = interface:id()

    -- screen brightness
    if id == 906 then
        -- dark
        player:game_option(GAME_OPTION_SCREEN_BRIGHTNESS, SCREEN_BRIGHTNESS_DARK)
    elseif id == 908 then
        -- normal
        player:game_option(GAME_OPTION_SCREEN_BRIGHTNESS, SCREEN_BRIGHTNESS_NORMAL)
    elseif id == 910 then
        -- bright
        player:game_option(GAME_OPTION_SCREEN_BRIGHTNESS, SCREEN_BRIGHTNESS_BRIGHT)
    elseif id == 912 then
        -- very bright
        player:game_option(GAME_OPTION_SCREEN_BRIGHTNESS, SCREEN_BRIGHTNESS_VERY_BRIGHT)
    end

    -- mouse buttons
    if id == 913 then
        -- two
        player:game_option(GAME_OPTION_MOUSE_BUTTONS, MOUSE_BUTTONS_TWO)
    elseif id == 914 then
        -- one
        player:game_option(GAME_OPTION_MOUSE_BUTTONS, MOUSE_BUTTONS_ONE)
    end

    -- chat effects
    if id == 915 then
        -- on
        player:game_option(GAME_OPTION_CHAT_EFFECTS, CHAT_EFFECTS_ON)
    elseif id == 916 then
        -- off
        player:game_option(GAME_OPTION_CHAT_EFFECTS, CHAT_EFFECTS_OFF)
    end

    -- split private chat
    if id == 957 then
        -- on
        player:game_option(GAME_OPTION_SPLIT_PRIVATE_CHAT, SPLIT_PRIVATE_CHAT_ON)
    elseif id == 958 then
        -- off
        player:game_option(GAME_OPTION_SPLIT_PRIVATE_CHAT, SPLIT_PRIVATE_CHAT_OFF)
    end

    -- accept aid
    if id == 12464 then
        -- yes
        player:game_option(GAME_OPTION_ACCEPT_AID, ACCEPT_AID_YES)
    elseif id == 12465 then
        -- no
        player:game_option(GAME_OPTION_ACCEPT_AID, ACCEPT_AID_NO)
    end

    -- music volume
    if id == 930 then
        -- off
        player:game_option(GAME_OPTION_MUSIC_VOLUME, "0")
    elseif id == 931 then
        -- 1
        player:game_option(GAME_OPTION_MUSIC_VOLUME, "1")
    elseif id == 932 then
        -- 2
        player:game_option(GAME_OPTION_MUSIC_VOLUME, "2")
    elseif id == 933 then
        -- 3
        player:game_option(GAME_OPTION_MUSIC_VOLUME, "3")
    elseif id == 934 then
        -- 4
        player:game_option(GAME_OPTION_MUSIC_VOLUME, "4")
    end

    -- effect volume
    if id == 941 then
        -- off
        player:game_option(GAME_OPTION_EFFECTS_VOLUME, "0")
    elseif id == 942 then
        -- 1
        player:game_option(GAME_OPTION_EFFECTS_VOLUME, "1")
    elseif id == 943 then
        -- 2
        player:game_option(GAME_OPTION_EFFECTS_VOLUME, "2")
    elseif id == 944 then
        -- 3
        player:game_option(GAME_OPTION_EFFECTS_VOLUME, "3")
    elseif id == 945 then
        -- 4
        player:game_option(GAME_OPTION_EFFECTS_VOLUME, "4")
    end
end
