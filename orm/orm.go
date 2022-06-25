package orm

import (
	"database/sql"
	"fmt"

	"github.com/pedrogao/log"
	"github.com/pedrogao/orm/session"
)

type Engine struct {
	db *sql.DB
}

func NewEngine(driver, source string) (e *Engine, err error) {
	var db *sql.DB
	db, err = sql.Open(driver, source)
	if err != nil {
		log.Errorf("open err: %s", err)
		return
	}

	if err = db.Ping(); err != nil {
		log.Errorf("ping err: %s", err)
		return
	}

	e = &Engine{db: db}
	return
}

func (e *Engine) Close() error {
	if err := e.db.Close(); err != nil {
		log.Errorf("close orm engine err: %s", err)
		return fmt.Errorf("close orm engine err: %s", err)
	}
	return nil
}

func (e *Engine) NewSession() *session.Session {
	return session.New(e.db)
}
