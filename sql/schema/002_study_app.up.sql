-- Ativar suporte a Foreign Keys no SQLite (Obrigatório em cada conexão)
PRAGMA foreign_keys = ON;

-- 1. SUBJECTS (Matérias)
CREATE TABLE subjects (
    id TEXT PRIMARY KEY, -- UUID v4 or v7
    name TEXT NOT NULL,
    color_hex TEXT, -- Para UI (ex: #FF5733)
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at TEXT -- Soft delete
);

-- 2. TOPICS (Assuntos)
CREATE TABLE topics (
    id TEXT PRIMARY KEY,
    subject_id TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at TEXT,
    FOREIGN KEY (subject_id) REFERENCES subjects(id) ON DELETE CASCADE
);

-- 3. STUDY_CYCLES (Ciclos de Estudo)
CREATE TABLE study_cycles (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    is_active INTEGER DEFAULT 0, -- Boolean (0 or 1)
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at TEXT
);

-- 4. CYCLE_ITEMS (Itens do Ciclo)
CREATE TABLE cycle_items (
    id TEXT PRIMARY KEY,
    cycle_id TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    order_index INTEGER NOT NULL, -- Posição no Round Robin
    planned_duration_minutes INTEGER DEFAULT 60, -- Meta de tempo
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (cycle_id) REFERENCES study_cycles(id) ON DELETE CASCADE,
    FOREIGN KEY (subject_id) REFERENCES subjects(id) ON DELETE CASCADE
);

-- 5. STUDY_SESSIONS (Sessões de Estudo)
CREATE TABLE study_sessions (
    id TEXT PRIMARY KEY,
    subject_id TEXT NOT NULL,
    cycle_item_id TEXT,
    
    started_at TEXT NOT NULL, -- ISO8601
    finished_at TEXT,
    
    gross_duration_seconds INTEGER DEFAULT 0,
    net_duration_seconds INTEGER DEFAULT 0,
    
    notes TEXT,
    
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    
    FOREIGN KEY (subject_id) REFERENCES subjects(id),
    FOREIGN KEY (cycle_item_id) REFERENCES cycle_items(id)
);

-- 6. SESSION_PAUSES (Pausas detalhadas)
CREATE TABLE session_pauses (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    started_at TEXT NOT NULL,
    ended_at TEXT,
    duration_seconds INTEGER GENERATED ALWAYS AS (strftime('%s', ended_at) - strftime('%s', started_at)) VIRTUAL,
    FOREIGN KEY (session_id) REFERENCES study_sessions(id) ON DELETE CASCADE
);

-- 7. EXERCISE_LOGS (Registro de Questões)
CREATE TABLE exercise_logs (
    id TEXT PRIMARY KEY,
    session_id TEXT,
    subject_id TEXT NOT NULL,
    topic_id TEXT,
    
    questions_count INTEGER NOT NULL CHECK (questions_count >= 0),
    correct_count INTEGER NOT NULL CHECK (correct_count >= 0),
    
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    
    CONSTRAINT valid_score CHECK (correct_count <= questions_count),
    
    FOREIGN KEY (session_id) REFERENCES study_sessions(id) ON DELETE SET NULL,
    FOREIGN KEY (subject_id) REFERENCES subjects(id),
    FOREIGN KEY (topic_id) REFERENCES topics(id)
);

-- Indexes
CREATE INDEX idx_topics_subject ON topics(subject_id);
CREATE INDEX idx_cycle_items_cycle ON cycle_items(cycle_id);
CREATE INDEX idx_sessions_subject ON study_sessions(subject_id);
CREATE INDEX idx_sessions_date ON study_sessions(started_at);
CREATE INDEX idx_exercises_subject ON exercise_logs(subject_id);
CREATE INDEX idx_exercises_session ON exercise_logs(session_id);
