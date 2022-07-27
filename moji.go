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

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type MojiDict struct {
	url      string
	header   http.Header
	query    map[string]interface{}
	path     string
	db       *sql.DB
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
	ObjectID   string    `json:"objectId"`
	Count      uint      `json:"count"`
	SearchedAt time.Time `json:"searchedAt"`
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
	h.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.0.0 Safari/537.36")

	db, err := sql.Open("sqlite3", path)
	checkErr(err)
	// 若不存在则建表
	sqlStmt := `CREATE TABLE IF NOT EXISTS moji_words(
		excerpt text,
		spell   text,
		accent  text,
		pron    text,
		object_id   text PRIMARY KEY NOT NULL,
		count      integer,
		searched_at datatime
	);`
	_, err = db.Exec(sqlStmt)
	checkErr(err)

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
			"exit":   true,
		},
	}
}

// 构造http请求并从moji上查找单词
func (moji MojiDict) Request(word string) MojiWord {
	//var errors []error
	defer showErrMsg()

	// 设置moji搜索单词的post请求内容
	moji.query["searchText"] = word
	bytesData, _ := json.Marshal(moji.query)
	body := bytes.NewReader(bytesData)
	request, _ := http.NewRequest("POST", moji.url, body)
	request.Header = moji.header
	resp, err := http.DefaultClient.Do(request)
	checkErr(err)

	// 获取并解析响应数据
	respBytes, _ := io.ReadAll(resp.Body)
	//fmt.Println(string(respBytes))
	var Result MojiResult
	_ = json.Unmarshal(respBytes, &Result)
	//re := fmt.Sprintf("%v", Result)
	//参考最下方的原始响应格式

	// 若查到单词则写入数据库
	words := Result.Result.Words
	if len(words) == 0 {
		panic("warning: 查无单词")
	}
	// moji.db.Create(&words[0])
	stmt, err := moji.db.Prepare("insert into moji_words(excerpt, spell, accent, pron, object_id, count, searched_at) values(?,?,?,?,?,?,?)")
	checkErr(err)
	defer stmt.Close()

	t := time.Now()
	w := words[0]
	_, err = stmt.Exec(w.Excerpt, w.Spell, w.Accent, w.Pron, w.ObjectID, w.Count, t)
	checkErr(err)
	return w
}

// 在本地数据库中查找单词，若不存在则返回对象的字段都是空字符串
func (moji MojiDict) Search(word string) (w MojiWord) {
	// moji.db.Where("spell = ?", word).Or("pron = ?", word).First(&w)
	// 选出拼写或其读音为word的单词
	querySql := fmt.Sprintf(`select * from moji_words where spell="%s" or pron="%s"`, word, word)
	row := moji.db.QueryRow(querySql)
	var tmp string
	// 将数据读入MojiWord对象
	err := row.Scan(&w.Excerpt, &w.Spell, &w.Accent, &w.Pron, &w.ObjectID, &w.Count, &tmp)
	if err != nil {
		// 说明不存在这个单词，返回空MojiWord，ObjectID为""
		return
	}
	// 日期需要手动转换为time.Time对象
	w.SearchedAt, err = time.Parse(time.RFC3339Nano, tmp)
	checkErr(err)

	// 存在该单词，更新查找次数count
	if w.ObjectID != "" {
		w.Count++
		_, err = moji.db.Exec(fmt.Sprintf(`update moji_words set count = %d where object_id = "%s"`, w.Count, w.ObjectID))
		checkErr(err)
	}
	return
}

// 将输入的一行解析为指令执行
func (moji MojiDict) Command(cmd string) (word string) {
	defer showErrMsg()
	c := strings.SplitN(cmd, " ", 3)
	lenC, mainC := len(c), c[0]

	if !moji.commands[mainC] {
		// 不是指令，返回作为单词去查询
		return mainC
	}

	switch mainC {
	// 列出查询过的单词，支持按时间升序降序
	case "list":
		if lenC < 2 {
			panic("warning: list须指定查询的数量")
		}
		num, _ := strconv.ParseInt(c[1], 10, 0)
		num = Min(100, num)

		// 默认降序查询
		order := "desc"
		if lenC == 3 && c[2] == "asc" {
			order = "asc"
		}

		// moji.db.Select("spell").Order("searched_at " + order).Limit(int(num)).Find(&words)
		rows, err := moji.db.Query(fmt.Sprintf("select spell from moji_words order by searched_at %s limit %d", order, num))
		checkErr(err)
		defer rows.Close()

		// 挨个取出单词，再一起输出
		var spell string
		wordStr := strings.Builder{}
		for rows.Next() {
			err = rows.Scan(&spell)
			checkErr(err)
			wordStr.WriteString(fmt.Sprintf("    %s\n", spell))
		}
		println(wordStr.String())

	// 删除空格隔开的所有单词
	case "del":
		// 将待删除的单词拼合成 s1","s2,..."sn
		wordStr := strings.Join(c[1:], "\",\"")
		// moji.db.Delete(MojiWord{}, fmt.Sprintf(`spell in ("%s")`, wordStr))
		_, err := moji.db.Exec(fmt.Sprintf(`delete from moji_words where spell in ("%s")`, wordStr))
		if err == nil {
			fmt.Println("info: 删除成功")
		}
		// 批量：delete from moji_words where spell in ("スタミナ", "宥める")

	// 输出某一单词的详细信息
	case "detail":
		res := moji.Search(c[1])
		fmt.Printf("    %+v\n", res)

	// 输出所有可用指令
	case "help":
		print("    commands: ")
		for k := range moji.commands {
			fmt.Printf("%s ", k)
		}
		println()

	// 退出程序
	case "exit":
		return "exit"
	}
	return ""
	//说明指令执行成功，不需再作为单词去查询
}
