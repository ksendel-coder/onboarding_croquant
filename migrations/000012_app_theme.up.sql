ALTER TABLE apps
    ADD COLUMN primary_color TEXT NOT NULL DEFAULT '#696CF1' CHECK (primary_color ~ '^#[0-9A-Fa-f]{6}$'),
    ADD COLUMN background_color TEXT NOT NULL DEFAULT '#FFFFFF' CHECK (background_color ~ '^#[0-9A-Fa-f]{6}$'),
    ADD COLUMN text_color TEXT NOT NULL DEFAULT '#181818' CHECK (text_color ~ '^#[0-9A-Fa-f]{6}$');
