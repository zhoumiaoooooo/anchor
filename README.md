# 锚点

为所爱之人，留下记忆的锚。

---

有人老了。记忆开始松脱，像旧墙上的钉子，一颗一颗往下掉。

你害怕那些事情不再被任何人记得——ta 手的温度、叫你名字的方式、削苹果时总是削成兔子形状。这些太细小了，小到不值得写进传记，却大到撑起了一个人。

**锚点**是一个安静的地方。AI 会问你几个问题，你来说，它帮你整理。就像一个耐心的朋友，不问那些沉重的，只问：你还记得什么？

记录不是为了告别。是为了证明我们爱过。

---

## 它是怎么工作的

打开锚点，告诉它你想记录谁。然后有五扇门：

- **手** — 大小、动作、茧、温度
- **声音** — 叫你名字的方式、笑声、唱的歌
- **地方** — ta 长大的街道、气味、记忆里的空间
- **那一天** — 大的日子也好，普通的下午也好
- **还有件事** — 想说但没说的话，只存一条，不追问

每一扇门都可以随时推开。AI 会引导你说下去，你说的每一句话都会被整理成记忆片段，存进你的记忆库。

**AI 只负责提问和整理。它不替代陪伴，也不制造告别。**

---

## 在自己的电脑上运行

### 你需要准备

- 一台电脑（Windows / Mac / Linux 都可以）
- [Go](https://go.dev/dl/) 1.22 或更高版本
- 一个 [DeepSeek](https://platform.deepseek.com/) API key（[在这里获取](https://platform.deepseek.com/api_keys)，新用户有免费额度）

### 步骤

```bash
# 1. 下载项目
git clone https://github.com/zhoumiaoooooo/anchor.git
cd anchor/server

# 2. 设置 API key
# Windows PowerShell:
$env:DEEPSEEK_API_KEY="sk-你的key"

# Mac / Linux:
export DEEPSEEK_API_KEY="sk-你的key"

# 3. 启动
go run ./cmd/server
```

打开浏览器，访问 **http://localhost:8080**。

所有数据都存储在你电脑上的 SQLite 文件里，不上传任何第三方服务器（除了调用 DeepSeek API 进行对话）。

### 可选配置

| 环境变量 | 默认值 | 说明 |
|---------|--------|-----|
| `PORT` | `8080` | 服务端口 |
| `DEEPSEEK_API_KEY` | — | **必填**，DeepSeek API 密钥 |
| `DEEPSEEK_BASE_URL` | `https://api.deepseek.com` | API 地址 |
| `DEEPSEEK_MODEL` | `deepseek-chat` | 模型名称 |
| `DATABASE_PATH` | `./data/anchor.db` | 数据库文件路径 |

---

## 技术

Go 后端 + 纯 HTML/CSS/JS 前端 + SQLite + DeepSeek API。

没有框架，没有构建工具，没有 npm。前端是单个 HTML 文件，后端是 Go 标准库 HTTP server。

---

## 关于

这个项目的起点是阿尔茨海默症家庭。但我们都知道，记忆不只是被疾病带走——时间本身就在做这件事。

锚点的意思是：在海水淹没之前，先把锚抛下去。

*记录不是为了告别，而是为了证明我们爱过。*
