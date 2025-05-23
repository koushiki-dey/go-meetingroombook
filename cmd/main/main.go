package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/koushikidey/go-meetingroombook/pkg/routes"
)

func main() {
	router := mux.NewRouter()
	routes.RegisterMeetingRoomRoutes(router)
	http.Handle("/", router)
	log.Fatal(http.ListenAndServe("localhost:9010", router))

}
