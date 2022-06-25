package main

import (
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pedrogao/log"
	"github.com/pedrogao/orm"
	"github.com/pedrogao/web"
)

func main() {
	app := web.New()

	storage, err := orm.NewEngine("sqlite3", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		storage.Close()
	}()

	s := storage.NewSession()
	_, err = s.Raw("DROP TABLE IF EXISTS User;").Exec()
	if err != nil {
		log.Errorf("drop table User err: %s", err)
	}

	_, err = s.Raw("CREATE TABLE User(Name text);").Exec()
	if err != nil {
		log.Errorf("create table User err: %s", err)
	}
	s.Clear()

	app.GET("/", func(ctx *web.Context) {
		ctx.String(http.StatusOK, "simple & easy")
	})

	app.POST("/users", func(ctx *web.Context) {
		name := ctx.Query("name")
		session := storage.NewSession()
		_, err := session.Raw("INSERT INTO User(`Name`) values (?)", name).Exec()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, map[string]any{"message": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, map[string]any{"message": "ok"})
	})

	app.GET("/users", func(ctx *web.Context) {
		name := ctx.Query("name")
		session := storage.NewSession()
		row := session.Raw("SELECT * FROM User WHERE Name = ?", name).QueryRow()

		var ret string
		err = row.Scan(&ret)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, map[string]any{"message": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, map[string]any{"name": ret})
	})

	if err = app.Run(":3000"); err != nil {
		log.Fatalf("start server err: %s", err)
	}
}
