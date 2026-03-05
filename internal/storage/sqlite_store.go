package storage

/*
#cgo LDFLAGS: -lsqlite3
#include <sqlite3.h>
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"fmt"
	"time"
	"unsafe"
)

var sqliteTransient = C.sqlite3_destructor_type(unsafe.Pointer(uintptr(^uintptr(0))))

type SQLiteStore struct {
	path string
	db   *C.sqlite3
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	s := &SQLiteStore{path: path}
	if err := s.open(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *SQLiteStore) open() error {
	cpath := C.CString(s.path)
	defer C.free(unsafe.Pointer(cpath))
	if rc := C.sqlite3_open(cpath, &s.db); rc != C.SQLITE_OK {
		return fmt.Errorf("sqlite open failed: %s", C.GoString(C.sqlite3_errmsg(s.db)))
	}
	return nil
}

func (s *SQLiteStore) Close() error {
	if s.db == nil {
		return nil
	}
	if rc := C.sqlite3_close(s.db); rc != C.SQLITE_OK {
		return fmt.Errorf("sqlite close failed: %s", C.GoString(C.sqlite3_errmsg(s.db)))
	}
	s.db = nil
	return nil
}

func (s *SQLiteStore) Init(ctx context.Context) error {
	_ = ctx
	q := `CREATE TABLE IF NOT EXISTS scenarios (
id TEXT PRIMARY KEY,
name TEXT NOT NULL,
content TEXT NOT NULL,
created_at TEXT NOT NULL
);`
	return s.exec(q)
}

func (s *SQLiteStore) SaveScenario(ctx context.Context, id, name, content string) error {
	_ = ctx
	stmt := `INSERT INTO scenarios(id,name,content,created_at) VALUES(?,?,?,?);`
	return s.execPrepared(stmt, id, name, content, time.Now().UTC().Format(time.RFC3339))
}

func (s *SQLiteStore) GetScenario(ctx context.Context, id string) (string, error) {
	_ = ctx
	stmtText := `SELECT content FROM scenarios WHERE id=? LIMIT 1;`
	cstmt := C.CString(stmtText)
	defer C.free(unsafe.Pointer(cstmt))

	var stmt *C.sqlite3_stmt
	if rc := C.sqlite3_prepare_v2(s.db, cstmt, -1, &stmt, nil); rc != C.SQLITE_OK {
		return "", fmt.Errorf("prepare failed: %s", C.GoString(C.sqlite3_errmsg(s.db)))
	}
	defer C.sqlite3_finalize(stmt)

	cid := C.CString(id)
	defer C.free(unsafe.Pointer(cid))
	C.sqlite3_bind_text(stmt, 1, cid, -1, sqliteTransient)

	if rc := C.sqlite3_step(stmt); rc == C.SQLITE_ROW {
		text := C.sqlite3_column_text(stmt, 0)
		if text == nil {
			return "", fmt.Errorf("scenario %s not found", id)
		}
		return C.GoString((*C.char)(unsafe.Pointer(text))), nil
	}
	return "", fmt.Errorf("scenario %s not found", id)
}

func (s *SQLiteStore) exec(query string) error {
	cquery := C.CString(query)
	defer C.free(unsafe.Pointer(cquery))
	var errMsg *C.char
	if rc := C.sqlite3_exec(s.db, cquery, nil, nil, &errMsg); rc != C.SQLITE_OK {
		defer C.sqlite3_free(unsafe.Pointer(errMsg))
		return fmt.Errorf("exec failed: %s", C.GoString(errMsg))
	}
	return nil
}

func (s *SQLiteStore) execPrepared(query string, params ...string) error {
	cquery := C.CString(query)
	defer C.free(unsafe.Pointer(cquery))
	var stmt *C.sqlite3_stmt
	if rc := C.sqlite3_prepare_v2(s.db, cquery, -1, &stmt, nil); rc != C.SQLITE_OK {
		return fmt.Errorf("prepare failed: %s", C.GoString(C.sqlite3_errmsg(s.db)))
	}
	defer C.sqlite3_finalize(stmt)
	for i, p := range params {
		cp := C.CString(p)
		C.sqlite3_bind_text(stmt, C.int(i+1), cp, -1, sqliteTransient)
		C.free(unsafe.Pointer(cp))
	}
	if rc := C.sqlite3_step(stmt); rc != C.SQLITE_DONE {
		return fmt.Errorf("insert failed: %s", C.GoString(C.sqlite3_errmsg(s.db)))
	}
	return nil
}
