package db

import (
	"database/sql"
	"os"

	_ "modernc.org/sqlite"
)

const createSchedulerTable = `
CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL,
    title VARCHAR NOT NULL,
    comment TEXT,
    repeat VARCHAR CHECK(length(repeat) <= 128)
);

CREATE INDEX IF NOT EXISTS idx_date ON scheduler (date);`

var DB *sql.DB

func InitDB(dbFile string) error {

	_, err := os.Stat(dbFile)
	var install bool
	if err != nil {
		install = true
	}

	DB, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}

	if install {
		_, err = DB.Exec(createSchedulerTable)
		if err != nil {
			return err
		}
	}

	return nil
}
