CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    source_type VARCHAR(10) NOT NULL,
    preview TEXT,
    chunk_count INT DEFAULT 0,
    created_at BIGINT NOT NULL
);
CREATE INDEX idx_documents_business_id ON documents(business_id);

CREATE TABLE chat_sessions (
    session_id VARCHAR(255) PRIMARY KEY,
    business_id VARCHAR(255) NOT NULL,
    onboarding_state VARCHAR(30) NOT NULL DEFAULT 'none'
);
CREATE INDEX idx_chat_sessions_business_id ON chat_sessions(business_id);

CREATE TABLE model_settings (
    business_id VARCHAR(255) PRIMARY KEY,
    model_key VARCHAR(100) NOT NULL
);

CREATE TABLE chat_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id VARCHAR(255) NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    model_key VARCHAR(100),
    tokens_used INT DEFAULT 0,
    created_at BIGINT NOT NULL
);
CREATE INDEX idx_chat_logs_business_id ON chat_logs(business_id);
CREATE INDEX idx_chat_logs_session_id ON chat_logs(session_id);
CREATE INDEX idx_chat_logs_created_at ON chat_logs(created_at);

CREATE TABLE admin_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    business_id VARCHAR(255) NOT NULL,
    created_at BIGINT NOT NULL
);
CREATE INDEX idx_admin_users_business_id ON admin_users(business_id);