package db

import (
	"asset_tool_go/internal/model"
	"database/sql"
	"fmt"
	"strings"
)

// CreateProject 新建项目
func CreateProject(name string) (int64, error) {
	res, err := conn.Exec("INSERT INTO projects(name) VALUES(?)", strings.TrimSpace(name))
	if err != nil {
		return 0, fmt.Errorf("create project: %w", err)
	}
	return res.LastInsertId()
}

// ListProjects 列出项目，按 search 模糊匹配
func ListProjects(search string) ([]model.Project, error) {
	q := `
		SELECT p.id, p.name, p.created_at,
		       COUNT(a.id)                                       AS asset_count,
		       SUM(CASE WHEN a.type='ip'     THEN 1 ELSE 0 END)  AS ip_count,
		       SUM(CASE WHEN a.type='domain' THEN 1 ELSE 0 END)  AS domain_count,
		       SUM(CASE WHEN a.status='alive' THEN 1 ELSE 0 END) AS alive_count
		FROM projects p
		LEFT JOIN assets a ON a.project_id = p.id
	`
	args := []any{}
	if s := strings.TrimSpace(search); s != "" {
		q += " WHERE p.name LIKE ?"
		args = append(args, "%"+s+"%")
	}
	q += " GROUP BY p.id ORDER BY p.created_at DESC"

	rows, err := conn.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var result []model.Project
	for rows.Next() {
		var p model.Project
		var ip, domain, alive sql.NullInt64
		if err := rows.Scan(&p.ID, &p.Name, &p.CreatedAt, &p.AssetCount, &ip, &domain, &alive); err != nil {
			return nil, err
		}
		p.IPCount = int(ip.Int64)
		p.DomainCount = int(domain.Int64)
		p.AliveCount = int(alive.Int64)
		result = append(result, p)
	}
	return result, nil
}

// RenameProject 重命名项目
func RenameProject(id int64, newName string) error {
	_, err := conn.Exec("UPDATE projects SET name=? WHERE id=?", strings.TrimSpace(newName), id)
	return err
}

// DeleteProject 删除项目（级联删除关联资产）
func DeleteProject(id int64) error {
	_, err := conn.Exec("DELETE FROM projects WHERE id=?", id)
	return err
}
