package db

import (
	"database/sql"
	"encoding/json"
	"fmt"

	 "log"
	// User Go pq driver
	_ "github.com/lib/pq"

	"github.com/Daniela8699/Go-REST-API/structs"
)

// ConnectDB create a connection db
func ConnectDB() *sql.DB {
	// Connect to the "truora" database.
	db, err := sql.Open("postgres", "postgresql://root@localhost:26257/domains?sslmode=disable")
	if err != nil {
		// log.Fatal(err)
		fmt.Println("ERROR connecting to the database: ", err)
	}
	// Create the "accounts" table.
	if _, err := db.Exec(
		"CREATE TABLE IF NOT EXISTS infoservers (host STRING PRIMARY KEY, report JSONB, last_updated string)"); err != nil {
		log.Fatal(err)
		fmt.Println("ERROR create table infoservers: ", err)
	}
	return db
}

// GetInfoServer get information about a server
func GetInfoServer(domain string, db *sql.DB) structs.DomainInfo {
	// Print out the infoserver.
	rows, err := db.Query("SELECT * FROM infoservers WHERE host = '" + domain + "'")
	if err != nil {

		fmt.Println("ERROR select infoservers: ", err)
	}
	defer rows.Close()

	var infoServer structs.DomainInfo

	var domainName, report, lastUpdated string
	for rows.Next() {
		if err := rows.Scan(&domainName, &report, &lastUpdated); err != nil {

			fmt.Println("ERROR get data for result set: ", err)
		}

		bytes := []byte(report)
		err := json.Unmarshal(bytes, &infoServer)
		if err != nil {

			fmt.Println("ERROR create data infoServer: ", err)
		}

		infoServer.LastUpdated = lastUpdated
	}

	return infoServer
}

// CreateInfoServer create infoServer into the infoservers table
func CreateInfoServer(domain string, infoServer structs.DomainInfo, db *sql.DB) bool {
	infoServerStr, err := json.Marshal(infoServer)
	if err != nil {
		// panic(err)
		fmt.Println("ERROR parse data infoServer to string: ", err)
		return false
	}
	// fmt.Println(string(infoServerStr))
	// Insert one row into the "infoservers" table.
	if _, err := db.Exec(
		"INSERT INTO infoservers (host, report, last_updated) VALUES ('" + domain + "', '" + string(infoServerStr) + "', '" + infoServer.LastUpdated + "')"); err != nil {
		// log.Fatal(err)
		fmt.Println("ERROR Insert one row infoservers table: ", err)
		return false
	}
	return true
}

// UpdateInfoServer update infoServer into the infoservers table
func UpdateInfoServer(domain string, infoServer structs.DomainInfo, db *sql.DB) bool {
	infoServerStr, err := json.Marshal(infoServer)
	if err != nil {
		// panic(err)
		fmt.Println("ERROR parse data infoServer to string: ", err)
		return false
	}
	// fmt.Println(string(infoServerStr))
	// Insert one row into the "infoservers" table.
	if _, err := db.Exec(
		"UPDATE infoservers SET (report, last_updated) = ('" + string(infoServerStr) + "', '" + infoServer.LastUpdated + "') WHERE host = '" + domain + "'"); err != nil {
		// log.Fatal(err)
		fmt.Println("ERROR Update one row infoservers table: ", err)
		return false
	}
	return true
}
