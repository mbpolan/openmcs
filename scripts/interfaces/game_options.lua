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

--- Handles updating the game options interface.
-- @param player The player
function interface_904_on_update(player)
    -- screen brightness: op code 166
    local brightness = player:game_option(GAME_OPTION_SCREEN_BRIGHTNESS)
    if brightness == SCREEN_BRIGHTNESS_DARK then
        player:interface_setting(166, 1)
    elseif brightness == SCREEN_BRIGHTNESS_NORMAL then
        player:interface_setting(166, 2)
    elseif brightness == SCREEN_BRIGHTNESS_BRIGHT then
        player:interface_setting(166, 3)
    elseif brightness == SCREEN_BRIGHTNESS_VERY_BRIGHT then
        player:interface_setting(166, 4)
    end

    -- chat effects: op code 171
    local chat_effects = player:game_option(GAME_OPTION_CHAT_EFFECTS)
    if chat_effects == CHAT_EFFECTS_ON then
        player:interface_setting(171, 0)
    elseif chat_effects == CHAT_EFFECTS_OFF then
        player:interface_setting(171, 1)
    end

    -- split private chat: op code 287
    local split_private = player:game_option(GAME_OPTION_SPLIT_PRIVATE_CHAT)
    if split_private == SPLIT_PRIVATE_CHAT_ON then
        player:interface_setting(287, 1)
    elseif split_private == SPLIT_PRIVATE_CHAT_OFF then
        player:interface_setting(287, 0)
    end

    -- mouse buttons: op code 170
    local mouse_buttons = player:game_option(GAME_OPTION_MOUSE_BUTTONS)
    if mouse_buttons == MOUSE_BUTTONS_ONE then
        player:interface_setting(170, 1)
    elseif mouse_buttons == MOUSE_BUTTONS_TWO then
        player:interface_setting(170, 0)
    end

    -- accept aid: op code 427
    local accept_aid = player:game_option(GAME_OPTION_ACCEPT_AID)
    if accept_aid == ACCEPT_AID_YES then
        player:interface_setting(427, 1)
    elseif accept_aid == ACCEPT_AID_NO then
        player:interface_setting(427, 0)
    end

    -- music volume: op code 168
    local music_volume = player:game_option(GAME_OPTION_MUSIC_VOLUME)
    if music_volume == "0" then
        player:interface_setting(168, 4)
    elseif music_volume == "1" then
        player:interface_setting(168, 3)
    elseif music_volume == "2" then
        player:interface_setting(168, 2)
    elseif music_volume == "3" then
        player:interface_setting(168, 1)
    elseif music_volume == "4" then
        player:interface_setting(168, 0)
    end

    -- effects volume: op code 169
    local effects_volume = player:game_option(GAME_OPTION_EFFECTS_VOLUME)
    if effects_volume == "0" then
        player:interface_setting(169, 4)
    elseif effects_volume == "1" then
        player:interface_setting(169, 3)
    elseif effects_volume == "2" then
        player:interface_setting(169, 2)
    elseif effects_volume == "3" then
        player:interface_setting(169, 1)
    elseif effects_volume == "4" then
        player:interface_setting(169, 0)
    end
end