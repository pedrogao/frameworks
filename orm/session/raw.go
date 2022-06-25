package session

import (
	"database/sql"
	"strings"

	"github.com/pedrogao/log"
)

type Session struct {
	db      *sql.DB
	sql     strings.Builder
	sqlVars []any
}

func New(db *sql.DB) *Session {
	return &Session{db: db}
}

func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
}

func (s *Session) DB() *sql.DB {
	return s.db
}

func (s *Session) Raw(sql string, values ...any) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, values...)
	return s
}

func (s *Session) Exec() (result sql.Result, err error) {
	defer s.Clear()

	log.Infof("exec: %s %+v", s.sql.String(), s.sqlVars)
	if result, err = s.DB().Exec(s.sql.String(), s.sqlVars...); err != nil {
		log.Errorf("exec err: %s", err)
	}
	return
}

func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()

	log.Infof("query row: %s %+v", s.sql.String(), s.sqlVars)
	return s.DB().QueryRow(s.sql.String(), s.sqlVars...)
}

func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	defer s.Clear()

	log.Infof("query rows: %s %+v", s.sql.String(), s.sqlVars)
	if rows, err = s.DB().Query(s.sql.String(), s.sqlVars...); err != nil {
		log.Errorf("query rows err: %s", err)
	}
	return
}
