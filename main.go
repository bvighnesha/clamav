package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"os"
)

func main() {

	clamServer := os.Getenv("CLAM_SERVER")
	if clamServer == "" {
		clamServer = "0.0.0.0"
		log.Println("Missing environment host for the clam server. Please set CLAM_SERVER OR default", clamServer,"is used.")
	}
	clamPort := os.Getenv("CLAM_PORT")
	if clamPort == "" {
		clamPort = "3310"
	}
	clamUri := fmt.Sprintf("tcp://%s:%s", clamServer, clamPort)

	router := httprouter.New()
	controller := ClamAV{clamUri}

	router.GET("/", controller.Index)
	router.GET("/version", controller.Version)
	router.GET("/stats", controller.Stats)
	router.POST("/scan", controller.Scan)
	router.GET("/health", controller.Health)

	log.Println("Service is starting...")
	log.Fatal(http.ListenAndServe(":8000", router))
}
