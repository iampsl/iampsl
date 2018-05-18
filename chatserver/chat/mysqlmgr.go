package chat

import (
	"chatserver/global"
	"database/sql"
	"fmt"
	"log"
)

//MysqlMgr 管理mysql db
type MysqlMgr struct {
	dbs map[string]*sql.DB
}

var mysqlmgr = MysqlMgr{dbs: make(map[string]*sql.DB)}

//GetDB 得到指定db
func (mgr *MysqlMgr) GetDB(name string) (*sql.DB, error) {
	db, b := mgr.dbs[name]
	if !b {
		return nil, fmt.Errorf("not found mysql db:%s", name)
	}
	return db, nil
}

func initMysqlMgr() {
	for _, v := range global.AppConfig.MysqlConfig.MysqlSplice {
		_, b := mysqlmgr.dbs[v.Name]
		if b {
			log.Fatalf("mysql %s have already\n", v.Name)
		}
		db, err := sql.Open("mysql", v.Dsn)
		if err != nil {
			log.Fatalln(err)
		}
		db.SetMaxOpenConns(int(v.MaxConnect))
		db.SetMaxIdleConns(int(v.MaxConnect))
		mysqlmgr.dbs[v.Name] = db
	}
}
