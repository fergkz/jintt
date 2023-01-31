package InfrastructureService

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"

	DomainEntity "github.com/fergkz/jintt/src/Domain/Entity"
	DomainTool "github.com/fergkz/jintt/src/Domain/Tool"
)

type jiraTaskService struct {
	JiraApiService               *jiraApiService
	Username                     string
	AccessToken                  string
	Hostname                     string
	CustomFieldStartEstimateDate string
	CacheExpiresSeconds          int
}

func NewJiraTaskService(
	Username string,
	AccessToken string,
	Hostname string,
	CustomFieldStartEstimateDate string,
	CacheExpiresSeconds int,
) *jiraTaskService {
	service := new(jiraTaskService)
	service.JiraApiService = NewJiraApiService(Username, AccessToken, Hostname)
	service.Username = Username
	service.AccessToken = AccessToken
	service.Hostname = Hostname
	service.CustomFieldStartEstimateDate = CustomFieldStartEstimateDate
	service.CacheExpiresSeconds = CacheExpiresSeconds
	return service
}

type jiraTaskServiceStructQueryResponse struct {
	Rows []interface{}
}

func (service jiraTaskService) Query(JQL string) (response jiraTaskServiceStructQueryResponse) {

	JQL = url.QueryEscape(JQL)

	var responseFull struct {
		Issues []interface{}
	}

	maxResults := 50
	currentIndex := 0

	for {
		url := service.Hostname + "/rest/api/2/search?jql=" + JQL + "&maxResults=" + strconv.Itoa(maxResults) + "&startAt=" + strconv.Itoa(currentIndex) + "&expand=changelog&fields=*all"

		service.JiraApiService.Get(url, &responseFull)

		if len(responseFull.Issues) == 0 {
			break
		}

		response.Rows = append(response.Rows, responseFull.Issues...)

		currentIndex = currentIndex + maxResults
	}

	return response
}

func (service jiraTaskService) GetSprints(SprintIds []DomainEntity.ProjectSprintId) (rows []DomainEntity.ProjectSprint) {

	for _, i := range SprintIds {
		url := service.Hostname + "/rest/agile/1.0/sprint/" + strconv.Itoa(int(i))
		var response struct {
			StartDate     interface{}
			EndDate       interface{}
			CompleteDate  interface{}
			Goal          string
			Id            int
			Name          string
			OriginBoardId int
			State         string
		}
		service.JiraApiService.Get(url, &response)

		sprint := DomainEntity.ProjectSprint{
			Id:    DomainEntity.ProjectSprintId(response.Id),
			Name:  response.Name,
			State: response.State,
		}

		if response.StartDate != nil {
			sprint.StartDate, _ = time.Parse("2006-01-02T15:04:05.000Z", response.StartDate.(string))
			sprint.StartDate = sprint.StartDate.Add(-3 * time.Hour) // Compatibilizar com horário Jira
			sprint.StartDate, _ = time.Parse("2006-01-02", sprint.StartDate.Format("2006-01-02"))
		}
		if response.EndDate != nil {
			sprint.EndDate, _ = time.Parse("2006-01-02T15:04:05.000Z", response.EndDate.(string))
			sprint.EndDate = sprint.EndDate.Add(-3 * time.Hour) // Compatibilizar com horário Jira
			sprint.EndDate, _ = time.Parse("2006-01-02", sprint.EndDate.Format("2006-01-02"))
			sprint.EndDate = sprint.EndDate.AddDate(0, 0, 1).Add(time.Second * -1)
		}
		if response.CompleteDate != nil {
			sprint.CompleteDate, _ = time.Parse("2006-01-02T15:04:05.000Z", response.CompleteDate.(string))
			sprint.CompleteDate = sprint.CompleteDate.Add(-3 * time.Hour) // Compatibilizar com horário Jira
			sprint.CompleteDate, _ = time.Parse("2006-01-02", sprint.CompleteDate.Format("2006-01-02"))
			sprint.CompleteDate = sprint.CompleteDate.AddDate(0, 0, 1).Add(time.Second * -1)
		}

		rows = append(rows, sprint)
	}

	return rows
}

func (service jiraTaskService) GetTasksFromSprints(SprintIds []DomainEntity.ProjectSprintId) (Tasks []DomainEntity.ProjectTask, Sprints []DomainEntity.ProjectSprint) {

	Sprints = service.GetSprints(SprintIds)

	sprintsIdsStrs := []string{}
	for _, i := range SprintIds {
		sprintsIdsStrs = append(sprintsIdsStrs, strconv.Itoa(int(i)))
	}
	sprintsIdsStr := strings.Join(sprintsIdsStrs, ",")

	response := new(jiraTaskServiceStructQueryResponse)

	cacheFilename := "cache/sprint-" + sprintsIdsStr

	if !DomainTool.Pretty.GetCache(cacheFilename, response) {
		jql := `sprint IN (` + sprintsIdsStr + `) order by rank ASC`
		*response = service.Query(jql)
		DomainTool.Pretty.SetCache(cacheFilename, response, service.CacheExpiresSeconds)
	}

	Tasks = service.parseToTasks(response.Rows, Sprints)

	return Tasks, Sprints
}

func (service jiraTaskService) parseToTasks(rows []interface{}, sprints []DomainEntity.ProjectSprint) (Tasks []DomainEntity.ProjectTask) {

	type taskDTOStruct struct {
		Id           int `json:",string"`
		Key          string
		FieldsStruct struct {
			Summary   string
			Issuetype struct {
				Name           string
				HierarchyLevel int
				Subtask        bool
			}
			Resolution struct {
				Name string
			}
			Resolutiondate interface{} //"2022-09-01T15:05:11.156-0300"
			Status         struct {
				Name           string
				StatusCategory struct {
					Key string // done
				}
			}
			TimeOriginalEstimate int
			TimeSpent            int
			Assignee             struct {
				AccountId    string
				EmailAddress string
				DisplayName  string
				AvatarUrls   struct {
					Image string `json:"48x48"`
				}
			}
			Reporter struct {
				AccountId    string
				EmailAddress string
				DisplayName  string
				AvatarUrls   struct {
					Image string `json:"48x48"`
				}
			}
			Parent struct {
				Key    string
				Fields struct {
					Summary string
					Status  struct {
						Name string
					}
				}
			}
			Subtasks []struct {
				Key    string
				Fields struct {
					Status struct {
						Name string
					}
				}
			}
			Updated  interface{}
			Created  interface{}
			Progress struct {
				Percent float64
			}
			Comment struct {
				Comments []struct {
					Author struct {
						AccountId    string
						EmailAddress string
						DisplayName  string
						AvatarUrls   struct {
							Image string `json:"48x48"`
						}
					}
					Body      string
					Created   interface{} //"2022-09-01T15:05:11.156-0300"
					Updated   interface{} //"2022-09-01T15:05:11.156-0300"
					JsdPublic bool
				}
			}
			Issuelinks []struct {
				OutwardIssue struct {
					Fields struct {
						Issuetype struct {
							Subtask bool
						}
						Summary string
					}
					Key string
				}
				Type struct {
					Inward  string
					Name    string
					Outward string
				}
			}
		}
		FieldsMap map[string]interface{} `json:"fields"`
	}

	allDTOs := map[string]taskDTOStruct{}
	allOrderedKeys := []string{}

	for _, row := range rows {
		dto := new(taskDTOStruct)
		byteRow, _ := json.Marshal(row)
		json.Unmarshal(byteRow, &dto)

		dbByte, _ := json.Marshal(dto.FieldsMap)
		_ = json.Unmarshal(dbByte, &dto.FieldsStruct)

		allDTOs[dto.Key] = *dto

		if !dto.FieldsStruct.Issuetype.Subtask {
			allOrderedKeys = append(allOrderedKeys, dto.Key)
		}
	}

	for _, key := range allOrderedKeys {
		dto := allDTOs[key]

		Task := DomainEntity.ProjectTask{}

		Task.Key = dto.Key
		Task.Summary = dto.FieldsStruct.Summary
		Task.Type = dto.FieldsStruct.Issuetype.Name
		Task.Status = dto.FieldsStruct.Status.Name

		dtoStartEstimate := dto.FieldsMap[service.CustomFieldStartEstimateDate]
		if dtoStartEstimate != nil {
			Task.StartEstimate, _ = time.Parse("2006-01-02", dtoStartEstimate.(string))
		}

		if dto.FieldsStruct.Status.StatusCategory.Key == "done" {
			Task.Done = true
		}

		Task.PercentComplete = 0
		Task.TimeEstimateHours = dto.FieldsStruct.TimeOriginalEstimate

		Task.CreatedAt, _ = time.Parse("2006-01-02T15:04:05.000-0700", dto.FieldsStruct.Created.(string))
		Task.CreatedAt = Task.CreatedAt.Add(-3 * time.Hour) // Compatibilizar com horário Jira
		Task.CreatedAt, _ = time.Parse("2006-01-02", Task.CreatedAt.Format("2006-01-02"))

		Task.UpdatedAt, _ = time.Parse("2006-01-02T15:04:05.000-0700", dto.FieldsStruct.Updated.(string))
		Task.UpdatedAt = Task.UpdatedAt.Add(-3 * time.Hour) // Compatibilizar com horário Jira
		Task.UpdatedAt, _ = time.Parse("2006-01-02", Task.UpdatedAt.Format("2006-01-02"))
		Task.UpdatedAt = Task.UpdatedAt.AddDate(0, 0, 1).Add(time.Second * -1)

		if dto.FieldsStruct.TimeOriginalEstimate > 0 {
			Task.TimeEstimateHours = int(float64(dto.FieldsStruct.TimeOriginalEstimate) / 60 / 60)
		}
		if dto.FieldsStruct.TimeSpent > 0 {
			Task.TimeSpentHours = int(float64(dto.FieldsStruct.TimeSpent) / 60 / 60)
		}

		Task.Assignee = DomainEntity.ProjectUser{
			Id:        dto.FieldsStruct.Assignee.AccountId,
			Email:     dto.FieldsStruct.Assignee.EmailAddress,
			Name:      dto.FieldsStruct.Assignee.DisplayName,
			AvatarUrl: dto.FieldsStruct.Assignee.AvatarUrls.Image,
		}

		countSubDone := 0
		for _, subtask := range dto.FieldsStruct.Subtasks {
			dtoSub := allDTOs[subtask.Key]

			sub := DomainEntity.ProjectSubTask{}
			sub.Key = dtoSub.Key
			sub.Status = dtoSub.FieldsStruct.Status.Name
			sub.Assignee = DomainEntity.ProjectUser{
				Id:        dtoSub.FieldsStruct.Assignee.AccountId,
				Email:     dtoSub.FieldsStruct.Assignee.EmailAddress,
				Name:      dtoSub.FieldsStruct.Assignee.DisplayName,
				AvatarUrl: dtoSub.FieldsStruct.Assignee.AvatarUrls.Image,
			}

			if dtoSub.FieldsStruct.Status.StatusCategory.Key == "done" {
				sub.Done = true
				countSubDone++
			}

			if Task.Key == "GPT-10350" {

			}

			Task.Subtasks = append(Task.Subtasks, sub)
		}

		if countSubDone > 0 && len(Task.Subtasks) > 0 {
			Task.PercentComplete = float64(float64(countSubDone) / float64(len(Task.Subtasks)) * 100)
		}

		Task.Reporter = DomainEntity.ProjectUser{
			Id:        dto.FieldsStruct.Reporter.AccountId,
			Email:     dto.FieldsStruct.Reporter.EmailAddress,
			Name:      dto.FieldsStruct.Reporter.DisplayName,
			AvatarUrl: dto.FieldsStruct.Reporter.AvatarUrls.Image,
		}

		Task.Epic = DomainEntity.ProjectEpic{
			Key:     dto.FieldsStruct.Parent.Key,
			Summary: dto.FieldsStruct.Parent.Fields.Summary,
			Status:  dto.FieldsStruct.Parent.Fields.Status.Name,
		}

		for _, iLink := range dto.FieldsStruct.Issuelinks {
			if iLink.OutwardIssue.Key != "" {
				Task.DependenciesKeys = append(Task.DependenciesKeys, iLink.OutwardIssue.Key)
			}
		}

		for _, sprint := range sprints {
			if Task.CreatedAt.Before(sprint.StartDate) {
				Task.Planned = true
			}
		}

		Task.PublicHtmlUrl = service.Hostname + "/browse/" + Task.Key

		Tasks = append(Tasks, Task)
	}

	return Tasks
}
