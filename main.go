package main

import(
	"encoding/json"
	"log"
	"fmt"
	"net/http"
	"io/ioutil"
	"time"
	
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)


// Server struct (Model)
type Server struct {
	Address  string `json:"address"`
	SSLGrade string `json:"ssl_grade"`
	Country  string `json:"country"`
	Owner    string `json:"owner"`
}

// DomainInfo struct
type DomainInfo struct {
	Servers          []Server `json:"servers"`
	ServersChanged   bool         `json:"servers_changed"`
	SSLGrade         string       `json:"ssl_grade"`
	PreviousSSLGrade string       `json:"previous_ssl_grade"`
	Logo             string       `json:"logo"`
	Title            string       `json:"title"`
	IsDown           bool         `json:"is_down"`
	
}

// Get information from a specific domain
func queryServers(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Content-Type", "application/json")
	
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
		time.Sleep( 5 * time.Second)

	}

	//Show errors
	if status == "ERROR" {
		status, test = query["statusMessage"].(string)
		if !test {
			status = "Unknown error"
		}
		ctx.Error(status, fasthttp.StatusInternalServerError)
		resp:= &DomainInfo{}
		resp.IsDown = true
		serialized, err := json.Marshal(response)
		if err != nil {
			ctx.Error("Unable to serialize data", fasthttp.StatusInternalServerError)
		}
		fmt.Fprint(ctx, string(serialized))
	}else{
		//enviar a base de datos


		//
		actual := &DomainInfo{}
		servers := make([]Server, 0)
		endpointSlice := query["endpoints"].([]interface{})

		for _, endpoint := range endpointSlice {
			server := &Server{}
			server.Address = endpoint.(map[string]interface{})["ipAddress"].(string)
			fmt.Println("IpAddress: "+ server.Address )
			grade, test := endpoint.(map[string]interface{})["grade"].(string)
			if test {
				server.SSLGrade = grade
				fmt.Println("Grade: "+ server.SSLGrade )
			}
			//whois

			servers = append(servers, *server)
		}
		actual.Servers = servers
		//logo

		//Calcultae worstGrade
		grade := calculateWorstGrade(actual.Servers)

		if grade != "Z" {
			actual.SSLGrade = grade
		} else {
			actual.SSLGrade = "Undetermined"
		}
		


	}
	
}
/*
The SSL grades of the servers goes from A to F where A is the biggest grade.
The SSL grade of a domain is the minor SSL grade of the servers
*/
func calculateWorstGrade(servers []Server) string {
	grade := "Z"

	for i := range servers {

		if servers[i].SSLGrade != "" && servers[i].SSLGrade < grade {
			grade = servers[i].SSLGrade
		}
	}

	return grade
}

func queryHistory(ctx *fasthttp.RequestCtx) {
	ctx.Request.Header.Set("Content-Type", "application/json")
	
}

// Main function
func main() {
	// Init router
	r := fasthttprouter.New()
	
	// Route handles & endpoints
	r.GET("/serversInformation/:domain", queryServers)
	r.GET("/history", queryHistory)
	

	// Start server
	log.Fatal(fasthttp.ListenAndServe(":8080", r.Handler))
}

