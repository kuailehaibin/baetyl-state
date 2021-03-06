package database

import (
	"database/sql"

	"github.com/baetyl/baetyl-go/kv"
)

var placeholderValue = "(?)"
var placeholderKeyValue = "(?,?)"
var schema = map[string][]string{
	"sqlite3": []string{
		`CREATE TABLE IF NOT EXISTS kv (
			key TEXT PRIMARY KEY,
			value BLOB,
			ts TIMESTAMP DEFAULT CURRENT_TIMESTAMP) WITHOUT ROWID`,
	},
}

func init() {
	Factories["sqlite3"] = _new
}

// sqldb the backend SQL DB to persist values
type sqldb struct {
	*sql.DB
	conf Conf
}

// New creates a new sql database
func _new(conf Conf) (DB, error) {
	db, err := sql.Open(conf.Driver, conf.Source)
	if err != nil {
		return nil, err
	}
	for _, v := range schema[conf.Driver] {
		if _, err = db.Exec(v); err != nil {
			db.Close()
			return nil, err
		}
	}
	return &sqldb{DB: db, conf: conf}, nil
}

// Conf returns the configuration
func (d *sqldb) Conf() Conf {
	return d.conf
}

// Set put key and value into SQL DB
func (d *sqldb) Set(kv *kv.KV) error {
	stmt, err := d.Prepare("insert into kv(key,value) values (?,?) on conflict(key) do update set value=excluded.value")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(kv.Key, kv.Value)
	return err
}

// Get gets value by key from SQL DB
func (d *sqldb) Get(key string) (*kv.KV, error) {
	rows, err := d.Query("select value from kv where key=?", key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	kv := &kv.KV{Key: key}
	if rows.Next() {
		err = rows.Scan(&kv.Value)
		if err != nil {
			return nil, err
		}
		return kv, nil
	}
	return kv, nil
}

// Del deletes key and value from SQL DB
func (d *sqldb) Del(key string) error {
	_, err := d.Exec("delete from kv where key=?", key)
	return err
}

// List list kvs with the prefix
func (d *sqldb) List(prefix string) (*kv.KVs, error) {
	rows, err := d.Query("select key, value from kv where key like ?", prefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	kvs := &kv.KVs{}
	for rows.Next() {
		kv := new(kv.KV)
		err = rows.Scan(&kv.Key, &kv.Value)
		if err != nil {
			return nil, err
		}
		kvs.Kvs = append(kvs.Kvs, kv)
	}
	return kvs, nil
}
