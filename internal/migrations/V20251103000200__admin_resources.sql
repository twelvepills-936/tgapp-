-- Admin resources: projects, proposals, channels, files, admins and links

-- files
CREATE TABLE IF NOT EXISTS files (
    id BIGSERIAL PRIMARY KEY,
    filename TEXT NOT NULL,
    path TEXT,
    url TEXT,
    mimetype TEXT,
    size BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- admins
CREATE TABLE IF NOT EXISTS admins (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'admin',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- channels
CREATE TABLE IF NOT EXISTS channels (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    platform TEXT NOT NULL CHECK (platform IN ('Telegram','Instagram','YouTube','X','VK','TikTok')),
    subscribers BIGINT NOT NULL DEFAULT 0,
    owner_profile_id BIGINT REFERENCES profiles(id) ON DELETE SET NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- link: profile to channels (as many-to-many)
CREATE TABLE IF NOT EXISTS profile_channels (
    profile_id BIGINT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    channel_id BIGINT NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    PRIMARY KEY (profile_id, channel_id)
);

-- projects
CREATE TABLE IF NOT EXISTS projects (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    location TEXT NOT NULL,
    theme TEXT NOT NULL,
    image TEXT,
    briefing TEXT,
    platform TEXT NOT NULL CHECK (platform IN ('Telegram','Instagram','YouTube','X','VK','TikTok')),
    subscribers BIGINT,
    reward BIGINT NOT NULL,
    deadline TIMESTAMPTZ NOT NULL,
    promoted BOOLEAN NOT NULL DEFAULT FALSE,
    status TEXT NOT NULL CHECK (status IN ('active','completed','cancelled')) DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- project rules
CREATE TABLE IF NOT EXISTS project_rules (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    text TEXT NOT NULL
);

-- link: saved projects by profile
CREATE TABLE IF NOT EXISTS profile_saved_projects (
    profile_id BIGINT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (profile_id, project_id)
);

-- proposals
CREATE TABLE IF NOT EXISTS proposals (
    id BIGSERIAL PRIMARY KEY,
    status_value TEXT NOT NULL CHECK (status_value IN (
        'waiting_approval','waiting_channel_approval','waiting_attachments_approval',
        'approved','rejected','channel_approved','channel_rejected','attachments_rejected','attachments_approved'
    )) DEFAULT 'waiting_channel_approval',
    status_details TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    channel_id BIGINT NOT NULL REFERENCES channels(id) ON DELETE RESTRICT,
    submit_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    initiator_profile_id BIGINT NOT NULL REFERENCES profiles(id) ON DELETE RESTRICT,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    erid TEXT,
    attachments_text TEXT,
    attachments_files JSONB,
    deadline TIMESTAMPTZ
);


