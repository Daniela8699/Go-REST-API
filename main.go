package main

import(
	"log"


	"github.com/Daniela8699/Go-REST-API/controller"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)



// Get information from a specific domain
func queryServers(ctx *fasthttp.RequestCtx) {
	controller.getQueryServers(ctx)
}


// Get information about user History
func queryHistory(ctx *fasthttp.RequestCtx) {
	controller.getQueryHistory(ctx)
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

