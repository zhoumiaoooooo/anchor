package interview

import "strings"

type DangerDetector struct{}

var switchMessages = []string{
	"好，我们换个方向。",
	"这一章先放一放。想聊的时候再回来。",
	"先不说这个。你累不累？",
	"我们聊点轻松的。",
}

func (d *DangerDetector) GetSwitchMessage() string {
	// Random selection - in production use crypto/rand
	return switchMessages[0]
}

func (d *DangerDetector) Detect(message string, recentUserMessages []string, inputDurationSecs int) bool {
	score := 0.0

	// 1. Explicit refusal
	if containsAnyStr(message, []string{"别提了", "不说这个", "太难受了", "不想说", "别问了"}) {
		score += 1.0
	}

	// 2. Length collapse
	if len(recentUserMessages) >= 3 {
		avgLen := 0
		for _, m := range recentUserMessages {
			avgLen += len([]rune(m))
		}
		avgLen /= len(recentUserMessages)
		if avgLen > 50 && len([]rune(message)) < 10 {
			score += 0.8
		}
	}

	// 3. Self-blame language
	if containsAnyStr(message, []string{"都怪我", "我对不起", "后悔", "没来得及", "欠她", "是我的错", "我不该"}) {
		score += 0.7
	}

	// 4. Long silence followed by short message
	if inputDurationSecs > 20 && len([]rune(message)) < 20 {
		score += 0.6
	}

	return score >= 0.8
}

func containsAnyStr(s string, substrs []string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
