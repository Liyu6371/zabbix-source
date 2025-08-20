package sqlite

import (
	"database/sql"
	"fmt"
)

type HostRecord struct {
	HostID       int
	TemplateName string
}

type ItemRecord struct {
	ItemID     int
	HostID     int
	Key        string
	TemplateID sql.NullInt64
}

func QueryHostTable(mysql *sql.DB) ([]HostRecord, error) {
	queryStr := `select hostid, name as template_name  from hosts where status =3`
	rows, err := mysql.Query(queryStr)
	if err != nil {
		return nil, fmt.Errorf("query host table failed: %w", err)
	}
	defer rows.Close()

	var hosts []HostRecord
	for rows.Next() {
		var host HostRecord
		if err := rows.Scan(&host.HostID, &host.TemplateName); err != nil {
			return nil, fmt.Errorf("scan host row failed: %w", err)
		}
		hosts = append(hosts, host)
	}
	return hosts, nil
}

// SyncHostsToSqlite 同步数据到 sqlite hosts 表
// 同步数据前会全量删除 过去的旧内容
func SyncHostsToSqlite(sqlite *sql.DB, hosts []HostRecord) error {
	tx, err := sqlite.Begin()
	if err != nil {
		return err
	}
	tx.Exec("DELETE FROM hosts")
	stmt, err := tx.Prepare("INSERT INTO hosts(hostid, template_name) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, h := range hosts {
		if _, err := stmt.Exec(h.HostID, h.TemplateName); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// QueryItemsTable 查询 items 表
func QueryItemsTable(mysql *sql.DB) ([]ItemRecord, error) {
	queryStr := `select itemid, hostid, key_, templateid from items`
	rows, err := mysql.Query(queryStr)
	if err != nil {
		return nil, fmt.Errorf("query items table failed: %w", err)
	}
	defer rows.Close()

	var items []ItemRecord
	for rows.Next() {
		var item ItemRecord
		if err := rows.Scan(&item.ItemID, &item.HostID, &item.Key, &item.TemplateID); err != nil {
			return nil, fmt.Errorf("scan item row failed: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

// SyncItemsToSqlite 同步数据到 sqlite items 表
// 同步数据前会全量删除 过去的旧内容
func SyncItemsToSqlite(sqlite *sql.DB, items []ItemRecord) error {
	tx, err := sqlite.Begin()
	if err != nil {
		return err
	}
	tx.Exec("DELETE FROM items")
	stmt, err := tx.Prepare("INSERT INTO items(itemid, hostid, key_, templateid) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, i := range items {
		if _, err := stmt.Exec(i.ItemID, i.HostID, i.Key, i.TemplateID); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
