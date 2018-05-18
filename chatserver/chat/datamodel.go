package chat

import (
	"database/sql"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

func getHallIDFromAgentCode(agentCode string) (uint32, error) {
	conn, err := redismgr.Get("game")
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	ret, err := redis.Uint64(conn.Do("hget", "agentWhitelist", agentCode))
	return uint32(ret), err
}

func getHallTableNames(id uint32) ([]string, error) {
	db, err := mysqlmgr.GetDB("master")
	if err != nil {
		return nil, err
	}
	rows, err := db.Query("SELECT DISTINCT table_name as tableName FROM user_route WHERE hall_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ret []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		ret = append(ret, tableName)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ret) == 0 {
		return nil, fmt.Errorf("getHallTableNames return emtpy id:%d", id)
	}
	return ret, nil
}

func getUserInfo(account string, agentCode string) (userID uint32, hallID uint32, agentID uint32, hallName string, agentName string, passwd string, state int, rank int, err error) {
	hallID, err = getHallIDFromAgentCode(agentCode)
	if err != nil {
		return
	}
	var tables []string
	tables, err = getHallTableNames(hallID)
	if err != nil {
		return
	}
	var db *sql.DB
	db, err = mysqlmgr.GetDB("master")
	if err != nil {
		return
	}
	for _, v := range tables {
		row := db.QueryRow(fmt.Sprintf("select uid,hall_id,agent_id,hall_name,agent_name,password,account_state,user_rank from %s where user_name = ? limit 1", v), account)
		err = row.Scan(&userID, &hallID, &agentID, &hallName, &agentName, &passwd, &state, &rank)
		if err != sql.ErrNoRows {
			return
		}
	}
	err = fmt.Errorf("account(%s) agentCode(%s) not found", account, agentCode)
	return
}

func getAllRoomIDs() ([]uint16, error) {
	db, err := mysqlmgr.GetDB("master")
	if err != nil {
		return nil, err
	}
	rows, err := db.Query("SELECT DISTINCT server_id FROM game_server_info WHERE enable = 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ret []uint16
	for rows.Next() {
		var roomID uint16
		if err := rows.Scan(&roomID); err != nil {
			return nil, err
		}
		ret = append(ret, roomID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ret) == 0 {
		return nil, fmt.Errorf("getAllRoomIDs return emtpy")
	}
	return ret, nil
}

func getLimits() (map[string]uint8, error) {
	db, err := mysqlmgr.GetDB("master")
	if err != nil {
		return nil, err
	}
	rows, err := db.Query("SELECT DISTINCT uid,hall_id FROM game_chat_limit_user WHERE limit_status = 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ret := make(map[string]uint8)
	for rows.Next() {
		var uid, hallid uint32
		if err := rows.Scan(&uid, &hallid); err != nil {
			return nil, err
		}
		ret[fmt.Sprintf("%d_%d", hallid, uid)] = 0
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}

func addLimitUser(userID uint32, userName string, hallID uint32, agentID uint32, hallName string, agentName string) error {
	db, err := mysqlmgr.GetDB("master")
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO game_chat_limit_user (uid,user_name,hall_id,agent_id,hall_name,agent_name,limit_status) VALUE(?,?,?,?,?,?,1) ON DUPLICATE KEY UPDATE limit_status = 1", userID, userName, hallID, agentID, hallName, agentName)
	return err
}

func getKeys() ([]string, error) {
	db, err := mysqlmgr.GetDB("master")
	if err != nil {
		return nil, err
	}
	rows, err := db.Query("SELECT keyword FROM game_chat_limit_keyword")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ret := make([]string, 0, 1000)
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			return nil, err
		}
		ret = append(ret, text)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}

func addMonitor(userID uint32, userName string, hallID uint32, agentID uint32, hallName string, agentName string, msg string) error {
	db, err := mysqlmgr.GetDB("master")
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO game_chat_monitor (uid,user_name,hall_id,agent_id,hall_name,agent_name,content,send_time) VALUE(?,?,?,?,?,?,?,NOW())", userID, userName, hallID, agentID, hallName, agentName, msg)
	return err
}
