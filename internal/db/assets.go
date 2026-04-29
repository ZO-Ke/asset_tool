package db

import (
	"asset_tool_go/internal/model"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// UpsertAssets 批量写入资产；已存在的合并 source 标签
func UpsertAssets(projectID int64, entries []model.AssetEntry) (model.ImportResult, error) {
	res := model.ImportResult{}
	if len(entries) == 0 {
		return res, nil
	}
	tx, err := conn.Begin()
	if err != nil {
		return res, err
	}
	defer tx.Rollback()

	for _, e := range entries {
		host := strings.TrimSpace(e.Host)
		if host == "" {
			continue
		}
		port := strings.TrimSpace(e.Port)

		var existingID int64
		var existingSrcRaw string
		err := tx.QueryRow(
			"SELECT id, sources FROM assets WHERE project_id=? AND host=? AND IFNULL(port,'')=IFNULL(?,'')",
			projectID, host, nullable(port),
		).Scan(&existingID, &existingSrcRaw)

		if err == sql.ErrNoRows {
			srcJSON := "[]"
			if e.Source != "" {
				if b, err := json.Marshal([]string{e.Source}); err == nil {
					srcJSON = string(b)
				}
			}
			_, err := tx.Exec(
				"INSERT INTO assets(project_id, type, host, port, sources) VALUES(?,?,?,?,?)",
				projectID, e.Type, host, nullable(port), srcJSON,
			)
			if err != nil {
				return res, fmt.Errorf("insert: %w", err)
			}
			if e.Type == "ip" {
				res.NewIP++
			} else {
				res.NewDomain++
			}
		} else if err == nil {
			var srcs []string
			_ = json.Unmarshal([]byte(existingSrcRaw), &srcs)
			if e.Source != "" && !contains(srcs, e.Source) {
				srcs = append(srcs, e.Source)
				if b, err := json.Marshal(srcs); err == nil {
					_, _ = tx.Exec("UPDATE assets SET sources=? WHERE id=?", string(b), existingID)
				}
			}
			res.Skipped++
		} else {
			return res, err
		}
	}

	return res, tx.Commit()
}

// AddPortAssets rustscan 等端口扫描结果：每个端口建独立记录
func AddPortAssets(projectID int64, host string, ports []int, source string) (int, error) {
	if len(ports) == 0 {
		return 0, nil
	}
	tx, err := conn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var baseType, baseSrc string
	err = tx.QueryRow(
		"SELECT type, sources FROM assets WHERE project_id=? AND host=? ORDER BY id LIMIT 1",
		projectID, host,
	).Scan(&baseType, &baseSrc)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	var srcs []string
	_ = json.Unmarshal([]byte(baseSrc), &srcs)
	if source != "" && !contains(srcs, source) {
		srcs = append(srcs, source)
	}
	srcJSON, _ := json.Marshal(srcs)

	added := 0
	for _, p := range ports {
		portStr := fmt.Sprintf("%d", p)
		_, err := tx.Exec(
			"INSERT OR IGNORE INTO assets(project_id, type, host, port, sources) VALUES(?,?,?,?,?)",
			projectID, baseType, host, portStr, string(srcJSON),
		)
		if err == nil {
			// 检查是否真的新增
			added++
		}
	}
	return added, tx.Commit()
}

// ListAssets 列出资产，支持按 type / status 过滤
func ListAssets(projectID int64, typeFilter, statusFilter string) ([]model.Asset, error) {
	q := `SELECT id, project_id, type, host, IFNULL(port,''), sources,
	             IFNULL(status,''), status_code, IFNULL(title,''), IFNULL(server,''),
	             IFNULL(tech,''), IFNULL(probed_at,''), created_at
	      FROM assets WHERE project_id=?`
	args := []any{projectID}
	if typeFilter != "" {
		q += " AND type=?"
		args = append(args, typeFilter)
	}
	if statusFilter != "" {
		q += " AND status=?"
		args = append(args, statusFilter)
	}
	q += " ORDER BY created_at DESC, id DESC"

	rows, err := conn.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.Asset
	for rows.Next() {
		var a model.Asset
		var srcRaw string
		var statusCode sql.NullInt64
		err := rows.Scan(
			&a.ID, &a.ProjectID, &a.Type, &a.Host, &a.Port, &srcRaw,
			&a.Status, &statusCode, &a.Title, &a.Server,
			&a.Tech, &a.ProbedAt, &a.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		if statusCode.Valid {
			c := int(statusCode.Int64)
			a.StatusCode = &c
		}
		_ = json.Unmarshal([]byte(srcRaw), &a.Sources)
		if a.Sources == nil {
			a.Sources = []string{}
		}
		result = append(result, a)
	}
	return result, nil
}

// DeleteAsset 删除单个资产
func DeleteAsset(id int64) error {
	_, err := conn.Exec("DELETE FROM assets WHERE id=?", id)
	return err
}

// DeleteAssets 批量删除
func DeleteAssets(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, id := range ids {
		if _, err := tx.Exec("DELETE FROM assets WHERE id=?", id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetAllHosts 取项目所有 host[:port] 用于探活/端口扫描输入
func GetAllHosts(projectID int64) ([]string, error) {
	rows, err := conn.Query(
		"SELECT host, IFNULL(port,'') FROM assets WHERE project_id=?",
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []string
	for rows.Next() {
		var h, p string
		if err := rows.Scan(&h, &p); err != nil {
			return nil, err
		}
		if p != "" {
			hosts = append(hosts, h+":"+p)
		} else {
			hosts = append(hosts, h)
		}
	}
	return hosts, nil
}

// UpdateProbeResult 写入 httpx 探活结果
func UpdateProbeResult(projectID int64, host string, status string, code *int, title, server, tech string) error {
	var codeVal any
	if code != nil {
		codeVal = *code
	}
	_, err := conn.Exec(
		`UPDATE assets SET status=?, status_code=?, title=?, server=?, tech=?, probed_at=datetime('now','localtime')
		 WHERE project_id=? AND host=?`,
		status, codeVal, title, server, tech, projectID, host,
	)
	return err
}

// ── helpers ──────────────────────────────────────────────────────────────

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func contains(arr []string, s string) bool {
	for _, x := range arr {
		if x == s {
			return true
		}
	}
	return false
}
