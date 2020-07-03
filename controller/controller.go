package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	connection "github.com/Daniela8699/Go-REST-API/db"
	"github.com/Daniela8699/Go-REST-API/structs"
	"github.com/valyala/fasthttp"
)

//GetQueryServers get information from a specific domain
func GetQueryServers(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Content-Type", "application/json")

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
		actual := GetDomainInfo(query)

		//enviar a base de datos
		infoServerDB := connection.GetInfoServer(domain, db)
		if infoServerDB.SSLGrade == "" {
			present := time.Now()
			lastUpdated := present.Format(time.RFC3339)
			actual.LastUpdated = lastUpdated
			connection.CreateInfoServer(domain, actual, db)
			infoServerDB = connection.GetInfoServer(domain, db)
		} else {
			// When infoServerDB exist then update it
			// Validate the previous state
			t := time.Now()
			present, _ := time.Parse(time.RFC3339, t.Format(time.RFC3339))
			past, _ := time.Parse(time.RFC3339, infoServerDB.LastUpdated)
			duration := present.Sub(past)
			fmt.Println("Present: ", present)
			fmt.Println("Past: ", past)
			fmt.Println("Duration: ", duration)
			fmt.Println("Duration.Hours: ", int(duration.Hours()))
			// fmt.Println("Duration.Minutes: ", int(duration.Minutes()))

			// Only or tests
			// if duration.Minutes() >= 0 {
			// Only update past one hour and grade ssl changed
			if duration.Hours() >= 0 {
				// Past one hour, validate grade ssl if changed
				// Only or tests
				// infoServer.SslGrade = "C"
				if actual.SSLGrade != infoServerDB.PreviousSSLGrade {
					actual.ServersChanged = true
					actual.PreviousSSLGrade = infoServerDB.SSLGrade
					present := time.Now()
					lastUpdated := present.Format(time.RFC3339)
					actual.LastUpdated = lastUpdated
				} else {
					actual.ServersChanged = false
					actual.PreviousSSLGrade = infoServerDB.SSLGrade
					actual.LastUpdated = infoServerDB.LastUpdated
				}
				connection.UpdateInfoServer(domain, actual, db)
			}
		}
		infoServerDB = connection.GetInfoServer(domain, db)
		fmt.Println("InfoServer: ", actual)

	}
	//

}

//GetDomainInfo get the domain information given by sslapi
func GetDomainInfo(query map[string]interface{}) structs.DomainInfo {

	actual := structs.DomainInfo{}
	actual.ServersChanged = false;
	servers := make([]structs.Server, 0)
	endpointSlice := query["endpoints"].([]interface{})

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

		servers = append(servers, *server)
	}
	actual.Servers = servers
	//logo
	logo, title, err := GetInfoWebsite(domain)

	if err == nil {
		actual.Title = title
		actual.Logo = logo
	}
	//Calcultae worstGrade
	grade := calculateWorstGrade(actual.Servers)

	if grade != "Z" {
		actual.SSLGrade = grade
	} else {
		actual.SSLGrade = "Undetermined"
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

func GetQueryHistory(ctx *fasthttp.RequestCtx) {
	ctx.Request.Header.Set("Content-Type", "application/json")

}
