-------------------------------------
-- Skill library functions
-------------------------------------

--- Validates that a player meets a minimum level based on their current stat in a skill.
-- @param player The player
-- @param skill The skill to validate
-- @param min_stat The minimum required stat level
-- @param message The message to send the player if they don't meet the requirement
-- @return true if the player meets the requirement, false if not
function stat_level_minimum(player, skill, min_stat, message)
    local level = player:stat_level(skill)
    if level < min_stat then
        player:server_message(message)
        return false
    end

    return true
end

--- Validates that a player meets a minimum level based on their base level in a skill.
-- @param player The player
-- @param skill The skill to validate
-- @param min_level The minimum required level
-- @param message The message to send the player if they don't meet the requirement
-- @return true if the player meets the requirement, false if not
function skill_level_minimum(player, skill, min_level, message)
    local level = player:skill_level(skill)
    if level < min_level then
        player:server_message(message)
        return false
    end

    return true
end
