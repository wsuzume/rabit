package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

type Todo struct {
	Key     int `json:"key"`
	Content struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
	} `json:"content"`
}

type TodoRecord struct {
	ID    int
	Key   int
	Title string
}

// データベース保存用の構造体に変換する
func (t *Todo) ToTodoRecord() TodoRecord {
	return TodoRecord{
		ID:    t.Content.ID,
		Key:   t.Key,
		Title: t.Content.Title,
	}
}

func (r *TodoRecord) ToTodo() Todo {
	return Todo{
		Key: r.Key,
		Content: struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
		}{
			ID:    r.ID,
			Title: r.Title,
		},
	}
}

func (r *TodoRecord) ToQueryString() string {
	return fmt.Sprintf("(%d, %d, '%s')", r.ID, r.Key, r.Title)
}

// データベース
var db *sql.DB

func main() {
	e := echo.New()

	// フロントエンドの提供
	e.File("/", "dist/index.html")
	e.Static("/assets", "dist/assets")

	// ToDo のリストを取得する
	e.GET("/api/todos", func(c echo.Context) error {
		var err error

		// データベースに接続
		connStr := "user=rabit password=password dbname=rabit_db sslmode=disable"
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			fmt.Println("データベース接続エラー: %v", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Internal Server Error",
			})
		}

		// defer でDB接続をクローズ
		defer db.Close()

		// 挿入したレコードを取得
		query := "SELECT id, react_key, title FROM todos"
		rows, err := db.Query(query)
		if err != nil {
			fmt.Println("クエリ実行エラー: %v", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Internal Server Error",
			})
		}

		// defer で読み取りをクローズ
		defer rows.Close()

		// レコードを構造体に格納する
		var todos []Todo = []Todo{}  // テーブルが空のとき nil ではなく空配列を返すため初期化しておく
		for rows.Next() {
			var record TodoRecord
			if err := rows.Scan(&record.ID, &record.Key, &record.Title); err != nil {
				fmt.Println("レコード取得エラー: %v", err)
				return c.JSON(http.StatusInternalServerError, echo.Map{
					"message": "Internal Server Error",
				})
			}
			todos = append(todos, record.ToTodo())
		}

		if err = rows.Err(); err != nil {
			fmt.Println("クエリエラー: %v", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Internal Server Error",
			})
		}

		fmt.Println("正常にデータを取得しました")

		return c.JSON(http.StatusOK, todos)
	})

	// ToDo のリストを更新する
	e.PUT("/api/todos", func(c echo.Context) error {
		var err error
		var todos []Todo = []Todo{} 
		var todoQueries []string

		// 受け取った JSON を構造体にバインドする
		if err := c.Bind(&todos); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"message": "Invalid JSON",
			})
		}

		// 受け取ったデータをクエリに変換する
		for _, todo := range todos {
			record := todo.ToTodoRecord()
			todoQueries = append(todoQueries, record.ToQueryString())
		}

		records := strings.Join(todoQueries, ", ")
		query := fmt.Sprintf("INSERT INTO todos (id, react_key, title) VALUES %s", records)

		// データベースに接続
		connStr := "user=rabit password=password dbname=rabit_db sslmode=disable"

		db, err = sql.Open("postgres", connStr)
		if err != nil {
			fmt.Println("データベース接続エラー: %v", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Internal Server Error",
			})
		}

		// deferでDB接続をクローズ
		defer db.Close()

		// テーブル内の全レコードを削除
		_, err = db.Exec("DELETE FROM todos")
		if err != nil {
			fmt.Println("レコード削除エラー: %v", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Internal Server Error",
			})
		}

		// 受け取った Todo リストが空なら挿入ここで終了
		if len(todos) == 0 {
			fmt.Println("挿入するレコードがありません")
			return c.JSON(http.StatusOK, todos)
		}

		// レコードを挿入
		_, err = db.Exec(query)
		if err != nil {
			fmt.Println("レコード挿入エラー: %v", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Internal Server Error",
			})
		}

		fmt.Println("レコードを挿入しました")

		// 更新された ToDo リストを返す
		return c.JSON(http.StatusOK, todos)
	})

	// id が一致する ToDo アイテムを削除する
	e.DELETE("/api/todos", func(c echo.Context) error {
		var err error
		var todosToDelete []Todo
		var todoIDToDelete []int

		// 受け取った JSON を構造体にバインドする
		if err = c.Bind(&todosToDelete); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"message": "Invalid JSON",
			})
		}

		// 一致する ID の Todo を削除
		for _, todoToDelete := range todosToDelete {
			todoIDToDelete = append(todoIDToDelete, todoToDelete.Content.ID)
		}
    	
		// todoIDToDelete をカンマ区切りの文字列に変換
		var idList []string
		for _, id := range todoIDToDelete {
			idList = append(idList, fmt.Sprintf("%d", id))
		}
		idString := strings.Join(idList, ", ")

		// レコードを削除
		query := fmt.Sprintf("DELETE FROM todos WHERE id IN (%s)", idString)

		connStr := "user=rabit password=password dbname=rabit_db sslmode=disable"

		db, err = sql.Open("postgres", connStr)
		if err != nil {
			fmt.Println("データベース接続エラー: %v", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Internal Server Error",
			})
		}

		// deferでDB接続をクローズ
		defer db.Close()

		// 挿入したレコードを削除
		_, err = db.Exec(query)
		if err != nil {
			fmt.Println("レコード削除エラー: %v", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Internal Server Error",
			})
		}

		// todoIDToDelete の中身を出力 (標準出力)
    	fmt.Println("削除対象のID:", query)
		
		// 削除後
		return c.JSON(http.StatusOK, echo.Map{
			"message": "Success",
		})
	})

	e.Logger.Fatal(e.Start(":1323"))
}
