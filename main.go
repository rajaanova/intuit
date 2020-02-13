package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rajaanova/intuit/controller"
	"net/http"
)

func main() {
	repoController := controller.NewController()
	mux := mux.NewRouter()
	//Get All the repos
	mux.HandleFunc("/allrepo", repoController.AllRepo).Methods("GET")
	//Get Summary of a repo
	mux.HandleFunc("/repo/{reponame}", repoController.SpecificRepo).Methods("GET")
	//Get drill down information such as issue info for a given repo
	mux.HandleFunc("/repo/{reponame}/issues/{issueid}", repoController.RepoIssues).Methods("GET")
	fmt.Println("starting server at port 8080")
	http.ListenAndServe(":8080", mux)
}
