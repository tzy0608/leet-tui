-- 题目缓存
CREATE TABLE IF NOT EXISTS problems (
    id              INTEGER PRIMARY KEY,
    title           TEXT NOT NULL,
    title_slug      TEXT NOT NULL UNIQUE,
    frontend_id     TEXT NOT NULL,
    difficulty      TEXT NOT NULL,
    content         TEXT,
    topic_tags      TEXT,
    code_snippets   TEXT,
    is_paid_only    INTEGER DEFAULT 0,
    ac_rate         REAL,
    site            TEXT DEFAULT 'us',
    status          TEXT DEFAULT NULL,
    fetched_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 题目标签多对多
CREATE TABLE IF NOT EXISTS problem_tags (
    problem_id INTEGER REFERENCES problems(id),
    tag        TEXT NOT NULL,
    PRIMARY KEY (problem_id, tag)
);

-- 提交记录
CREATE TABLE IF NOT EXISTS submissions (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    problem_id   INTEGER REFERENCES problems(id),
    lang         TEXT NOT NULL,
    code         TEXT NOT NULL,
    status       TEXT NOT NULL,
    runtime_ms   INTEGER,
    memory_kb    INTEGER,
    submitted_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- FSRS 复习卡片
CREATE TABLE IF NOT EXISTS review_cards (
    problem_id     INTEGER PRIMARY KEY REFERENCES problems(id),
    due            DATETIME NOT NULL,
    stability      REAL NOT NULL DEFAULT 0,
    difficulty     REAL NOT NULL DEFAULT 0,
    elapsed_days   REAL NOT NULL DEFAULT 0,
    scheduled_days REAL NOT NULL DEFAULT 0,
    reps           INTEGER NOT NULL DEFAULT 0,
    lapses         INTEGER NOT NULL DEFAULT 0,
    state          INTEGER NOT NULL DEFAULT 0,
    last_review    DATETIME
);

-- 复习日志
CREATE TABLE IF NOT EXISTS review_logs (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    problem_id     INTEGER REFERENCES problems(id),
    rating         INTEGER NOT NULL,
    state          INTEGER NOT NULL,
    elapsed_days   REAL,
    scheduled_days REAL,
    time_spent_sec INTEGER,
    reviewed_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 学习计划
CREATE TABLE IF NOT EXISTS study_plans (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT NOT NULL,
    slug          TEXT NOT NULL UNIQUE,
    description   TEXT,
    is_predefined INTEGER DEFAULT 0,
    is_active     INTEGER DEFAULT 0,
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 计划中的题目
CREATE TABLE IF NOT EXISTS study_plan_problems (
    plan_id      INTEGER REFERENCES study_plans(id),
    problem_id   INTEGER REFERENCES problems(id),
    day_number   INTEGER NOT NULL,
    topic_group  TEXT,
    sort_order   INTEGER DEFAULT 0,
    is_completed INTEGER DEFAULT 0,
    completed_at DATETIME,
    PRIMARY KEY (plan_id, problem_id)
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_problems_difficulty ON problems(difficulty);
CREATE INDEX IF NOT EXISTS idx_problems_site ON problems(site);
CREATE INDEX IF NOT EXISTS idx_problem_tags_tag ON problem_tags(tag);
CREATE INDEX IF NOT EXISTS idx_review_cards_due ON review_cards(due);
CREATE INDEX IF NOT EXISTS idx_review_cards_state ON review_cards(state);
CREATE INDEX IF NOT EXISTS idx_submissions_problem ON submissions(problem_id);
CREATE INDEX IF NOT EXISTS idx_study_plan_problems_plan ON study_plan_problems(plan_id);
