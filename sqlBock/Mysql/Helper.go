package Mysql

import (
	"os"
	"encoding/json"
	"strconv"
	mysql "database/sql"
	_ "github.com/go-sql-driver/mysql"
	"strings"
)

func query(tablename string, params map[string]interface{}, fields []string, sql string, vaules []interface{}) map[string]interface{} {
	if vaules == nil {
		vaules = make([]interface{}, 0)
	}
	rs := execQeury("select "+strings.Join(fields, ",")+" from "+tablename, vaules)
	return rs
}

func Query(tablename string, params map[string]interface{}, fields []string) map[string]interface{} {
	return query(tablename, params, fields, "", nil)
}

func Insert(tablename string, params map[string]interface{}) map[string]interface{} {
	values := make([]interface{}, 0)
	sql := "INSERT INTO `" + tablename + "` (" //+strings.Join(allFields, ",")+") VALUES ("
	var ks []string
	var vs []string
	for k, v := range params {
		ks = append(ks, "`" + k + "`")
		vs = append(vs, "?")
		values = append(values, v)
	}
	sql += strings.Join(ks, ",") + ") VALUES (" + strings.Join(vs, ",") + ")"
	return execute(sql, values)
}

func Update(tablename string, params map[string]interface{}, id string) map[string]interface{} {
	sql := "Update " + tablename + " set "
	values := make([]interface{}, 0)
	return execute(sql, values)
}

func Delete(tablename string, params map[string]interface{}, id string) map[string]interface{} {
	sql := "Delete from " + tablename + " where"
	values := make([]interface{}, 0)
	return execute(sql, values)
}

func execute(sql string, values []interface{}) (rs map[string]interface{}) {
	rs = make(map[string]interface{})
	defer func() {
		if r := recover(); r != nil {
			rs["code"] = 500
			rs["err"] = "Exception, " + r.(error).Error()
		}
	}()

	var configs interface{}
	fr, err := os.Open("./configs.json")
	decoder := json.NewDecoder(fr)
	err = decoder.Decode(&configs)
	if err != nil {
		rs["code"] = 204
		rs["err"] = "Open database config error, " + err.Error()
		return rs
	}

	confs := configs.(map[string]interface{})
	dialect := confs["database_dialect"].(string)

	dbConf := confs["db_"+dialect+"_config"].(map[string]interface{})
	dbHost := dbConf["db_host"].(string)
	dbPort := strconv.FormatFloat(dbConf["db_port"].(float64), 'f', -1, 32)
	dbUser := dbConf["db_user"].(string)
	dbPass := dbConf["db_pass"].(string)
	dbName := dbConf["db_name"].(string)
	dbCharset := dbConf["db_charset"].(string)

	dao, err := mysql.Open(dialect, dbUser+":"+dbPass+"@tcp("+dbHost+":"+dbPort+")/"+dbName+"?charset="+dbCharset)
	defer dao.Close()
	if err != nil {
		rs["code"] = 204
		rs["err"] = err.Error()
		return rs
	}

	stmt, _ := dao.Prepare(sql)
	ers, err := stmt.Exec(values...)
	if err != nil {
		rs["code"] = 204
		rs["err"] = err.Error()
	} else {
		id, _ := ers.LastInsertId()
		affect, _ := ers.RowsAffected()
		rs["code"] = 200
		rs["info"] = "Insert operation success."
		rs["LastInsertId"] = id
		rs["RowsAffected"] = affect

	}
	return rs
}

func execQeury(sql string, values []interface{}) (rs map[string]interface{}) {
	rs = make(map[string]interface{})
	defer func() {
		if r := recover(); r != nil {
			rs["code"] = 500
			rs["err"] = "Exception, " + r.(error).Error()
		}
	}()

	var configs interface{}
	fr, err := os.Open("./configs.json")
	decoder := json.NewDecoder(fr)
	err = decoder.Decode(&configs)
	if err != nil {
		rs["code"] = 204
		rs["err"] = "Open database config error, " + err.Error()
		return rs
	}

	confs := configs.(map[string]interface{})
	dialect := confs["database_dialect"].(string)

	dbConf := confs["db_"+dialect+"_config"].(map[string]interface{})
	dbHost := dbConf["db_host"].(string)
	dbPort := strconv.FormatFloat(dbConf["db_port"].(float64), 'f', -1, 32)
	dbUser := dbConf["db_user"].(string)
	dbPass := dbConf["db_pass"].(string)
	dbName := dbConf["db_name"].(string)
	dbCharset := dbConf["db_charset"].(string)

	dao, err := mysql.Open(dialect, dbUser+":"+dbPass+"@tcp("+dbHost+":"+dbPort+")/"+dbName+"?charset="+dbCharset)
	defer dao.Close()
	if err != nil {
		rs["code"] = 204
		rs["err"] = err.Error()
		return rs
	}
	stmt, err := dao.Prepare(sql)
	if err != nil {
		rs["code"] = 204
		rs["err"] = err.Error()
		return rs
	}
	rows, err := stmt.Query(values...)
	if err != nil {
		rs["code"] = 204
		rs["err"] = err.Error()
		return rs
	}

	columns, err := rows.Columns()
	if err != nil {
		rs["code"] = 204
		rs["err"] = err.Error()
		return rs
	}
	vs := make([]mysql.RawBytes, len(columns))
	scans := make([]interface{}, len(columns))

	for i := range vs {
		scans[i] = &vs[i]
	}

	var result []map[string]string
	for rows.Next() {
		_ = rows.Scan(scans...)
		each := make(map[string]string)

		for i, col := range vs {
			each[columns[i]] = string(col)
		}

		result = append(result, each)
	}
	rs["code"] = 200
	//data, _ := json.Marshal(result)
	rs["rows"] = result
	return rs
}

func escape(source string) string {
	var j int
	if len(source) == 0 {
		return ""
	}
	tempStr := source[:]
	desc := make([]byte, len(tempStr)*2)
	for i := 0; i < len(tempStr); i++ {
		flag := false
		var escape byte
		switch tempStr[i] {
		case '\r':
			flag = true
			escape = '\r'
		case '\n':
			flag = true
			escape = '\n'
		case '\\':
			flag = true
			escape = '\\'
		case '\'':
			flag = true
			escape = '\''
		case '"':
			flag = true
			escape = '"'
		case '\032':
			flag = true
			escape = 'Z'
		default:
		}
		if flag {
			desc[j] = '\\'
			desc[j+1] = escape
			j = j + 2
		} else {
			desc[j] = tempStr[i]
			j = j + 1
		}
	}
	return string(desc[0:j])
}
