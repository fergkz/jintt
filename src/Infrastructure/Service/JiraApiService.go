package InfrastructureService

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type jiraApiService struct {
	Username    string
	AccessToken string
	Hostname    string
}

func NewJiraApiService(
	Username string,
	AccessToken string,
	Hostname string,
) *jiraApiService {
	service := new(jiraApiService)
	service.Username = Username
	service.AccessToken = AccessToken
	service.Hostname = Hostname
	return service
}

func (service jiraApiService) Get(url string, responseFull interface{}) (err error) {

	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		log.Panicln("JiraApiService", "ERROR:http.NewRequest\n", err)
		return err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(service.Username + ":" + service.AccessToken))

	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err == nil && res.StatusCode != 200 {
		err = errors.New("STATUS RESPONSE: " + strconv.Itoa(res.StatusCode) + " - " + http.StatusText(res.StatusCode))
		log.Panicln("ERROR: ", err)
	}
	if err != nil {
		log.Panicln("JiraApiService", "ERROR:client.Do\n", err)
		return err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Panicln("JiraApiService", "ERROR:ioutil.ReadAll\n", err)
		return err
	}

	json.Unmarshal(body, &responseFull)

	return nil
}
