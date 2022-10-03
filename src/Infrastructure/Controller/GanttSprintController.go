package InfrastructureController

import (
	"io"
	"net/http"
	"strconv"
	"time"

	ApplicationUsecase "github.com/fergkz/jintt/src/Application/Usecase"
	DomainEntity "github.com/fergkz/jintt/src/Domain/Entity"
	DomainService "github.com/fergkz/jintt/src/Domain/Service"
	"github.com/gorilla/mux"
)

type GanttSprintController struct {
	TasksRequestService DomainService.TasksRequestService
	RenderHtmlService   DomainService.RenderHtmlService
	ReplaceTeamMembers  map[string]DomainService.RenderHtmlServiceTeamMember
	Dayoffs             []time.Time
	StatusMapping       DomainEntity.ProjectTaskStatusMapping
}

func NewGanttSprintController(
	TasksRequestService DomainService.TasksRequestService,
	RenderHtmlService DomainService.RenderHtmlService,
	ReplaceTeamMembers map[string]DomainService.RenderHtmlServiceTeamMember,
	Dayoffs []time.Time,
	StatusMapping DomainEntity.ProjectTaskStatusMapping,
) *GanttSprintController {
	controller := new(GanttSprintController)
	controller.TasksRequestService = TasksRequestService
	controller.RenderHtmlService = RenderHtmlService
	controller.ReplaceTeamMembers = ReplaceTeamMembers
	controller.Dayoffs = Dayoffs
	controller.StatusMapping = StatusMapping
	return controller
}

func (controller *GanttSprintController) Get(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,access-control-allow-origin, access-control-allow-headers")

	sprintId := mux.Vars(r)["SprintId"]
	sprintIdInt, _ := strconv.Atoi(sprintId)

	usecase := ApplicationUsecase.NewGenerateGanttHtml(
		controller.TasksRequestService,
		controller.RenderHtmlService,
		controller.ReplaceTeamMembers,
		controller.Dayoffs,
		controller.StatusMapping,
	)
	html := usecase.Run([]DomainEntity.ProjectSprintId{DomainEntity.ProjectSprintId(sprintIdInt)})

	io.WriteString(w, string(html))
}
