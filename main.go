package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/fcgi"
	"time"

	DomainEntity "github.com/fergkz/jintt/src/Domain/Entity"
	DomainService "github.com/fergkz/jintt/src/Domain/Service"
	InfrastructureController "github.com/fergkz/jintt/src/Infrastructure/Controller"
	InfrastructureService "github.com/fergkz/jintt/src/Infrastructure/Service"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}

	config := new(struct {
		Jira struct {
			Username                     string
			AccessToken                  string
			Hostname                     string
			CustomFieldStartEstimateDate string
			CacheExpiresSeconds          int
			Status                       DomainEntity.ProjectTaskStatusMapping
		}
		Server struct {
			Port string
		}
		Dayoffs []string
		Team    []struct {
			Email       string
			DisplayName string
			Office      string
		}
	})

	viper.Unmarshal(config)

	replaceMembers := map[string]DomainService.RenderHtmlServiceTeamMember{}
	for _, m := range config.Team {
		replaceMembers[m.Email] = DomainService.RenderHtmlServiceTeamMember{
			DisplayName:          m.DisplayName,
			PublicImageUrl:       "",
			Email:                "",
			PercentContribuition: 0,
		}
	}
	dayoffs := []time.Time{}
	for _, d := range config.Dayoffs {
		p, _ := time.Parse("2006-01-02", d)
		dayoffs = append(dayoffs, p)
	}

	router := mux.NewRouter()
	controller := InfrastructureController.NewGanttSprintController(
		InfrastructureService.NewJiraTaskService(config.Jira.Username, config.Jira.AccessToken, config.Jira.Hostname, config.Jira.CustomFieldStartEstimateDate, config.Jira.CacheExpiresSeconds),
		InfrastructureService.NewRenderHtmlService("template.twig"),
		replaceMembers,
		dayoffs,
		config.Jira.Status,
	)
	router.HandleFunc("/sprint/{SprintId:[0-9]+}", controller.Get).Methods("GET")
	router.Handle("/profile-unknow.png", http.FileServer(http.Dir("./public")))

	if viper.GetString("server.method") == "http" {
		fmt.Printf("Server started at port %s", config.Server.Port)
		log.Fatal(http.ListenAndServe("127.0.0.1:"+config.Server.Port, router))
	} else {
		fcgi.Serve(nil, router)
	}

}
