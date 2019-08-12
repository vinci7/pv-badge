package main

import (
	"os"
	"log"
	"fmt"
	"bytes"
	"strings"
	"strconv"
	"net/http"
	"encoding/json"
	"io/ioutil"

	"github.com/narqo/go-badge"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	. "github.com/vinci7/pv-badge/conf"
)


var (
	XLcId string
	XLcKey string
	contentType string
	totalPvUrl string
	todayPvUrl string
)

type Result struct {
	RepoID string `json:"repo_id"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Pv int `json:"pv"`
	ObjectID string `json:"objectId"`
}

type TotalPv struct {
	Results []Result `json:"results"`
}

func main(){

	//init config
	err := InitConfig()

	if err != nil {
		log.Fatal(err)
	} else {
		XLcId = Conf.XLcId
		XLcKey = Conf.XLcKey
		contentType = Conf.ContentType
		totalPvUrl = Conf.TotalPvUrl
		todayPvUrl = Conf.TodayPvUrl
	}

	e := echo.New()

	/* total-pv

	   return a svg badge with total's pv number of 'repo_id'

	 */
	e.GET("/total.svg", totalSvc)
	e.GET("/total.svg/:repo_id", totalSvc)


	/* today-pv

	   return a svg badge with today's pv number of 'repo_id'

	 */
	e.GET("today.svg", todaySvc)
	e.GET("today.svg/:repo_id", todaySvc)


	/* health check

	   expose API for health check service

	 */
	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})


	//enable logger
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339}: method=${method}, uri=${uri}, status=${status}\n",
	}))


	//get the port from env
	port := "7777"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}


	//Run echo service
	e.Start(":" + port)
}


func totalSvc(c echo.Context) error {
	repoID := c.QueryParam("repo_id")
	if repoID == "" {
		repoID = c.Param("repo_id")
	}

	var pv int

	if result := isRepoIdExist(repoID); result.ObjectID != "" {
		pv = accessRepoId(result)
	} else {
		pv = createRepoId(repoID)
	}


	var buf bytes.Buffer

	err := badge.Render("Total Visits", strconv.Itoa(pv), "#5272B4", &buf)

	if err != nil {
		log.Fatal(err)
	}

	c.Response().Header().Set("Cache-Control", "no-cache,max-age=0")
	// set a expired date to unable cache
	c.Response().Header().Set("Expires", "Mon, 12 Aug 2019 14:09:23 GMT")
	return c.Stream(http.StatusOK, "image/svg+xml", &buf)
}

func todaySvc(c echo.Context) error {
	repoID := c.QueryParam("repo_id")
	if repoID == "" {
		repoID = c.Param("repo_id")
	}

	var buf bytes.Buffer

	err := badge.Render("Visits Today", "100", "#5272B4", &buf)

	if err != nil {
		log.Fatal(err)
	}


	c.Response().Header().Set("Cache-Control", "no-cache,max-age=0")
	c.Response().Header().Set("Expires", "Mon, 12 Aug 2019 14:09:23 GMT")
	return c.Stream(http.StatusOK, "image/svg+xml", &buf)
}

func isRepoIdExist(repoID string) Result {

	client := &http.Client{}

	request, err := http.NewRequest("GET", totalPvUrl, nil)

	if err != nil {
		log.Fatal(err)
	}

	request.Header.Add("X-LC-Id", XLcId)
	request.Header.Add("X-LC-Key", XLcKey)
	request.Header.Add("Content-Type", contentType)


	q := request.URL.Query()
	where := fmt.Sprintf("{\"repo_id\":\"%s\"}", repoID)
	q.Add("where", where)
	request.URL.RawQuery = q.Encode()

	response, err := client.Do(request)

	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatal(err)
	}

	totalPv := TotalPv{}

	err = json.Unmarshal(body, &totalPv)

	if err != nil {
		log.Fatal(err)
	}

	log.Println(totalPv)

	if len(totalPv.Results) == 0 {
		return Result{}
	}


	return totalPv.Results[0]
}

func accessRepoId(result Result) int {

	client := &http.Client{}

	payload := strings.NewReader("{\"pv\":{\"__op\":\"Increment\",\"amount\":1}}")

	request, err := http.NewRequest("PUT", totalPvUrl+"/"+result.ObjectID, payload)

	if err != nil {
		log.Fatal(err)
	}

	request.Header.Add("X-LC-Id", XLcId)
	request.Header.Add("X-LC-Key", XLcKey)
	request.Header.Add("Content-Type", contentType)

	response, err := client.Do(request)

	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatal(err)
	}



	r := Result{}

	err = json.Unmarshal(body, &r)

	if r.UpdatedAt != "" {
		return result.Pv+1
	}

	return result.Pv

}

func createRepoId(repoID string) int {

	client := &http.Client{}

	payload := strings.NewReader("{\"repo_id\": \""+repoID+"\", \"pv\": 1}")

	log.Println(payload)

	request, err := http.NewRequest("POST", totalPvUrl, payload)

	if err != nil {
		log.Fatal(err)
	}

	request.Header.Add("X-LC-Id", XLcId)
	request.Header.Add("X-LC-Key", XLcKey)
	request.Header.Add("Content-Type", contentType)
	request.Header.Add("Cache-Control", "no-cache,max-age=0")

	response, err := client.Do(request)

	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()


	if response.StatusCode == http.StatusCreated {
		return 1
	}

	return 0

}