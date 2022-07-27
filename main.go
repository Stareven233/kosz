package main

// go build  -o tiny_moji.exe ./
// ./tiny_moji.exe

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	_ = exec.Command("cmd", "/c", "title tiny moji v2.4.2").Run()
	moji := NewMojiDict("F:/CODE/Go/translate_meow_go/local_moji.db")
	defer moji.db.Close()

	var req string
	var res MojiWord
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">> ")
		//_, _ = fmt.Scanln(&req)  // Scanln跟Scan一样遇到空格断开
		data, _, _ := reader.ReadLine()
		req = string(data)
		req = moji.Command(req)
		if req == "exit" && moji.commands[req] {
			break
		}
		// 没有输入、输入控制字符都忽略
		if reqLen := len(req); reqLen == 0 || reqLen == 1 && req[0] < 128 {
			continue
		}
		//fmt.Printf("the input is:  ***%s***\n", req)

		// 先从数据库中寻找现有的
		res = moji.Search(req)
		if res.ObjectID == "" {
			// 本地找不到就联网查找
			res = moji.Request(req)
		}
		if res.ObjectID == "" {
			// 未找到单词
			fmt.Printf("    %+v\n", res)
			continue
		} else if req != res.Spell {
			// 找到跟输入不一致的单词
			fmt.Printf("warning: searched '%s' but found '%s'\n", req, res.Spell)
		}
		// 格式化输出找到的单词
		fmt.Printf("\t%s %s\n\tcnt: %d\n\t%s\n", res.Pron, res.Accent, res.Count, res.Excerpt)
	}
}

//接口原响应：
//{"result":{
//	"originalSearchText":"りんご",
//	"searchResults":[{
//		"searchText":"りんご",
//		"count":3847,
//		"tarId":"198928453",
//		"title":"",
//		"type":0,
//		"createdAt":"2019-05-07T02:53:12.456Z",
//		"updatedAt":"2021-06-25T17:50:22.734Z",
//		"objectId":"z6LtDQTuyo"
//	}],
//	"words":[{
//		"excerpt":"[名詞] 苹果。（バラ科の落葉高木。また、その果実。葉は卵円形。4、5月ごろ、葉とともに白または淡紅",
//		"spell":"林檎",
//		"accent":"◎",
//		"pron":"りんご",
//		"romaji":"ringo",
//		"createdAt":"2019-05-07T03:48:00.874Z",
//		"updatedAt":"2020-08-27T10:28:48.286Z",
//		"updatedBy":"isX02DXFUN","objectId":"198928453"}]
//	}
//}
