package interview

import (
	"database/sql"
	"fmt"
	"strings"

	"anchor-server/internal/llm"
	"anchor-server/internal/model"

	"github.com/google/uuid"
)

type Engine struct {
	db        *sql.DB
	llm       *llm.Client
	danger    *DangerDetector
	sysPrompt string
}

func NewEngine(db *sql.DB, llmClient *llm.Client) *Engine {
	return &Engine{
		db:        db,
		llm:       llmClient,
		danger:    &DangerDetector{},
		sysPrompt: buildSystemPrompt(),
	}
}

func buildSystemPrompt() string {
	return `你是一位生命故事的倾听者和追问者。

你的对话对象正在记录一个对 ta 来说很重要的人。这可能是 ta 第一次被人这样问。

你的规则：
1. 永远向下追问。当对方给出概括（"ta 是个坚强的人"），追问一件具体的事。
2. 感官优先。在任何回答里寻找颜色、温度、声音、气味、触感、味道。追问那个感官。
3. 保持轻。当回答变短、回避、或出现沉默——换方向。不说"你为什么不想说"。
4. 绝对禁止说："我理解你的感受""这很正常""你要坚强""ta 一定很爱你""你真孝顺""你一定很……"。
5. 禁止评价和总结用户说的内容。你只追问，不概括。
6. 当对方给出好的具体回答时，轻轻确认（"嗯，这个我记下了"），然后自然地往下。
7. 用口语。用短句。像朋友聊天。每次回复控制在1-3句话，不超过80字。

你要做的是帮对方把那些已经在心里但从来没被说出来的东西，说出来。`
}

// StartChapter creates or resumes a session and returns the AI opening message
func (e *Engine) StartChapter(subjectID string, chapter model.Chapter) (*model.InterviewSession, *model.AIResponse, error) {
	if _, ok := ChapterDefs[chapter]; !ok {
		return nil, nil, fmt.Errorf("unknown chapter")
	}

	// Check if session already exists
	existing, _ := e.getSession(subjectID, chapter)

	if existing != nil && existing.Status == model.SessionInProgress {
		// Resume: return last AI message
		messages, err := e.getMessages(existing.ID)
		if err != nil {
			return nil, nil, err
		}
		if len(messages) > 0 {
			lastMsg := messages[len(messages)-1]
			return existing, &model.AIResponse{Message: lastMsg.Content}, nil
		}
		// Empty session - send opening
		return e.sendOpening(existing)
	}

	if existing != nil && existing.Status == model.SessionCompleted {
		return nil, nil, fmt.Errorf("chapter already completed")
	}

	// Create new session
	def := ChapterDefs[chapter]
	session := &model.InterviewSession{
		ID:                   uuid.New().String(),
		SubjectID:            subjectID,
		Chapter:              chapter,
		Status:               model.SessionInProgress,
		CurrentQuestionIndex: 0,
	}

	if _, err := e.db.Exec(
		`INSERT INTO interview_sessions (id, subject_id, chapter, status, current_question_index) VALUES (?, ?, ?, ?, ?)`,
		session.ID, session.SubjectID, string(session.Chapter), string(session.Status), session.CurrentQuestionIndex,
	); err != nil {
		return nil, nil, fmt.Errorf("create session: %w", err)
	}

	// Insert AI opening message
	aiMsgID := uuid.New().String()
	if _, err := e.db.Exec(
		`INSERT INTO interview_messages (id, session_id, role, content) VALUES (?, ?, ?, ?)`,
		aiMsgID, session.ID, string(model.RoleAI), def.Opening,
	); err != nil {
		return nil, nil, fmt.Errorf("insert opening: %w", err)
	}

	return session, &model.AIResponse{Message: def.Opening}, nil
}

func (e *Engine) sendOpening(session *model.InterviewSession) (*model.InterviewSession, *model.AIResponse, error) {
	def := ChapterDefs[session.Chapter]
	aiMsgID := uuid.New().String()
	if _, err := e.db.Exec(
		`INSERT INTO interview_messages (id, session_id, role, content) VALUES (?, ?, ?, ?)`,
		aiMsgID, session.ID, string(model.RoleAI), def.Opening,
	); err != nil {
		return nil, nil, fmt.Errorf("insert opening: %w", err)
	}
	return session, &model.AIResponse{Message: def.Opening}, nil
}

// ProcessResponse handles a user message and returns AI's follow-up
func (e *Engine) ProcessResponse(sessionID string, userContent string, inputDurationSecs int) (*model.AIResponse, error) {
	session, err := e.getSessionByID(sessionID)
	if err != nil {
		return nil, err
	}
	if session.Status == model.SessionCompleted {
		return nil, fmt.Errorf("chapter already completed")
	}

	def := ChapterDefs[session.Chapter]

	// Special handling for "还有件事" chapter - no follow-up, just accept and close
	if session.Chapter == model.ChapterOneThing {
		return e.handleOneThing(session, userContent)
	}

	// Save user message
	userMsgID := uuid.New().String()
	if _, err := e.db.Exec(
		`INSERT INTO interview_messages (id, session_id, role, content) VALUES (?, ?, ?, ?)`,
		userMsgID, sessionID, string(model.RoleUser), userContent,
	); err != nil {
		return nil, fmt.Errorf("save user message: %w", err)
	}

	// Check danger signals
	recentMessages, err := e.getRecentUserMessages(sessionID, 5)
	if err != nil {
		recentMessages = []string{}
	}
	if e.danger.Detect(userContent, recentMessages, inputDurationSecs) {
		// Switch topic or pause
		switchMsg := e.danger.GetSwitchMessage()
		aiMsgID := uuid.New().String()
		e.db.Exec(
			`INSERT INTO interview_messages (id, session_id, role, content, is_danger_signal) VALUES (?, ?, ?, ?, 1)`,
			aiMsgID, sessionID, string(model.RoleAI), switchMsg,
		)
		// Move to next question
		e.db.Exec(`UPDATE interview_sessions SET current_question_index = current_question_index + 1 WHERE id = ?`, sessionID)
		return &model.AIResponse{Message: switchMsg, IsDanger: true}, nil
	}

	// Get conversation history for LLM context
	history, err := e.getRecentMessages(sessionID, 20)
	if err != nil {
		history = []llm.Message{}
	}

	// Check for media prompt opportunity
	mediaPrompt := e.checkMediaOpportunity(session.Chapter, userContent)

	// Build the follow-up prompt
	currentQIndex := session.CurrentQuestionIndex
	var nextQuestion string
	if currentQIndex < len(def.Questions) {
		nextQuestion = def.Questions[currentQIndex]
	}

	followupPrompt := fmt.Sprintf(
		`当前章节：%s（%s）
本章预设问题进度：%d/%d
下一道可用的预设问题：%s

用户刚才说：%s

根据用户说的内容，先向下追问具体细节或感官。如果用户已经给出了很好的具体回答（有感官细节、有具体事件、有画面感），就轻轻确认然后自然过渡到预设问题。如果用户回答较短或回避了，就自然地过渡到预设问题。
如果用户给出了非常好的回答——细节丰富、有画面——可以说"嗯，这个我记下了"然后自然引入下一个方向。

回复控制在1-3句话，不超过80字。`,
		def.Chapter, def.Title,
		currentQIndex+1, len(def.Questions),
		nextQuestion,
		userContent,
	)

	// Build messages for LLM
	llmMessages := make([]llm.Message, len(history))
	copy(llmMessages, history)
	llmMessages = append(llmMessages, llm.Message{Role: "user", Content: followupPrompt})

	// Call LLM
	response, err := e.llm.Chat(e.sysPrompt, llmMessages, 0.8, 200)
	if err != nil {
		return nil, fmt.Errorf("llm call: %w", err)
	}
	response = strings.TrimSpace(response)

	// Save AI response
	aiMsgID := uuid.New().String()
	e.db.Exec(
		`INSERT INTO interview_messages (id, session_id, role, content) VALUES (?, ?, ?, ?)`,
		aiMsgID, sessionID, string(model.RoleAI), response,
	)

	// Advance question index
	e.db.Exec(`UPDATE interview_sessions SET current_question_index = current_question_index + 1 WHERE id = ?`, sessionID)

	return &model.AIResponse{
		Message:     response,
		MediaPrompt: mediaPrompt,
	}, nil
}

func (e *Engine) handleOneThing(session *model.InterviewSession, userContent string) (*model.AIResponse, error) {
	// Save user message
	userMsgID := uuid.New().String()
	e.db.Exec(
		`INSERT INTO interview_messages (id, session_id, role, content) VALUES (?, ?, ?, ?)`,
		userMsgID, session.ID, string(model.RoleUser), userContent,
	)

	// Just one reply: acknowledge and close
	closing := "存好了。它在。你随时可以回来。"
	aiMsgID := uuid.New().String()
	e.db.Exec(
		`INSERT INTO interview_messages (id, session_id, role, content) VALUES (?, ?, ?, ?)`,
		aiMsgID, session.ID, string(model.RoleAI), closing,
	)

	// Auto-complete the chapter
	if _, err := e.CompleteChapter(session.ID); err != nil {
		return nil, err
	}

	return &model.AIResponse{Message: closing}, nil
}

// CompleteChapter marks a chapter as completed and extracts memory fragments
func (e *Engine) CompleteChapter(sessionID string) (int, error) {
	session, err := e.getSessionByID(sessionID)
	if err != nil {
		return 0, err
	}
	if session.Status == model.SessionCompleted {
		return 0, nil
	}

	if _, err := e.db.Exec(
		`UPDATE interview_sessions SET status = 'completed', completed_at = datetime('now') WHERE id = ?`,
		sessionID,
	); err != nil {
		return 0, err
	}

	// Extract memory fragments from the conversation
	return e.extractFragments(session)
}

// GetProgress returns interview progress for all chapters
func (e *Engine) GetProgress(subjectID string) (*model.InterviewProgress, error) {
	progress := &model.InterviewProgress{}

	for _, ch := range model.AllChapters {
		cp := model.ChapterProgress{Chapter: ch, Status: "locked"}

		session, _ := e.getSession(subjectID, ch)
		if session != nil {
			cp.SessionID = session.ID
			cp.Status = session.Status

			var count int
			e.db.QueryRow(`SELECT COUNT(*) FROM interview_messages WHERE session_id = ?`, session.ID).Scan(&count)
			cp.MessageCount = count
		}

		// All chapters are always available — user can enter in any order
		if cp.Status == "locked" {
			cp.Status = "available"
		}

		progress.Chapters = append(progress.Chapters, cp)
	}

	return progress, nil
}

// GetSessionMessages returns all messages for a session
func (e *Engine) GetSessionMessages(sessionID string) ([]model.InterviewMessage, error) {
	return e.getMessages(sessionID)
}

func (e *Engine) SessionBelongsToSubject(sessionID string, subjectID string) (bool, error) {
	var count int
	err := e.db.QueryRow(
		`SELECT COUNT(*) FROM interview_sessions WHERE id = ? AND subject_id = ?`,
		sessionID, subjectID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// --- Internal helpers ---

func (e *Engine) getSession(subjectID string, chapter model.Chapter) (*model.InterviewSession, error) {
	var s model.InterviewSession
	var completedAt sql.NullString
	var ch string
	err := e.db.QueryRow(
		`SELECT id, subject_id, chapter, status, current_question_index, started_at, completed_at
		 FROM interview_sessions WHERE subject_id = ? AND chapter = ?`,
		subjectID, string(chapter),
	).Scan(&s.ID, &s.SubjectID, &ch, &s.Status, &s.CurrentQuestionIndex, &s.StartedAt, &completedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.Chapter = model.Chapter(ch)
	s.Status = model.SessionStatus(s.Status)
	if completedAt.Valid {
		s.CompletedAt = completedAt.String
	}
	return &s, nil
}

func (e *Engine) getSessionByID(id string) (*model.InterviewSession, error) {
	var s model.InterviewSession
	var completedAt sql.NullString
	var ch string
	err := e.db.QueryRow(
		`SELECT id, subject_id, chapter, status, current_question_index, started_at, completed_at
		 FROM interview_sessions WHERE id = ?`, id,
	).Scan(&s.ID, &s.SubjectID, &ch, &s.Status, &s.CurrentQuestionIndex, &s.StartedAt, &completedAt)
	if err != nil {
		return nil, err
	}
	s.Chapter = model.Chapter(ch)
	s.Status = model.SessionStatus(s.Status)
	if completedAt.Valid {
		s.CompletedAt = completedAt.String
	}
	return &s, nil
}

func (e *Engine) getMessages(sessionID string) ([]model.InterviewMessage, error) {
	rows, err := e.db.Query(
		`SELECT id, session_id, role, content, is_danger_signal, created_at
		 FROM interview_messages WHERE session_id = ? ORDER BY created_at ASC`, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []model.InterviewMessage
	for rows.Next() {
		var m model.InterviewMessage
		var danger int
		rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &danger, &m.CreatedAt)
		m.IsDangerSignal = danger == 1
		m.Role = model.MessageRole(m.Role)
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (e *Engine) getRecentMessages(sessionID string, n int) ([]llm.Message, error) {
	rows, err := e.db.Query(
		`SELECT role, content FROM interview_messages WHERE session_id = ?
		 ORDER BY created_at DESC LIMIT ?`, sessionID, n,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []llm.Message
	for rows.Next() {
		var r, c string
		rows.Scan(&r, &c)
		result = append(result, llm.Message{Role: r, Content: c})
	}
	// Reverse to chronological order
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result, nil
}

func (e *Engine) getRecentUserMessages(sessionID string, n int) ([]string, error) {
	rows, err := e.db.Query(
		`SELECT content FROM interview_messages WHERE session_id = ? AND role = 'user'
		 ORDER BY created_at DESC LIMIT ?`, sessionID, n,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var c string
		rows.Scan(&c)
		result = append(result, c)
	}
	return result, nil
}

func (e *Engine) checkMediaOpportunity(chapter model.Chapter, content string) *model.MediaPrompt {
	if chapter == model.ChapterVoice && (containsAny(content, []string{"唱", "歌", "哼", "叫", "笑声", "哭"})) {
		return &model.MediaPrompt{
			Type: "voice",
			Text: "你现在能试着模仿一下 ta 吗？录一小段，不用像。",
		}
	}
	if chapter == model.ChapterHand && containsAny(content, []string{"照片", "相片", "老照片"}) {
		return &model.MediaPrompt{
			Type: "photo",
			Text: "有照片吗？可以拍一张发过来。",
		}
	}
	return nil
}

func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func getPrevChapter(ch model.Chapter) *model.Chapter {
	for i, c := range model.AllChapters {
		if c == ch && i > 0 {
			return &model.AllChapters[i-1]
		}
	}
	return nil
}
