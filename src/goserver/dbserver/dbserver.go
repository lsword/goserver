package dbserver

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"time"

	"goserver/config"
)

type DBItem struct {
	DriverName     string
	DataSourceName string
	MaxIdleConns   int
	MaxOpenConns   int
	Connected      int //1:connected, 0:notconnected
	DB             *sql.DB
}

type Database struct {
	DBItems map[string]*DBItem
}

var database *Database

func Run(c *config.Config) {
	if c.DBServer.Switch != "on" {
		return
	}

	if database == nil {
		database = &Database{
			DBItems: make(map[string]*DBItem),
		}
	}

	for i := 0; i < len(c.DBServer.DBItems); i++ {
		database.AddItem(
			c.DBServer.DBItems[i].DBName,
			c.DBServer.DBItems[i].DriverName,
			c.DBServer.DBItems[i].DataSourceName,
			c.DBServer.DBItems[i].MaxIdleConns,
			c.DBServer.DBItems[i].MaxOpenConns)
	}
	database.Connect()

	if c.DBServer.ConnCheckInterval > 0 {
		go func() {
			timer := time.NewTimer(time.Second * 5)
			for {
				select {
				case <-timer.C:
					database.Connect()
					timer.Reset(time.Second * 5)
				}
			}
		}()
	}
}

func GetDatabase() *Database {
	if database == nil {
		database = &Database{
			DBItems: make(map[string]*DBItem),
		}
	}
	return database
}

func Status() string {
	status := "DBName\tDriver\tMaxIdleConns\tMaxOpenConns\tConnected\tOpenConnections\n"
	for dbname, dbinfo := range database.DBItems {
		openConnections := 0
		db := database.DBItems[dbname].DB
		if db != nil {
			dbstats := db.Stats()
			openConnections = dbstats.OpenConnections
		}
		status = status + fmt.Sprintf("%s\t%s\t%d\t%d\t%d\t%d\n", dbname, dbinfo.DriverName, dbinfo.MaxIdleConns, dbinfo.MaxOpenConns, dbinfo.Connected, openConnections)
	}
	return status
}

func (database *Database) AddItem(itemName string, driverName string, dataSourceName string, maxIdleConns int, maxOpenConns int) {
	if database.DBItems[itemName] != nil {
		if database.DBItems[itemName].Connected == 1 {
			return
		}
	}
	var item DBItem
	item.DriverName = driverName
	item.DataSourceName = dataSourceName
	item.MaxIdleConns = maxIdleConns
	item.MaxOpenConns = maxOpenConns
	item.Connected = 0
	item.DB = nil
	database.DBItems[itemName] = &item
}

func (database *Database) DelItem(itemName string) {
	delete(database.DBItems, itemName)
}

func (database *Database) Connect() {
	for k, v := range database.DBItems {
		if v.Connected == 0 || v.DB == nil {
			db, err := sql.Open(v.DriverName, v.DataSourceName)
			if err != nil {
				continue
			}
			v.DB = db
			v.DB.SetMaxOpenConns(v.MaxOpenConns)
			v.DB.SetMaxIdleConns(v.MaxIdleConns)
			err = v.DB.Ping()
			if err != nil {
				//record log here
				println("connect db error:", err.Error())
				v.DB = nil
			} else {
				println("connect db ok:", k)
				v.Connected = 1
			}
		}
	}
}

func (database *Database) GetDB(itemname string) *sql.DB {
	if database.DBItems[itemname] == nil {
		return nil
	}
	return database.DBItems[itemname].DB
}

/*
You should remember to close sql.Rows
*/
func (database *Database) Query(dbname, sqlstr string) (*sql.Rows, error) {
	if database.DBItems[dbname] == nil {
		return nil, fmt.Errorf("db(%s) not found", dbname)
	}

	db := database.DBItems[dbname].DB

	if db == nil {
		database.DBItems[dbname].Connected = 0
		return nil, fmt.Errorf("db(%s) not connected", dbname)
	}

	rows, err := db.Query(sqlstr)
	return rows, err
}

/*
This function is for small results
*/
func (database *Database) QueryData(dbname, sqlstr string) (int, *[]map[string]interface{}, error) {
	if database.DBItems[dbname] == nil {
		return -1, nil, fmt.Errorf("db(%s) not found", dbname)
	}

	db := database.DBItems[dbname].DB

	if db == nil {
		database.DBItems[dbname].Connected = 0
		return -1, nil, fmt.Errorf("db(%s) not connected", dbname)
	}

	rows, err := db.Query(sqlstr)
	if err != nil {
		return -1, nil, err
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	records := make([]map[string]interface{}, 0)
	for rows.Next() {
		record := make(map[string]interface{})
		err := rows.Scan(scanArgs...)
		if err != nil {
			continue
		}
		for i, col := range values {
			if col != nil {
				record[columns[i]] = col
			} else {
				record[columns[i]] = nil
			}
		}
		records = append(records, record)
	}

	/*
		for row := range records {
			for k, v := range records[row] {
				println(k, ":", v)
			}
		}
	*/
	return len(records), &records, nil
}

func (database *Database) Exec(dbname, sqlstr string, args ...interface{}) (int64, int64, error) {
	if database.DBItems[dbname] == nil {
		return -1, -1, fmt.Errorf("db(%s) not found", dbname)
	}

	db := database.DBItems[dbname].DB

	if db == nil {
		database.DBItems[dbname].Connected = 0
		fmt.Println(dbname + " not connected")
		return -1, -1, fmt.Errorf("db(%s) not connected", dbname)
	}

	stmt, err := db.Prepare(sqlstr)
	if err != nil {
		return -1, -1, fmt.Errorf("db(%s) prepare error:%s", dbname, err.Error())
	}
	defer stmt.Close()

	res, err := stmt.Exec(args...)
	if err != nil {
		return -1, -1, fmt.Errorf("db(%s) exec error:%s", dbname, err.Error())
	}

	affectCnt, _ := res.RowsAffected()
	lastId, _ := res.LastInsertId()

	return lastId, affectCnt, nil
}
