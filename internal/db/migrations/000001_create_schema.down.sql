-- Drop all tables and triggers in the correct order
DROP TRIGGER IF EXISTS add_default_pack_trigger ON kanji_go.users;
DROP FUNCTION IF EXISTS kanji_go.add_default_pack();

DROP TABLE IF EXISTS kanji_go.temp_creation;
DROP TABLE IF EXISTS kanji_go.kanji_creations;
DROP TABLE IF EXISTS kanji_go.user_starred_kanji;
DROP TABLE IF EXISTS kanji_go.user_saved_kanji;
DROP TABLE IF EXISTS kanji_go.user_kanji_packs;
DROP TABLE IF EXISTS kanji_go.kanji;
DROP TABLE IF EXISTS kanji_go.sessions;
DROP TABLE IF EXISTS kanji_go.users;

-- Drop the schema
DROP SCHEMA IF EXISTS kanji_go CASCADE;
