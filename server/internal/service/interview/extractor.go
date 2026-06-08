package interview

import (
	"encoding/json"
	"fmt"
	"strings"

	"anchor-server/internal/model"

	"github.com/google/uuid"
)

// extractFragments analyzes interview messages and creates memory fragments
func (e *Engine) extractFragments(session *model.InterviewSession) (int, error) {
	messages, err := e.getMessages(session.ID)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, msg := range messages {
		if msg.Role != model.RoleUser || msg.IsDangerSignal {
			continue
		}
		if len([]rune(msg.Content)) < 15 {
			continue
		}

		fragment := e.buildFragment(session.SubjectID, session.Chapter, msg)
		if fragment == nil {
			continue
		}

		if err := e.saveFragment(fragment); err != nil {
			continue // skip bad fragments
		}
		count++
	}

	return count, nil
}

func (e *Engine) buildFragment(subjectID string, chapter model.Chapter, msg model.InterviewMessage) *model.MemoryFragment {
	content := msg.Content

	fragment := &model.MemoryFragment{
		ID:              uuid.New().String(),
		SubjectID:       subjectID,
		Chapter:         chapter,
		SourceMessageID: msg.ID,
		RawText:         content,
	}

	// Extract sensory tags
	sensory := extractSensoryTags(content)
	b, _ := json.Marshal(sensory)
	fragment.SensoryTags = string(b)

	// Extract people
	people := extractPeople(content)
	b, _ = json.Marshal(people)
	fragment.PeopleTags = string(b)

	// Extract time references
	times := extractTimeTags(content)
	b, _ = json.Marshal(times)
	fragment.TimeTags = string(b)

	// Extract places
	places := extractPlaces(content)
	b, _ = json.Marshal(places)
	fragment.PlaceTags = string(b)

	// Extract emotions
	emotions := extractEmotions(content)
	b, _ = json.Marshal(emotions)
	fragment.EmotionTags = string(b)

	// Check for music
	fragment.HasMusic = containsAny(content, []string{"歌", "唱", "哼", "曲", "音乐", "调", "听", "唱机", "收音机", "邓丽君", "周璇"})

	// Check for procedural memory
	fragment.IsProcedural = containsAny(content, []string{"每天", "每次", "总是", "一直", "天天", "习惯了", "从来都是"})

	// Calculate anchor potential
	fragment.AnchorPotential = calcAnchorPotential(fragment)

	return fragment
}

func (e *Engine) saveFragment(f *model.MemoryFragment) error {
	_, err := e.db.Exec(
		`INSERT INTO memory_fragments
		 (id, subject_id, chapter, source_message_id, raw_text, polished_text,
		  sensory_tags, people_tags, time_tags, place_tags, emotion_tags,
		  anchor_potential, is_procedural, has_music)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		f.ID, f.SubjectID, string(f.Chapter), f.SourceMessageID, f.RawText, f.PolishedText,
		f.SensoryTags, f.PeopleTags, f.TimeTags, f.PlaceTags, f.EmotionTags,
		f.AnchorPotential, boolToInt(f.IsProcedural), boolToInt(f.HasMusic),
	)
	return err
}

func boolToInt(b bool) int {
	if b { return 1 }
	return 0
}

// --- Tag extraction (rule-based for MVP) ---

func extractSensoryTags(text string) map[string][]string {
	tags := map[string][]string{
		"sight": {}, "sound": {}, "smell": {}, "taste": {}, "touch": {},
	}

	sightWords := []string{"红", "蓝", "绿", "黄", "白", "黑", "灰", "暗", "亮", "光", "大", "小", "高", "矮", "胖", "瘦", "老", "年轻", "好看", "丑"}
	soundWords := []string{"唱", "说", "笑", "哭", "喊", "叫", "响", "吵", "安静", "歌", "声", "嗓", "音", "哼"}
	smellWords := []string{"香", "臭", "味道", "气味", "闻", "烟", "饭", "菜", "花"}
	tasteWords := []string{"甜", "酸", "苦", "辣", "咸", "淡", "好吃", "难吃", "烫", "凉", "热", "冷"}
	touchWords := []string{"软", "硬", "凉", "热", "冷", "暖", "烫", "扎", "刺", "滑", "粗", "细", "茧", "手", "摸", "碰", "抱", "拍", "握"}

	for _, w := range sightWords { if strings.Contains(text, w) { tags["sight"] = append(tags["sight"], w) } }
	for _, w := range soundWords { if strings.Contains(text, w) { tags["sound"] = append(tags["sound"], w) } }
	for _, w := range smellWords { if strings.Contains(text, w) { tags["smell"] = append(tags["smell"], w) } }
	for _, w := range tasteWords { if strings.Contains(text, w) { tags["taste"] = append(tags["taste"], w) } }
	for _, w := range touchWords { if strings.Contains(text, w) { tags["touch"] = append(tags["touch"], w) } }

	return tags
}

func extractPeople(text string) []string {
	roles := []string{"妈妈", "爸爸", "外婆", "外公", "奶奶", "爷爷", "姐姐", "妹妹", "哥哥", "弟弟", "阿姨", "叔叔", "舅舅", "姑姑", "朋友", "同事", "邻居", "老师", "同学", "我"}
	var found []string
	for _, r := range roles {
		if strings.Contains(text, r) {
			found = append(found, r)
		}
	}
	return found
}

func extractTimeTags(text string) []map[string]interface{} {
	var times []map[string]interface{}

	years := extractYears(text)
	for _, y := range years {
		t := map[string]interface{}{"year": y}
		if strings.Contains(text, "冬") { t["season"] = "冬" }
		if strings.Contains(text, "夏") { t["season"] = "夏" }
		if strings.Contains(text, "春") { t["season"] = "春" }
		if strings.Contains(text, "秋") { t["season"] = "秋" }
		if strings.Contains(text, "早上") || strings.Contains(text, "早晨") { t["time_of_day"] = "早上" }
		if strings.Contains(text, "晚上") { t["time_of_day"] = "晚上" }
		if strings.Contains(text, "下午") { t["time_of_day"] = "下午" }
		times = append(times, t)
	}
	return times
}

func extractYears(text string) []int {
	var years []int
	// Simple: find 4-digit numbers that look like years
	for i := 0; i < len(text)-3; i++ {
		if text[i] == '1' || text[i] == '2' {
			sub := text[i : i+4]
			n := 0
			fmt.Sscanf(sub, "%d", &n)
			if n >= 1900 && n <= 2030 {
				years = append(years, n)
			}
		}
	}
	return years
}

func extractPlaces(text string) []string {
	placeWords := []string{"家", "厂", "学校", "医院", "街", "路", "村", "城", "市", "省", "店", "门口", "门口", "河边", "田", "山上", "车间", "办公室", "食堂", "院子"}
	var found []string
	for _, p := range placeWords {
		if strings.Contains(text, p) {
			found = append(found, p)
		}
	}
	return found
}

func extractEmotions(text string) []map[string]interface{} {
	emotionMap := map[string]string{
		"开心": "开心", "高兴": "高兴", "笑": "开心", "乐": "开心",
		"难过": "难过", "哭": "难过", "伤心": "难过", "泪": "难过",
		"生气": "生气", "气": "生气", "怒": "生气",
		"怕": "害怕", "害怕": "害怕", "担心": "害怕",
		"想": "思念", "想她": "思念", "想他": "思念", "想念": "思念",
		"骄傲": "骄傲", "自豪": "骄傲",
		"辛苦": "辛苦", "累": "辛苦", "不容易": "辛苦",
	}
	var emotions []map[string]interface{}
	for keyword, emotion := range emotionMap {
		if strings.Contains(text, keyword) {
			intensity := 3
			if strings.Contains(text, "非常") || strings.Contains(text, "特别") || strings.Contains(text, "最") || strings.Contains(text, "太") {
				intensity = 5
			}
			emotions = append(emotions, map[string]interface{}{"emotion": emotion, "intensity": intensity})
		}
	}
	return emotions
}

func calcAnchorPotential(f *model.MemoryFragment) int {
	score := 0
	if f.HasMusic { score += 3 }
	// Check for high emotion
	var emotions []map[string]interface{}
	json.Unmarshal([]byte(f.EmotionTags), &emotions)
	for _, e := range emotions {
		if intensity, ok := e["intensity"].(float64); ok && intensity >= 4 {
			score += 2
			break
		}
	}
	if f.IsProcedural { score += 2 }
	var sensory map[string][]string
	json.Unmarshal([]byte(f.SensoryTags), &sensory)
	if len(sensory["sound"]) > 0 { score += 1 }
	if len(sensory["taste"]) > 0 { score += 1 }
	if len(sensory["smell"]) > 0 { score += 1 }
	if len(sensory["touch"]) > 0 { score += 1 }
	if score > 10 { score = 10 }
	if score < 0 { score = 0 }
	return score
}
