package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/text/encoding/simplifiedchinese"
)

type MojiDict struct {
	url      string
	header   http.Header
	query    map[string]interface{}
	path     string
	db       *sql.DB
	words    string
	commands map[string]string
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
	CheckErr(err)
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
	CheckErr(err)

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
		commands: map[string]string{
			"list":   "列出查询过的单词, list num (asc/desc)",
			"del":    "删除某些单词, del word1 word2 ...",
			"detail": "列出某个单词的详细信息, detail word",
			"help":   "打印当前显示的帮助信息, help",
			"q":      "退出程序，exit的简写方法, q",
			"exit":   "退出程序, exit",
			"gse":    "利用gse将列出的所有句子分词, gse sentence1 sentence2 ...",
			"fgs":    "调用fugashi将句子分词并获取每个词的信息, a sentence",
		},
	}
}

// 构造http请求并从moji上查找单词
func (moji MojiDict) Request(word string) MojiWord {
	//var errors []error
	defer ShowErrMsg()

	// 设置moji搜索单词的post请求内容
	moji.query["searchText"] = word
	bytesData, _ := json.Marshal(moji.query)
	body := bytes.NewReader(bytesData)
	request, _ := http.NewRequest("POST", moji.url, body)
	request.Header = moji.header
	resp, err := http.DefaultClient.Do(request)
	CheckErr(err)

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
	CheckErr(err)
	defer stmt.Close()

	t := time.Now()
	w := words[0]
	_, err = stmt.Exec(w.Excerpt, w.Spell, w.Accent, w.Pron, w.ObjectID, w.Count, t)
	CheckErr(err)
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
	CheckErr(err)

	// 存在该单词，更新查找次数count
	if w.ObjectID != "" {
		w.Count++
		_, err = moji.db.Exec(fmt.Sprintf(`update moji_words set count = %d where object_id = "%s"`, w.Count, w.ObjectID))
		CheckErr(err)
	}
	return
}

// 将输入的一行解析为指令执行
func (moji MojiDict) Command(cmd string) (word string) {
	defer ShowErrMsg()
	c := strings.SplitN(cmd, " ", 3)
	lenC, mainC := len(c), c[0]

	if _, exist := moji.commands[mainC]; !exist {
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
		CheckErr(err)
		defer rows.Close()

		// 挨个取出单词，再一起输出
		var spell string
		wordStr := strings.Builder{}
		for rows.Next() {
			err = rows.Scan(&spell)
			CheckErr(err)
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
		word := moji.Search(c[1])
		// fmt.Printf("    %+v\n", word)
		res, err := json.MarshalIndent(word, "", "  ")
		CheckErr(err)
		fmt.Println(string(res))

	// 输出所有可用指令
	case "help":
		for k, v := range moji.commands {
			fmt.Printf("%s: %s\n", k, v)
		}
		println()

	// 退出程序
	case "exit":
		fallthrough
	case "q":
		return "exit"

	// 将句子分词
	case "gse":
		text := strings.Join(c[1:], " ")
		segs := TokenizeJp(text)
		fmt.Println(segs)

	case "fgs":
		// text := bytes.NewBufferString(c[1])
		// text := strings.NewReader(c[1])
		// cmd := exec.Command("python", "./analyse.py")
		// cmd.Stdin = text
		// // buffer只支持bytes，传过去总是乱码
		// result, err := cmd.CombinedOutput()

		cmd := exec.Command("python", "./analyse.py", c[1])
		cmd.Dir = "./tokenize"
		result, err := cmd.CombinedOutput()
		CheckErr(err)
		// result, err = unicode.UTF8.NewDecoder().Bytes(result)
		result, err = simplifiedchinese.GB18030.NewDecoder().Bytes(result)
		// japanese.ShiftJIS反而不行，可能与操作系统语言有关
		CheckErr(err)
		fmt.Println(string(result))
	}

	return ""
	//说明指令执行成功，不需再作为单词去查询
}
