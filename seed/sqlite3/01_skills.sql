-- Migration: 01_skills.up.sql
-- Description: builds baseline data for skills

INSERT INTO SKILL_LOOKUP (
    ID, NAME
) VALUES
      (0, 'Attack'),
      (1, 'Defense'),
      (2, 'Strength'),
      (3, 'Hitpoints'),
      (4, 'Ranged'),
      (5, 'Prayer'),
      (6, 'Magic'),
      (7, 'Cooking'),
      (8, 'Woodcutting'),
      (9, 'Fletching'),
      (10, 'Fishing'),
      (11, 'Firemaking'),
      (12, 'Crafting'),
      (13, 'Smithing'),
      (14, 'Mining'),
      (15, 'Herblore'),
      (16, 'Agility'),
      (17, 'Thieving'),
      (18, 'Slayer'),
      (19, 'Farming'),
      (20, 'Runecraft')
;