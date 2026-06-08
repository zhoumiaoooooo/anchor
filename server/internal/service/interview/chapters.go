package interview

import "anchor-server/internal/model"

type ChapterDef struct {
	Chapter   model.Chapter
	Title     string
	Opening   string   // AI's first message when chapter starts
	Questions []string // preset questions to navigate through
}

var ChapterDefs = map[model.Chapter]ChapterDef{
	model.ChapterHand: {
		Chapter: model.ChapterHand,
		Title:   "手",
		Opening: "跟我说说 ta 的手吧。大小、形状、有没有茧。你闭上眼睛能看见吗？",
		Questions: []string{
			"这双手每天在做什么？不是 ta 做什么工作——是 ta 手上的动作。择菜怎么择的？数钱怎么数的？",
			"这双手最厉害的时候做过什么？拎两桶水上五楼？绣一朵花？把你从河里捞起来？",
			"ta 生气的时候手在干嘛？攥着？指着你？不敢看你，只盯着自己的手指？",
			"你能做一遍 ta 最常做的那个动作吗？闭上眼睛，做一遍。像 ta 吗？",
			"ta 的手和你自己的手——有什么像、有什么不像？",
			"你小时候 ta 用手帮你做过什么？现在还会做吗？",
		},
	},
	model.ChapterVoice: {
		Chapter: model.ChapterVoice,
		Title:   "声音",
		Opening: "ta 叫你的名字。你现在脑子里能听到吗？是哪个字最重？是升调还是降调？",
		Questions: []string{
			"ta 笑的时候是什么声？哈哈，嘿嘿，还是不出声只抖肩膀？",
			"ta 说重话的时候用什么词？方言词——普通话有时候没那个味道。",
			"有没有一首歌 ta 反复唱？在哪唱——做饭？洗澡？骑自行车？走调吗？走哪句？",
			"你现在能学一下 ta 叫你的声音吗？录下来。不像没关系。",
			"ta 哭过吗？你见过几次？什么声音？",
			"你觉得你说话像 ta 吗？哪句话最像？",
		},
	},
	model.ChapterPlace: {
		Chapter: model.ChapterPlace,
		Title:   "地方",
		Opening: "ta 在哪长大的？那条街叫什么？",
		Questions: []string{
			"那个地方的早晨是什么声音？什么味道？",
			"ta 现在还会提起那里吗？提起的时候是笑的还是安静的？",
			"ta 这辈子最想回去的地方是哪里？回去了吗？",
			"如果有一天你能带 ta 回去——站在那个地方，你最想看什么？",
			"ta 这辈子住过几个地方？哪个地方 ta 最像自己？",
		},
	},
	model.ChapterThatDay: {
		Chapter: model.ChapterThatDay,
		Title:   "那一天",
		Opening: "选一天。哪一天都行。大的——结婚、生了第一个孩子、拿了什么奖。小的——一个普通的星期三下午，什么都没发生，但你就是记得。",
		Questions: []string{
			"那天早上是怎么开始的？什么光？什么温度？",
			"ta 在做什么？穿了什么？说了什么？",
			"那天 ta 说的哪句话你记住了？",
			"那天结束的时候——有什么你当时没在意，后来才发现你一直记得？",
			"如果让你回到那天做一件事，你会做什么？",
		},
	},
	model.ChapterOneThing: {
		Chapter:   model.ChapterOneThing,
		Title:     "还有件事",
		Opening:   "有些东西你可能从来没跟 ta 说。不是不想说。是没找到时候，或者不知道怎么开口。\n\n写下来。存着。永远不会被自动发出去。\n\n你想什么时候拿出来，或者永远不拿出来，都可以。它在。",
		Questions: []string{}, // No questions - this chapter just accepts one message and closes
	},
}
