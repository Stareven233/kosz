package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MojiDict struct {
	url      string
	header   http.Header
	query    map[string]interface{}
	path     string
	db       *gorm.DB
	words    string
	commands map[string]bool
}

type MojiWord struct {
	Excerpt string `json:"excerpt"`
	Spell   string `json:"spell"`
	Accent  string `json:"accent"`
	Pron    string `json:"pron"`
	//Romaji string `json:"romaji"`
	//CreatedAt time.Time `json:"createdAt"`
	//UpdatedAt time.Time `json:"updatedAt"`
	//UpdatedBy string `json:"updatedBy"`
	ObjectID   string    `json:"objectId" gorm:"primarykey"`
	Count      uint      `json:"count" gorm:"default:1"`
	SearchedAt time.Time `json:"searchedAt" gorm:"autoCreateTime;index"`
}

type MojiResult struct {
	Result struct {
		OriginalSearchText string `json:"originalSearchText"`
		SearchResults      []struct {
			SearchText string    `json:"searchText"`
			Count      int       `json:"count"`
			TarID      string    `json:"tarId"`
			Title      string    `json:"title"`
			Type       int       `json:"type"`
			CreatedAt  time.Time `json:"createdAt"`
			UpdatedAt  time.Time `json:"updatedAt"`
			ObjectID   string    `json:"objectId"`
		} `json:"searchResults"`
		// 实际上确定没用到的字段去掉不会影响解析
		Words []MojiWord `json:"words"`
	} `json:"result"`
}

//当不确定json格式的时候可以使用mapstructure库

func NewMojiDict(path string) MojiDict {
	h := http.Header{}
	h.Add("Content-Type", "application/json;charset=UTF-8")
	h.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "+
		"(KHTML, like Gecko) Chrome/90.0.4430.85 Safari/537.36 Edg/90.0.818.49")

	//db, _ := gorm.Open(sqlite.Open("f:\\code\\trans_meow.db"), &gorm.Config{})
	db, _ := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	_ = db.AutoMigrate(&MojiWord{})

	return MojiDict{
		url: "https://api.mojidict.com/parse/functions/search_v3",
		query: map[string]interface{}{
			"langEnv":         "zh-CN_ja",
			"needWords":       true,
			"searchText":      "",
			"_ApplicationId":  "E62VyFVLMiW7kvbtVq3p",
			"_ClientVersion":  "js2.12.0",
			"_InstallationId": "7d959a18-48c4-243c-7486-632147466544",
		},
		header: h,
		path:   path,
		db:     db,
		commands: map[string]bool{
			"list":   true,
			"del":    true,
			"detail": true,
			"help":   true,
		},
	}
}

func (moji MojiDict) Request(word string) MojiWord {
	//var errors []error
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("%+v\n", err)
		}
	}()
	moji.query["searchText"] = word
	bytesData, _ := json.Marshal(moji.query)
	body := bytes.NewReader(bytesData)
	request, _ := http.NewRequest("POST", moji.url, body)
	request.Header = moji.header
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		//panic("error: 接口无响应")
		panic(errors.Wrap(err, ""))
	}
	respBytes, _ := io.ReadAll(resp.Body)
	//fmt.Println(string(respBytes))

	var mojiResult MojiResult
	_ = json.Unmarshal(respBytes, &mojiResult)
	//re := fmt.Sprintf("%v", mojiResult)
	//参考最下方的原始响应格式

	w := mojiResult.Result.Words
	if len(w) == 0 {
		panic("error: 查无单词")
	}
	moji.db.Create(&w[0])
	return w[0]
}

func (moji MojiDict) Search(word string) (w MojiWord) {
	moji.db.Where("spell = ?", word).Or("pron = ?", word).First(&w)
	if w.ObjectID != "" {
		w.Count++
		moji.db.Save(w)
	}
	return
}

func (moji MojiDict) Command(cmd string) (word string) {
	c := strings.SplitN(cmd, " ", 3)
	lenC, mainC := len(c), c[0]
	if !moji.commands[mainC] {
		return mainC
	} else if lenC < 2 {
		return ""
	}
	switch mainC {
	case "list":
		num, _ := strconv.ParseInt(c[1], 10, 0)
		num = Min(100, num)
		var words []MojiWord
		order := "desc"
		if lenC == 3 && c[2] == "asc" {
			order = "asc"
		}
		moji.db.Select("spell").Order("searched_at " + order).Limit(int(num)).Find(&words)
		wordStr := strings.Builder{}
		for _, w := range words {
			wordStr.WriteString(fmt.Sprintf("    %s\n", w.Spell))
		}
		println(wordStr.String())
	case "del":
		wordStr := strings.Join(c[1:], "\",\"")
		moji.db.Delete(MojiWord{}, fmt.Sprintf(`spell in ("%s")`, wordStr))
		// 批量：DELETE FROM "main"."moji_words" WHERE spell in ("スタミナ", "宥める")
	case "detail":
		res := moji.Search(c[1])
		fmt.Printf("    %+v\n", res)
	case "help":
		// 应用help ?总之加个参数，不然无法通过前面的检查
		print("    commands: ")
		for k, _ := range moji.commands {
			fmt.Printf("%s ", k)
		}
		println()
	}
	return ""
	//说明指令执行成功，不需再作为单词去查询
}
