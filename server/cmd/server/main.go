package main

import (
	"fmt"
	"log"
	"net/http"

	"anchor-server/internal/config"
	"anchor-server/internal/db"
	"anchor-server/internal/handler"
	"anchor-server/internal/llm"
	"anchor-server/internal/router"
	"anchor-server/internal/service/interview"
)

func main() {
	cfg := config.Load()

	// Database
	database, err := db.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer database.Close()
	log.Println("数据库已连接")

	// DeepSeek client
	llmClient := llm.New(cfg.DeepSeekAPIKey, cfg.DeepSeekBaseURL, cfg.DeepSeekModel)
	log.Println("DeepSeek 客户端已就绪")

	// Interview engine
	engine := interview.NewEngine(database, llmClient)
	log.Println("采访引擎已就绪")

	// HTTP handler
	h := handler.New(cfg, database, engine)

	// Router
	r := router.New(h)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("锚点服务启动: http://localhost%s", addr)
	log.Printf("API 文档:")
	log.Printf("  POST /api/v1/subjects         - 创建人物")
	log.Printf("  GET  /api/v1/subjects          - 人物列表")
	log.Printf("  GET  /api/v1/subjects/{id}/interviews/progress - 采访进度")
	log.Printf("  POST /api/v1/subjects/{id}/interviews/{chapter}/start - 开始采访")
	log.Printf("  POST /api/v1/subjects/{id}/interviews/respond         - 回复采访")
	log.Printf("  POST /api/v1/subjects/{id}/interviews/complete        - 完成章节")
	log.Printf("  GET  /api/v1/subjects/{id}/memories                   - 记忆列表")

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
