-- Create instances table
CREATE TABLE IF NOT EXISTS instances (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    token VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'close',
    owner_jid VARCHAR(255),
    profile_name VARCHAR(255),
    profile_pic_url TEXT,
    number VARCHAR(50),
    business_id VARCHAR(255),
    integration VARCHAR(100) DEFAULT 'WHATSAPP-BAILEYS',
    webhook_url TEXT,
    webhook_events TEXT[],
    webhook_headers JSONB,
    webhook_base64 BOOLEAN DEFAULT FALSE,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_seen TIMESTAMP WITH TIME ZONE
);

-- Create index on name for faster lookups
CREATE INDEX IF NOT EXISTS idx_instances_name ON instances(name);

-- Create index on status for filtering
CREATE INDEX IF NOT EXISTS idx_instances_status ON instances(status);

-- Create messages table
CREATE TABLE IF NOT EXISTS messages (
    id VARCHAR(255) PRIMARY KEY,
    instance_id VARCHAR(255) NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    remote_jid VARCHAR(255) NOT NULL,
    from_me BOOLEAN DEFAULT FALSE,
    message TEXT,
    message_type VARCHAR(50),
    media_url TEXT,
    caption TEXT,
    mime_type VARCHAR(100),
    file_name VARCHAR(255),
    quoted_id VARCHAR(255),
    quoted_message TEXT,
    view_once BOOLEAN DEFAULT FALSE,
    timestamp TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) DEFAULT 'sent',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index on instance_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_messages_instance_id ON messages(instance_id);

-- Create index on remote_jid for faster lookups
CREATE INDEX IF NOT EXISTS idx_messages_remote_jid ON messages(remote_jid);

-- Create index on timestamp for sorting
CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);

-- Create contacts table
CREATE TABLE IF NOT EXISTS contacts (
    id VARCHAR(255) PRIMARY KEY,
    instance_id VARCHAR(255) NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    remote_jid VARCHAR(255) NOT NULL,
    push_name VARCHAR(255),
    profile_pic_url TEXT,
    is_on_whatsapp BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(instance_id, remote_jid)
);

-- Create index on instance_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_contacts_instance_id ON contacts(instance_id);

-- Create index on remote_jid for faster lookups
CREATE INDEX IF NOT EXISTS idx_contacts_remote_jid ON contacts(remote_jid);

-- Create chats table
CREATE TABLE IF NOT EXISTS chats (
    id VARCHAR(255) PRIMARY KEY,
    instance_id VARCHAR(255) NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    remote_jid VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    unread_messages INTEGER DEFAULT 0,
    is_group BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(instance_id, remote_jid)
);

-- Create index on instance_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_chats_instance_id ON chats(instance_id);

-- Create index on remote_jid for faster lookups
CREATE INDEX IF NOT EXISTS idx_chats_remote_jid ON chats(remote_jid);

-- Create trigger to update updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to instances table
CREATE TRIGGER update_instances_updated_at BEFORE UPDATE ON instances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Apply trigger to contacts table
CREATE TRIGGER update_contacts_updated_at BEFORE UPDATE ON contacts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Apply trigger to chats table
CREATE TRIGGER update_chats_updated_at BEFORE UPDATE ON chats
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
