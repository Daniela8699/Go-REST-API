package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	connection "github.com/Daniela8699/Go-REST-API/db"
	extra "github.com/Daniela8699/Go-REST-API/extrainfo"
	"github.com/Daniela8699/Go-REST-API/structs"
	"github.com/valyala/fasthttp"
)

//GetQueryServers get information from a specific domain
func GetQueryServers(ctx *fasthttp.RequestCtx) {

	db := connection.ConnectDB()
	var query map[string]interface{}
	domain := ctx.UserValue("domain").(string)
	response, err := http.Get("https://api.ssllabs.com/api/v3/analyze?host=" + domain + "&fromCache=on&maxAge=1")
	if err != nil {
		ctx.Error("Can not read data from api.ssllabs.com", fasthttp.StatusInternalServerError)
	}
	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(data, &query)

	status, test := query["status"].(string)

	//If something it's wrong 'status' always will be 'Error'
	if !test {
		status = "ERROR"
	}
	fmt.Println("Domain: ", domain, " Status:", status)
	for status != "READY" {
		if status == "ERROR" {
			fmt.Println("error..............")
			break
		}
		response, err := http.Get("https://api.ssllabs.com/api/v3/analyze?host=" + domain)

		if err != nil {
			fmt.Printf("The HTTP request failed with error %s\n", err)
			break
		}
		query = make(map[string]interface{})
		defer response.Body.Close()
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			ctx.Error("Can not read data  from api.ssllabs.com", fasthttp.StatusInternalServerError)
		}
		json.Unmarshal(data, &query)

		status = query["status"].(string)
		fmt.Println("Domain:", domain, " Status:", status)
		if status == "ERROR" {
			break
		}
		time.Sleep(5 * time.Second)

	}

	//Show errors
	if status == "ERROR" {
		status, test = query["statusMessage"].(string)
		if !test {
			status = "Unknown error"
		}
		ctx.Error(status, fasthttp.StatusInternalServerError)
		resp := &structs.DomainInfo{}
		resp.IsDown = true
		serialized, err := json.Marshal(response)
		if err != nil {
			ctx.Error("Unable to serialize data", fasthttp.StatusInternalServerError)
		}
		fmt.Fprint(ctx, string(serialized))
	} else {
		actual := GetDomainInfo(query, domain)
		actual.ServersChanged = false
		//send to db
		serversDB := connection.GetDomainInfo(domain, db)
		if serversDB.SSLGrade == "" {
			present := time.Now()
			lastUpdated := present.Format(time.RFC3339)

			actual.LastUpdated = lastUpdated
			connection.CreateDomainInfo(domain, actual, db)
			serversDB = connection.GetDomainInfo(domain, db)
		} else {
			//When domain exist in db
			t := time.Now()
			present, _ := time.Parse(time.RFC3339, t.Format(time.RFC3339))
			past, _ := time.Parse(time.RFC3339, serversDB.LastUpdated)
			duration := present.Sub(past)
			fmt.Println("Duration: ", duration)
			fmt.Println("Duration in Hours: ", int(duration.Hours()))

			if duration.Hours() >= 1 {

				if actual.SSLGrade != serversDB.PreviousSSLGrade {

					present := time.Now()
					lastUpdated := present.Format(time.RFC3339)
					actual.LastUpdated = lastUpdated
					actual.ServersChanged = true
					actual.PreviousSSLGrade = serversDB.SSLGrade

					fmt.Println("El servidor cambio")
					fmt.Println("SSLGRADE Antes: " + serversDB.PreviousSSLGrade + " Ahora:" + actual.SSLGrade)
				} else {
					actual.ServersChanged = false
					actual.PreviousSSLGrade = serversDB.SSLGrade
					actual.LastUpdated = serversDB.LastUpdated
				}
				connection.UpdateDomainInfo(domain, actual, db)
			}
		}
		serversDB = connection.GetDomainInfo(domain, db)
		message, err := json.Marshal(actual)
		var message2 string
		if err != nil {
			json.Unmarshal([]byte(message), &message2)
			fmt.Fprintf(ctx, message2)
		}
		if err := json.NewEncoder(ctx).Encode(actual); err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		}

	}
	//
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Content-Type", "application/json")

}

//GetDomainInfo get the domain information given by sslapi
func GetDomainInfo(query map[string]interface{}, domain string) structs.DomainInfo {

	actual := structs.DomainInfo{}
	endpointSlice := query["endpoints"].([]interface{})
	servers := make([]structs.Server, 0)
	//logo
	logo, title:= extra.GetInfoWebsite(domain)
	actual.Logo=logo
	actual.Title = title

	for _, endpoint := range endpointSlice {
		server := &structs.Server{}
		server.Address = endpoint.(map[string]interface{})["ipAddress"].(string)
		fmt.Println("IpAddress: " + server.Address)
		grade, test := endpoint.(map[string]interface{})["grade"].(string)
		if test {
			server.SSLGrade = grade
			fmt.Println("Grade: " + server.SSLGrade)
		}
		//whois
		owner, country := extra.GetWhoIsData(server.Address)
		
			server.Country = country
			server.Owner = owner
		
	

		servers = append(servers, *server)
	}
	actual.Servers = servers

	//Calcultae worstGrade
	grade := calculateWorstGrade(actual.Servers)

	if grade != "Z" {
		actual.SSLGrade = grade
	} else {
		actual.SSLGrade = "Unknown"
	}

	return actual
}

/*
The SSL grades of the servers goes from A to F where A is the biggest grade.
The SSL grade of a domain is the minor SSL grade of the servers
*/
func calculateWorstGrade(servers []structs.Server) string {
	grade := "Z"

	for i := range servers {

		if servers[i].SSLGrade != "" && servers[i].SSLGrade < grade {
			grade = servers[i].SSLGrade
		}
	}

	return grade
}

//GetQueryHistory get the history of the domains consulted
func GetQueryHistory(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Request.Header.Set("Content-Type", "application/json")

	db := connection.ConnectDB()

	var domains []structs.ServersHistoryElement

	domains = connection.GetHistoryServer(db)
	history := structs.ServersHistory{
		Items: domains,
	}
	jsonMsn, err := json.Marshal(history)
	var msn string
	if err != nil {
		json.Unmarshal([]byte(jsonMsn), &msn)
		fmt.Fprintf(ctx, msn)
	}
	if err := json.NewEncoder(ctx).Encode(history); err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	}

}
