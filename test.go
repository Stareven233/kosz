package main

// go build  -o t.exe .\test.go .\types.go .\utils.go | ./t.exe
// go build  -o t.exe .\tiny_moji.go .\test.go .\utils.go | ./t.exe

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func test_sqlite() {
	db, err := sql.Open("sqlite3", "./nya_dict.db")
	CheckErr(err)
	defer db.Close()

	sqlStmt := `CREATE TABLE IF NOT EXISTS moji_words(
		excerpt text,
		spell   text,
		accent  text,
		pron    text,
		object_id   text PRIMARY KEY NOT NULL,
		count      integer,
		searched_at datetime
	);`
	_, err = db.Exec(sqlStmt)
	CheckErr(err)

	// 插入
	// stmt, err := db.Prepare("INSERT INTO moji_words(excerpt, spell, accent, pron, object_id, count, searched_at) values(?,?,?,?,?,?,?)")
	// CheckErr(err)
	// t := time.Now()
	// _, err = stmt.Exec("方法", "213", "df", "dsf", "a1", 79, t)
	// CheckErr(err)
	// t = time.Now()
	// _, err = stmt.Exec("是否为", "ert", "321", "656", "a2", 322, t)
	// CheckErr(err)
	// stmt.Close()

	// 查询
	// num := 5
	// order := "desc"
	// var spell string
	// querySql := fmt.Sprintf("select spell from moji_words order by searched_at %s limit %d", order, num)
	// fmt.Println(querySql)
	// rows, err := db.Query(querySql)
	// CheckErr(err)
	// defer rows.Close()
	// for rows.Next() {
	// 	err = rows.Scan(&spell)
	// 	CheckErr(err)
	// 	fmt.Println(spell)
	// }

	// 删除
	// delSql := fmt.Sprintf(`delete from moji_words where spell in ("%s")`, "ttt")
	// fmt.Println(delSql)
	// _, err = db.Exec(delSql)
	// CheckErr(err)

	// 更新
	// _, err = db.Exec(fmt.Sprintf(`update moji_words set count = %d where object_id = "%s"`, 0, "a2"))
	// CheckErr(err)

	// 测试修改单一行
	// word := "dfwe"
	// querySql := fmt.Sprintf(`select * from moji_words where spell="%s" or pron="%s"`, word, word)
	// row := db.QueryRow(querySql)
	// var tmp string
	// // 将数据读入MojiWord对象
	// w := MojiWord{}
	// _ = row.Scan(&w.Excerpt, &w.Spell, &w.Accent, &w.Pron, &w.Object_ID, &w.Count, &tmp)
	// // fmt.Printf("    %+v\n", w)
	// fmt.Println(w.Object_ID == "")
}
