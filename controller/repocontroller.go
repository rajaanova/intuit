package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rajaanova/intuit/model"
	"io/ioutil"
	"net/http"
	"strconv"
)

//Since the data seems to be very less, we can load all the data upfront when the system starts
//And server the request from memory itself
//Another approach could have been to get the data as per request and cache it in that case first request will take some time
//Below solution was to cover most of the features, few of the error checks, nil checks, statuscode checks were left out

type RepoController struct {
	apiInfoByName map[string]model.Api
	issuesByName  map[string]map[int]model.Issue
}

func NewController() *RepoController {
	resp, err := http.Get("https://api.github.com/users/intuit/repos")
	if err != nil {
		fmt.Println("error getting api response  ", err)
		panic(err)
	}
	//TODO check for response status code for now only happy path
	repoResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error reading api response body ", err)
		panic(err)
	}
	res := []model.Api{}
	err = json.Unmarshal(repoResp, &res)
	if err != nil {
		fmt.Println("error unmarshalling the resposne", err)
		panic(err)
	}
	repoController := &RepoController{apiInfoByName: make(map[string]model.Api), issuesByName: make(map[string]map[int]model.Issue)}
	for _, val := range res {
		repoController.apiInfoByName[val.Name] = val
		issueResp, err := http.Get("https://api.github.com/repos/intuit/" + val.Name + "/issues")
		if err != nil {
			//For now panicing here we can devise a solution wherein we can skip the error and later on issue request can be done real time
			panic(fmt.Sprintf("error getting issue reponse for repo %s %v", val.Name, err))
		}
		//TODO check for response status code for now only happy path
		issueRespBytes, err := ioutil.ReadAll(issueResp.Body)
		if err != nil {
			fmt.Println("error reading api response body ", err)
			panic(err)
		}
		issueSlice := []model.Issue{}
		err = json.Unmarshal(issueRespBytes, &issueSlice)
		if err != nil {
			panic(err)
		}
		repoController.issuesByName[val.Name] = make(map[int]model.Issue)
		for _, issuesVal := range issueSlice {
			issueMap := repoController.issuesByName[val.Name]
			issueMap[issuesVal.Number] = issuesVal
		}
	}
	return repoController
}

func (a *RepoController) SpecificRepo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	repoName := params["reponame"]
	//TODO: validation should be done for param
	specific := model.SpecificApiRepo{}
	val := a.apiInfoByName[repoName]
	//assuming that cache has that entry only happy path
	specific.Name = val.Name
	specific.Watchers = val.Watchers
	specific.OpenIssues = val.OpenIssues
	resp, err := json.Marshal(specific)
	if err != nil {
		fmt.Print(err)
		//TODO: send error back
		return
	}
	w.Write(resp)
}

func (a *RepoController) AllRepo(w http.ResponseWriter, r *http.Request) {
	repoSummary := make([]model.RepoSummary, 0)
	for _, val := range a.apiInfoByName {
		repo := model.RepoSummary{}
		repo.Name = val.Name
		repo.Description = val.Description
		repoSummary = append(repoSummary, repo)
	}
	resp, err := json.Marshal(repoSummary)
	if err != nil {
		fmt.Println(err)
		//TODO: send error back
		return
	}
	w.Write(resp)
}

func (a *RepoController) RepoIssues(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	repoName := params["reponame"]
	issueID := params["issueid"]
	issueIDInt, err := strconv.Atoi(issueID)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	//TODO: validation should be done for param
	issuesByRepo := a.issuesByName[repoName]
	if issuesByRepo == nil {
		w.WriteHeader(404)
		return
	}
	if val, ok := issuesByRepo[issueIDInt]; ok {
		w.WriteHeader(200)
		if response, err := json.Marshal(val); err == nil {
			//returning the whole response as of now , but here the conversion between backend and frontend model should be there so as not to pass all the value
			w.Write(response)
			return
		}
	}
	w.WriteHeader(404)

}
