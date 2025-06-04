-- Create the kanji_go schema
CREATE SCHEMA IF NOT EXISTS kanji_go;

-- Create users table
CREATE TABLE kanji_go.users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create kanji_packs junction table
CREATE TABLE kanji_go.user_kanji_packs (
    user_id INT REFERENCES kanji_go.users(id) ON DELETE CASCADE,
    pack_name VARCHAR(10) NOT NULL,
    PRIMARY KEY (user_id, pack_name)
);

-- Create starred and saved kanji junction tables
CREATE TABLE kanji_go.user_starred_kanji (
    user_id INT REFERENCES kanji_go.users(id) ON DELETE CASCADE,
    kanji_char_id INT NOT NULL,
    starred_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, kanji_char_id)
);

CREATE TABLE kanji_go.user_saved_kanji (
    user_id INT REFERENCES kanji_go.users(id) ON DELETE CASCADE,
    kanji_char_id INT NOT NULL,
    saved_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, kanji_char_id)
);

-- Create session table
CREATE TABLE kanji_go.sessions (
    session_id VARCHAR(255) PRIMARY KEY,
    curr_user VARCHAR(255) NULL REFERENCES kanji_go.users(username) ON DELETE SET NULL,
    curr_jlpt_level VARCHAR(2) DEFAULT 'n5' CHECK (curr_jlpt_level IN ('n1', 'n2', 'n3', 'n4', 'n5')),
    curr_page VARCHAR(255) DEFAULT 'practice',
    contact_popup_active BOOLEAN DEFAULT FALSE,
    login_popup_active BOOLEAN DEFAULT FALSE,
    payment_popup_active BOOLEAN DEFAULT FALSE,
    left_sidebar_active BOOLEAN DEFAULT FALSE,
    dark_mode_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create kanji table
CREATE TABLE kanji_go.kanji (
    kanji_char_id SERIAL PRIMARY KEY,
    kanji_char VARCHAR(10) NOT NULL UNIQUE,
    romaji_onyomi VARCHAR(255),
    romaji_kunyomi VARCHAR(255),
    hiragana_onyomi VARCHAR(255),
    hiragana_kunyomi VARCHAR(255),
    jlpt_level VARCHAR(2) CHECK (jlpt_level IN ('n1', 'n2', 'n3', 'n4', 'n5')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create kanji_creations table
CREATE TABLE kanji_go.kanji_creations (
    kanji_creation_id SERIAL PRIMARY KEY,
    kanji_char_id INT NOT NULL REFERENCES kanji_go.kanji(kanji_char_id) ON DELETE CASCADE,
    created_by VARCHAR(255) REFERENCES kanji_go.users(username) ON DELETE SET NULL,
    created_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    image_url VARCHAR(1024),
    mapping_url VARCHAR(1024),
    explanation TEXT NOT NULL,
    is_public BOOLEAN DEFAULT FALSE,
    stars INT DEFAULT 0,
    flags INT DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create temp_creation table
CREATE TABLE kanji_go.temp_creation (
    temp_id SERIAL PRIMARY KEY,
    kanji_char_id INT NOT NULL REFERENCES kanji_go.kanji(kanji_char_id) ON DELETE CASCADE,
    image_url VARCHAR(1024),
    mapping_url VARCHAR(1024),
    explanation TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Add default kanji pack for all users
CREATE OR REPLACE FUNCTION kanji_go.add_default_pack()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO kanji_go.user_kanji_packs (user_id, pack_name)
    VALUES (NEW.id, 'n5');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER add_default_pack_trigger
AFTER INSERT ON kanji_go.users
FOR EACH ROW
EXECUTE FUNCTION kanji_go.add_default_pack();

-- Add indexes for performance
CREATE INDEX idx_user_kanji_packs_user_id ON kanji_go.user_kanji_packs(user_id);
CREATE INDEX idx_user_starred_kanji_user_id ON kanji_go.user_starred_kanji(user_id);
CREATE INDEX idx_user_saved_kanji_user_id ON kanji_go.user_saved_kanji(user_id);
CREATE INDEX idx_kanji_jlpt_level ON kanji_go.kanji(jlpt_level);
CREATE INDEX idx_kanji_creations_kanji_char_id ON kanji_go.kanji_creations(kanji_char_id);
CREATE INDEX idx_kanji_creations_created_by ON kanji_go.kanji_creations(created_by);
CREATE INDEX idx_temp_creation_kanji_char_id ON kanji_go.temp_creation(kanji_char_id);
