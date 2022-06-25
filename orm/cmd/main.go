package main

import (
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pedrogao/log"
	"github.com/pedrogao/orm"
)

func main() {
	engine, err := orm.NewEngine("sqlite3", "test.db")
	if err != nil {
		log.Fatal(err)
	}

	defer engine.Close()

	s := engine.NewSession()
	_, _ = s.Raw("DROP TABLE IF EXISTS User;").Exec()
	_, _ = s.Raw("CREATE TABLE User(Name text);").Exec()
	_, _ = s.Raw("CREATE TABLE User(Name text);").Exec()
	result, _ := s.Raw("INSERT INTO User(`Name`) values (?), (?)", "Tom", "Sam").Exec()
	count, _ := result.RowsAffected()
	fmt.Printf("Exec success, %d affected\n", count)
}
