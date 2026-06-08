package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func Open(dbPath string) (*sql.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS subjects (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			relationship TEXT NOT NULL,
			birth_year INTEGER DEFAULT 0,
			hometown TEXT DEFAULT '',
			avatar_path TEXT DEFAULT '',
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,

		`CREATE TABLE IF NOT EXISTS interview_sessions (
			id TEXT PRIMARY KEY,
			subject_id TEXT NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
			chapter TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'in_progress',
			current_question_index INTEGER NOT NULL DEFAULT 0,
			started_at TEXT NOT NULL DEFAULT (datetime('now')),
			completed_at TEXT
		)`,

		`CREATE TABLE IF NOT EXISTS interview_messages (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL REFERENCES interview_sessions(id) ON DELETE CASCADE,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			is_danger_signal INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,

		`CREATE TABLE IF NOT EXISTS memory_fragments (
			id TEXT PRIMARY KEY,
			subject_id TEXT NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
			chapter TEXT NOT NULL,
			source_message_id TEXT,
			raw_text TEXT NOT NULL,
			polished_text TEXT DEFAULT '',
			sensory_tags TEXT NOT NULL DEFAULT '{}',
			people_tags TEXT NOT NULL DEFAULT '[]',
			time_tags TEXT NOT NULL DEFAULT '[]',
			place_tags TEXT NOT NULL DEFAULT '[]',
			emotion_tags TEXT NOT NULL DEFAULT '[]',
			anchor_potential INTEGER NOT NULL DEFAULT 5,
			is_procedural INTEGER NOT NULL DEFAULT 0,
			has_music INTEGER NOT NULL DEFAULT 0,
			times_used INTEGER NOT NULL DEFAULT 0,
			positive_reactions INTEGER NOT NULL DEFAULT 0,
			total_reactions INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,

		`CREATE TABLE IF NOT EXISTS media_assets (
			id TEXT PRIMARY KEY,
			subject_id TEXT NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
			memory_fragment_id TEXT REFERENCES memory_fragments(id) ON DELETE SET NULL,
			type TEXT NOT NULL,
			file_path TEXT NOT NULL,
			original_filename TEXT DEFAULT '',
			mime_type TEXT DEFAULT '',
			file_size_bytes INTEGER DEFAULT 0,
			metadata TEXT NOT NULL DEFAULT '{}',
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,

		`CREATE TABLE IF NOT EXISTS anchor_packs (
			id TEXT PRIMARY KEY,
			subject_id TEXT NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
			memory_fragment_id TEXT NOT NULL REFERENCES memory_fragments(id),
			pack_type TEXT NOT NULL DEFAULT 'daily',
			image_path TEXT DEFAULT '',
			music_path TEXT DEFAULT '',
			music_title TEXT DEFAULT '',
			text_content TEXT NOT NULL,
			guide_question TEXT DEFAULT '',
			generated_versions TEXT NOT NULL DEFAULT '[]',
			selected_version_index INTEGER DEFAULT 0,
			quality_score REAL DEFAULT 0,
			pushed_at TEXT,
			opened_at TEXT,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,

		`CREATE TABLE IF NOT EXISTS reactions (
			id TEXT PRIMARY KEY,
			anchor_pack_id TEXT NOT NULL REFERENCES anchor_packs(id) ON DELETE CASCADE,
			reaction_type TEXT NOT NULL,
			note TEXT DEFAULT '',
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,

		`CREATE INDEX IF NOT EXISTS idx_sessions_subject ON interview_sessions(subject_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_session ON interview_messages(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_fragments_subject ON memory_fragments(subject_id)`,
		`CREATE INDEX IF NOT EXISTS idx_media_subject ON media_assets(subject_id)`,
		`CREATE INDEX IF NOT EXISTS idx_anchors_subject ON anchor_packs(subject_id, pushed_at)`,
		`CREATE INDEX IF NOT EXISTS idx_reactions_anchor ON reactions(anchor_pack_id)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("exec: %w\n%s", err, stmt)
		}
	}
	return nil
}
