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
	db, err := sql.Open("postgres", "postgresql://root@localhost:26257/domains?sslmode=disable")
	if err != nil {

		fmt.Println("ERROR connecting to the database: ", err)
	}
	if _, err := db.Exec(
		"CREATE TABLE IF NOT EXISTS infoservers (host STRING PRIMARY KEY, report JSONB, last_updated string)"); err != nil {
		log.Fatal(err)
		fmt.Println("ERROR create table infoservers: ", err)
	}
	return db
}
// CreateDomainInfo create domainInfo into the infoservers table
func CreateDomainInfo(domain string, infoServer structs.DomainInfo, db *sql.DB) bool {
	domainMsn, err := json.Marshal(infoServer)
	if err != nil {
		fmt.Println("ERROR parse data infoServer to string: ", err)
		return false
	}

	// Insert one row into the "infoservers" table.
	if _, err := db.Exec(
		"INSERT INTO infoservers (host, report, last_updated) VALUES ('" + domain + "', '" + string(domainMsn) + "', '" + infoServer.LastUpdated + "')"); err != nil {
		fmt.Println("ERROR Insert one row infoservers table: ", err)
		return false
	}
	return true
}

// GetDomainInfo get information about a server
func GetDomainInfo(domain string, db *sql.DB) structs.DomainInfo {
	rows, err := db.Query("SELECT * FROM infoservers WHERE host = '" + domain + "'")
	if err != nil {

		fmt.Println("ERROR select infoservers: ", err)
	}
	defer rows.Close()

	var infoDomain structs.DomainInfo

	var domainName, report, lastUpdated string
	for rows.Next() {
		if err := rows.Scan(&domainName, &report, &lastUpdated); err != nil {

			fmt.Println("ERROR get data for result set: ", err)
		}

		bytes := []byte(report)
		err := json.Unmarshal(bytes, &infoDomain)
		if err != nil {

			fmt.Println("ERROR create data infoServer: ", err)
		}

		infoDomain.LastUpdated = lastUpdated
	}

	return infoDomain
}


// UpdateDomainInfo update domainInfo into the infoservers table
func UpdateDomainInfo(domain string, domainInfo structs.DomainInfo, db *sql.DB) bool {
	domainMsn, err := json.Marshal(domainInfo)
	if err != nil {
		// panic(err)
		fmt.Println("ERROR parse data infoServer to string: ", err)
		return false
	}

	if _, err := db.Exec(
		"UPDATE infoservers SET (report, last_updated) = ('" + string(domainMsn) + "', '" + domainInfo.LastUpdated + "') WHERE host = '" + domain + "'"); err != nil {
		// log.Fatal(err)
		fmt.Println("ERROR Update one row infoservers table: ", err)
		return false
	}
	return true
}

// GetHistoryServer get information about a server
func GetHistoryServer(db *sql.DB) []structs.ServersHistoryElement {
	// Print out the infoserver.
	rows, err := db.Query("SELECT host FROM infoservers ORDER BY last_updated DESC")
	domains := make([]structs.ServersHistoryElement, 0)
	if err != nil {
		fmt.Println("error...", err)
	}
	defer rows.Close()
	for rows.Next() {
		domain := structs.ServersHistoryElement{}

		err := rows.Scan(&domain.Host)
		if err != nil {
			log.Fatal(err)

		}

		domains = append(domains, domain)

	}

	return domains
}
