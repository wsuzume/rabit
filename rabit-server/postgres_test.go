package main

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var db *sql.DB

// テスト前にPostgreSQLに接続し、Pingをテスト
func TestMain(m *testing.M) {
	// PostgreSQLの接続情報
	connStr := "user=rabit password=password dbname=rabit_db sslmode=disable"

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("データベース接続エラー: %v", err)
	}

	// deferでDB接続をクローズ
	defer db.Close()

	// 接続確認
	err = db.Ping()
	if err != nil {
		log.Fatalf("データベースに接続できませんでした: %v", err)
	}

	// テストの実行
	code := m.Run()

	// 終了コードを返す
	os.Exit(code)
}


func TestTodoOperations(t *testing.T) {
	// レコードを挿入
	_, err := db.Exec("INSERT INTO todos (id, react_key, title) VALUES (1, 100, 'First Todo'), (2, 200, 'Second Todo')")
	if err != nil {
		t.Fatalf("レコード挿入エラー: %v", err)
	}
	t.Log("レコードを挿入しました")

	// 挿入したレコードを取得
	rows, err := db.Query("SELECT id, react_key, title FROM todos")
	if err != nil {
		t.Fatalf("クエリ実行エラー: %v", err)
	}
	defer rows.Close()

	var todos []TodoRecord
	for rows.Next() {
		var todo TodoRecord
		if err := rows.Scan(&todo.ID, &todo.ReactKey, &todo.Title); err != nil {
			t.Fatalf("レコード取得エラー: %v", err)
		}
		todos = append(todos, todo)
	}
	if err = rows.Err(); err != nil {
		t.Fatalf("クエリエラー: %v", err)
	}

	// 結果の出力と確認
	if len(todos) == 0 {
		t.Fatalf("レコードが存在しません")
	}
	t.Logf("取得したレコード数: %d", len(todos))
	for _, todo := range todos {
		t.Logf("ID: %d, ReactKey: %d, Title: %s", todo.ID, todo.ReactKey, todo.Title)
	}

	// 挿入したレコードを削除
	_, err = db.Exec("DELETE FROM todos WHERE id IN (1, 2)")
	if err != nil {
		t.Fatalf("レコード削除エラー: %v", err)
	}
	t.Log("挿入したレコードを削除しました")
}
