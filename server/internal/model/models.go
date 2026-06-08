package model

// Time fields are strings (SQLite TEXT format: "2006-01-02 15:04:05")

type Subject struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Relationship string `json:"relationship"`
	BirthYear    int    `json:"birth_year,omitempty"`
	Hometown     string `json:"hometown,omitempty"`
	AvatarPath   string `json:"avatar_path,omitempty"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type Chapter string

const (
	ChapterHand     Chapter = "手"
	ChapterVoice    Chapter = "声音"
	ChapterPlace    Chapter = "地方"
	ChapterThatDay  Chapter = "那一天"
	ChapterOneThing Chapter = "还有件事"
)

var AllChapters = []Chapter{ChapterHand, ChapterVoice, ChapterPlace, ChapterThatDay, ChapterOneThing}

type SessionStatus string

const (
	SessionInProgress SessionStatus = "in_progress"
	SessionCompleted  SessionStatus = "completed"
)

type InterviewSession struct {
	ID                   string        `json:"id"`
	SubjectID            string        `json:"subject_id"`
	Chapter              Chapter       `json:"chapter"`
	Status               SessionStatus `json:"status"`
	CurrentQuestionIndex int           `json:"current_question_index"`
	StartedAt            string        `json:"started_at"`
	CompletedAt          string        `json:"completed_at,omitempty"`
}

type MessageRole string

const (
	RoleAI   MessageRole = "ai"
	RoleUser MessageRole = "user"
)

type InterviewMessage struct {
	ID             string      `json:"id"`
	SessionID      string      `json:"session_id"`
	Role           MessageRole `json:"role"`
	Content        string      `json:"content"`
	IsDangerSignal bool        `json:"is_danger_signal"`
	CreatedAt      string      `json:"created_at"`
}

type MemoryFragment struct {
	ID                string `json:"id"`
	SubjectID         string `json:"subject_id"`
	Chapter           Chapter `json:"chapter"`
	SourceMessageID   string `json:"source_message_id,omitempty"`
	RawText           string `json:"raw_text"`
	PolishedText      string `json:"polished_text,omitempty"`
	SensoryTags       string `json:"sensory_tags"`
	PeopleTags        string `json:"people_tags"`
	TimeTags          string `json:"time_tags"`
	PlaceTags         string `json:"place_tags"`
	EmotionTags       string `json:"emotion_tags"`
	AnchorPotential   int    `json:"anchor_potential"`
	IsProcedural      bool   `json:"is_procedural"`
	HasMusic          bool   `json:"has_music"`
	TimesUsed         int    `json:"times_used"`
	PositiveReactions int    `json:"positive_reactions"`
	TotalReactions    int    `json:"total_reactions"`
	CreatedAt         string `json:"created_at"`
}

type PackType string

const (
	PackDaily       PackType = "daily"
	PackSunset      PackType = "sunset"
	PackFirst       PackType = "first"
	PackSpecialDate PackType = "special_date"
)

type AnchorPack struct {
	ID                   string  `json:"id"`
	SubjectID            string  `json:"subject_id"`
	MemoryFragmentID     string  `json:"memory_fragment_id"`
	PackType             PackType `json:"pack_type"`
	ImagePath            string  `json:"image_path,omitempty"`
	MusicPath            string  `json:"music_path,omitempty"`
	MusicTitle           string  `json:"music_title,omitempty"`
	TextContent          string  `json:"text_content"`
	GuideQuestion        string  `json:"guide_question,omitempty"`
	GeneratedVersions    string  `json:"generated_versions,omitempty"`
	SelectedVersionIndex int     `json:"selected_version_index"`
	QualityScore         float64 `json:"quality_score"`
	PushedAt             string  `json:"pushed_at,omitempty"`
	OpenedAt             string  `json:"opened_at,omitempty"`
	CreatedAt            string  `json:"created_at"`
}

type ReactionType string

const (
	ReactionSmiled   ReactionType = "smiled"
	ReactionNothing  ReactionType = "nothing"
	ReactionSurprise ReactionType = "surprise"
)

type Reaction struct {
	ID           string       `json:"id"`
	AnchorPackID string       `json:"anchor_pack_id"`
	Type         ReactionType `json:"reaction_type"`
	Note         string       `json:"note,omitempty"`
	CreatedAt    string       `json:"created_at"`
}

// --- API DTOs ---

type CreateSubjectRequest struct {
	Name         string `json:"name"`
	Relationship string `json:"relationship"`
	BirthYear    int    `json:"birth_year,omitempty"`
	Hometown     string `json:"hometown,omitempty"`
}

type RespondRequest struct {
	SessionID string `json:"session_id"`
	Content   string `json:"content"`
}

type ReactRequest struct {
	ReactionType ReactionType `json:"reaction_type"`
	Note         string       `json:"note,omitempty"`
}

type InterviewProgress struct {
	Chapters []ChapterProgress `json:"chapters"`
}

type ChapterProgress struct {
	Chapter      Chapter       `json:"chapter"`
	Status       SessionStatus `json:"status"`
	SessionID    string        `json:"session_id,omitempty"`
	MessageCount int           `json:"message_count"`
}

type AIResponse struct {
	Message     string       `json:"message"`
	MediaPrompt *MediaPrompt `json:"media_prompt,omitempty"`
	IsDanger    bool         `json:"is_danger"`
}

type MediaPrompt struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
