package main

import (
	"log"
	

	"github.com/Daniela8699/Go-REST-API/controller"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)
var db *sql.DB

// Get information from a specific domain
func queryServers(ctx *fasthttp.RequestCtx) {
	controller.GetQueryServers(ctx, db)
}

// Get information about user History
func queryHistory(ctx *fasthttp.RequestCtx) {
	controller.GetQueryHistory(ctx, db)
}


// Main function
func main() {
	// Init router & DB
	r := fasthttprouter.New()
	startDB()

	// Route handles & endpoints
	r.GET("/serversInformation/:domain", queryServers)
	r.GET("/history", queryHistory)

	// Start server
	log.Fatal(fasthttp.ListenAndServe(":8080", r.Handler))
}
