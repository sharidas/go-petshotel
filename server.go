package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sharidas/go-petshotel/config"
	"github.com/sharidas/go-petshotel/user/controllers"
)

func main() {

	//Read the config
	configData := &config.Configuration{}

	configData.ConfigParser()

	fmt.Println("configData = ", configData)

	fmt.Println("Listening on port 5000!")

	r := mux.NewRouter()
	r.StrictSlash(true)

	controllers.Init(r)
	srv := &http.Server{
		Handler:      r,
		Addr:         ":5000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	//Walk and list the routes
	r.Walk(func(route *mux.Route, rotuer *mux.Router, ancestors []*mux.Route) error {
		t, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		fmt.Println(t)
		return nil
	})
	log.Fatal(srv.ListenAndServe())
}

func urlTrimMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.Trim(r.URL.Path, "/")
		next.ServeHTTP(w, r)
	})
}
