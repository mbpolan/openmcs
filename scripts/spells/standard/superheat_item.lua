--- Casts the Superheat Item spell on an inventory item.
-- @param player The player casting the spell
-- @param item The item the spell is being cast on
-- @param slot_id The ID of the inventory slot containing the item
-- @return true if the spell is complete and has no pending actions, false if not
function spell_superheat_item(player, item, slot_id)
    local ok = skill_level_minimum(player, SKILL_MAGIC, 43, "You need magic level 43 to cast this spell.")
    if not ok then
        return true
    end

    -- determine the source ore and target bar to smelt
    local source_item_id = item:id()
    local output_item_id = -1
    local min_smith_level = 0
    local smith_exp = 0

    -- always consume one of the source item
    local consume_item_ids = {source_item_id, 1}

    if source_item_id == 436 then
        -- copper: produce a bronze bar
        local num_tin_ore = player:num_inventory_items(438)
        if num_tin_ore < 1 then
            player:server_message("You need at least 1 tin ore to smelt a bronze bar.")
            return true
        end

        output_item_id = 2349
        min_smith_level = 1
        smith_exp = 6.2

        -- consume 1 tin ore
        table.insert(consume_item_ids, 438)
        table.insert(consume_item_ids, 1)
    elseif source_item_id == 438 then
        -- tin: produce a bronze bar
        local num_copper_ore = player:num_inventory_items(436)
        if num_copper_ore < 1 then
            player:server_message("You need at least 1 tin ore to smelt a bronze bar.")
            return true
        end

        output_item_id = 2349
        min_smith_level = 1
        smith_exp = 6.2

        -- consume 1 copper ore
        table.insert(consume_item_ids, 436)
        table.insert(consume_item_ids, 1)
    elseif source_item_id == 440 then
        -- iron: produce steel bar if there is at least 2 coal, otherwise produce iron bar
        local num_coal = player:num_inventory_items_of(453)
        if num_coal >= 2 then
            output_item_id = 2353
            min_smith_level = 30
            smith_exp = 17.5

            -- consume 2 coal
            table.insert(consume_item_ids, 453)
            table.insert(consume_item_ids, 2)
        else
            output_item_id = 2351
            min_smith_level = 15
            smith_exp = 12.5
        end
    elseif source_item_id == 442 then
        -- silver: produce a silver bar
        output_item_id = 2355
        min_smith_level = 20
        smith_exp = 13.7
    elseif source_item_id == 444 then
        -- gold: produce a gold bar
        output_item_id = 2357
        min_smith_level = 40
        smith_exp = 22.5
    elseif source_item_id == 447 then
        -- mithril: produce a mithril bar
        local num_coal = player:num_inventory_items_of(453)
        if num_coal < 4 then
            player:server_message("You need at least 4 coal to smelt a mithril bar.")
            return true
        end

        output_item_id = 2359
        min_smith_level = 50
        smith_exp = 30

        -- consume 4 coal
        table.insert(consume_item_ids, 453)
        table.insert(consume_item_ids, 4)
    elseif source_item_id == 449 then
        -- adamantite: produce an adamantite bar
        local num_coal = player:num_inventory_items_of(453)
        if num_coal < 6 then
            player:server_message("You need at least 6 coal to smelt an adamantite bar.")
            return true
        end

        output_item_id = 2361
        min_smith_level = 70
        smith_exp = 37.5

        -- consume 6 coal
        table.insert(consume_item_ids, 453)
        table.insert(consume_item_ids, 6)
    elseif source_item_id == 451 then
        -- runite: produce a runeite bar
        local num_coal = player:num_inventory_items_of(453)
        if num_coal < 8 then
            player:server_message("You need at least 8 coal to smelt a runite bar.")
            return true
        end

        output_item_id = 2363
        min_smith_level = 85
        smith_exp = 50

        -- consume 8 coal
        table.insert(consume_item_ids, 453)
        table.insert(consume_item_ids, 8)
    else
        player:server_message("You cannot cast this spell on this item.")
        return true
    end

    -- validate the player has the minimum required smithing level
    ok = skill_level_minimum(player, SKILL_SMITH, min_smith_level, "You need smithing level " .. min_smith_level .. " to smelt this bar.")
    if not ok then
        return true
    end

    -- require 4 fire runes and 1 nature rune
    ok = player:consume_items(554, 4, 561, 1)
    if not ok then
        player:server_message("You do not have enough runes to cast this spell.")
        return true
    end

    -- consume the source items
    ok = player:consume_items(unpack(consume_item_ids))
    if not ok then
        return true
    end

    -- produce the target item
    player:add_item(output_item_id, 1)

    -- animate the player with a graphic
    player:animate(725, 3)
    player:graphic(148, 75, 4, 4)

    -- switch back to the spell book
    player:sidebar_tab(CLIENT_TAB_SPELLS)

    -- grant 53 magic exp and additional smithing exp after a 3 tick delay
    player:grant_experience(SKILL_MAGIC, 65)
    player:grant_experience(SKILL_SMITH, smith_exp)
    player:delay(3)
    return false
end