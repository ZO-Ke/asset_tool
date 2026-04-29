package db

import "database/sql"

// GetSetting 读取设置项
func GetSetting(key string) string {
	var v string
	err := conn.QueryRow("SELECT value FROM settings WHERE key=?", key).Scan(&v)
	if err == sql.ErrNoRows {
		return ""
	}
	return v
}

// SetSetting 写入设置项（upsert）
func SetSetting(key, value string) error {
	_, err := conn.Exec(
		`INSERT INTO settings(key,value) VALUES(?,?)
		 ON CONFLICT(key) DO UPDATE SET value=excluded.value`,
		key, value,
	)
	return err
}
