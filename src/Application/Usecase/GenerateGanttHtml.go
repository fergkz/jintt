package ApplicationUsecase

import (
	"math"
	"sort"
	"strings"
	"time"

	DomainEntity "github.com/fergkz/jintt/src/Domain/Entity"
	DomainService "github.com/fergkz/jintt/src/Domain/Service"
)

type generateGanttHtml struct {
	TasksRequestService DomainService.TasksRequestService
	RenderHtmlService   DomainService.RenderHtmlService
	ReplaceTeamMembers  map[string]DomainService.RenderHtmlServiceTeamMember
	Dayoffs             []time.Time
	StatusMapping       DomainEntity.ProjectTaskStatusMapping
	WorkingHourPerDay   float64
}

func NewGenerateGanttHtml(
	TasksRequestService DomainService.TasksRequestService,
	RenderHtmlService DomainService.RenderHtmlService,
	ReplaceTeamMembers map[string]DomainService.RenderHtmlServiceTeamMember,
	Dayoffs []time.Time,
	StatusMapping DomainEntity.ProjectTaskStatusMapping,
	WorkingHourPerDay float64,
) *generateGanttHtml {
	usecase := new(generateGanttHtml)
	usecase.TasksRequestService = TasksRequestService
	usecase.RenderHtmlService = RenderHtmlService
	usecase.ReplaceTeamMembers = ReplaceTeamMembers
	usecase.Dayoffs = Dayoffs
	usecase.StatusMapping = StatusMapping
	usecase.WorkingHourPerDay = WorkingHourPerDay
	return usecase
}

func (usecase *generateGanttHtml) Run(sprintIds []DomainEntity.ProjectSprintId) string {

	tasks, sprint := usecase.TasksRequestService.GetTasksFromSprints(sprintIds)

	RenderSprint := DomainService.RenderHtmlServiceSprint{}
	RenderSprint.Title = sprint[0].Name
	RenderSprint.DateStart = sprint[0].StartDate
	RenderSprint.DateEnd = sprint[0].EndDate
	RenderSprint.Timeline = []*DomainService.RenderHtmlServiceTimelineStep{}
	RenderSprint.TeamMembers = map[string]*DomainService.RenderHtmlServiceSprintTeam{}

	allTasks := map[string]*DomainEntity.ProjectTask{}
	allRows := map[string]*DomainService.RenderHtmlServiceRow{}
	for _, task := range tasks {
		allTasks[task.Key] = &task
	}

	for _, task := range tasks {

		row := new(DomainService.RenderHtmlServiceRow)
		row.TaskKey = task.Key
		row.TaskName = task.Summary
		row.EpicKey = task.Epic.Key
		row.EpicName = task.Epic.Summary
		row.CreateDate = task.CreatedAt
		row.TaskDone = task.Done

		if task.StartEstimate.IsZero() {
			row.StartDate = usecase.fixValidDate(RenderSprint.DateStart)
			if task.CreatedAt.After(RenderSprint.DateStart) {
				row.StartDate = usecase.fixValidDate(task.CreatedAt)
			}
		} else {
			row.StartDate = usecase.fixValidDate(task.StartEstimate)
		}
		row.DurationHours = task.TimeEstimateHours
		row.DurationDays = int(math.Ceil(float64(task.TimeEstimateHours) / float64(usecase.WorkingHourPerDay)))
		row.EndDate = row.StartDate.AddDate(0, 0, 1).Add(time.Second * -1)
		row.EndDate = usecase.sumDays(row.EndDate, row.DurationDays-1)

		row.PercentComplete = task.PercentComplete
		row.Status = task.Status
		row.LinkPublicUrl = task.PublicHtmlUrl

		for _, dK := range task.DependenciesKeys {
			if _, ok := allTasks[dK]; ok {
				row.DependencyKeys = append(row.DependencyKeys, dK)
			}
		}

		row.MainAssignee = &DomainService.RenderHtmlServiceTeamMember{
			DisplayName:    task.Assignee.Name,
			Email:          task.Assignee.Email,
			PublicImageUrl: task.Assignee.AvatarUrl,
		}

		/* Assignees */
		rowTeamIsset := map[string]*DomainService.RenderHtmlServiceTeamMember{}
		if row.MainAssignee != nil {
			rowTeamIsset[row.MainAssignee.Email] = row.MainAssignee
		}

		var sumAllContribuition float64 = 0
		for _, sub := range task.Subtasks {
			assigneeEmail := task.Assignee.Email
			if sub.Assignee.Email != "" {
				assigneeEmail = sub.Assignee.Email
				if _, ok := rowTeamIsset[sub.Assignee.Email]; !ok {
					rowTeamIsset[sub.Assignee.Email] = &DomainService.RenderHtmlServiceTeamMember{
						DisplayName:          sub.Assignee.Name,
						Email:                sub.Assignee.Email,
						PublicImageUrl:       sub.Assignee.AvatarUrl,
						PercentContribuition: 0,
					}
				}
			}

			var nextContrib float64 = 1 / float64(len(task.Subtasks)) * 100

			if sumAllContribuition >= 100 {
				nextContrib = 0
			}

			if (sumAllContribuition + nextContrib) >= 100 {
				nextContrib = 100 - sumAllContribuition
			}

			rowTeamIsset[assigneeEmail].PercentContribuition += nextContrib
			sumAllContribuition += nextContrib
		}

		for _, rt := range rowTeamIsset {
			row.Assignees = append(row.Assignees, rt)
		}

		allRows[task.Key] = row
		RenderSprint.Rows = append(RenderSprint.Rows, row)
	}

	/* REPLACE START AND END DATE */
	for range allRows {
		for rid := range allRows {
			row := allRows[rid]
			for _, depK := range row.DependencyKeys {
				if _, ok := allRows[depK]; ok {
					if row.StartDate.Before(usecase.sumDays(allRows[depK].EndDate, 1)) {
						row.StartDate = usecase.fixValidDate(allRows[depK].EndDate.Add(time.Second * 1))
						row.EndDate = usecase.fixValidDate(usecase.sumDays(row.StartDate, row.DurationDays-1))
					}
				}
			}
			allRows[rid] = row
		}
	}
	/* END: REPLACE START AND END DATE */

	/* DEFINE TIMELINE */
	minDate := RenderSprint.DateStart
	maxDate := RenderSprint.DateEnd
	for _, row := range allRows {
		if row.StartDate.Before(minDate) {
			minDate = row.StartDate
		}
		if row.EndDate.After(maxDate) {
			maxDate = row.EndDate
		}
	}
	RenderSprint.Timeline = usecase.createTimeline(minDate, maxDate, allRows)
	for _, row := range allRows {
		row.TimelineStartKey = row.StartDate.Format("2006-01-02")
		row.TimelineEndKey = row.EndDate.Format("2006-01-02")
		for idx, tl := range RenderSprint.Timeline {
			if tl.Key == row.TimelineStartKey {
				row.TimelineStartIndex = idx
			}
			if tl.Key == row.TimelineEndKey {
				row.TimelineEndIndex = idx
			}
		}
	}
	/* END: DEFINE TIMELINE */

	/* DEFINE TEAM ALOCATION */
	for _, row := range allRows {
		for _, assig := range row.Assignees {
			if _, ok := RenderSprint.TeamMembers[assig.Email]; !ok {
				RenderSprint.TeamMembers[assig.Email] = new(DomainService.RenderHtmlServiceSprintTeam)
				RenderSprint.TeamMembers[assig.Email].TeamMember = *assig
				RenderSprint.TeamMembers[assig.Email].TeamMember.PercentContribuition = 0
				RenderSprint.TeamMembers[assig.Email].TimelinePerc = map[string]float64{}

				for _, t := range RenderSprint.Timeline {
					RenderSprint.TeamMembers[assig.Email].TimelinePerc[t.Key] = 0
				}
			}

			RenderSprint.TeamMembers[assig.Email].TaskKeys = append(RenderSprint.TeamMembers[assig.Email].TaskKeys, row.TaskKey)
		}
	}
	for _, tl := range RenderSprint.Timeline {
		for _, row := range allRows {
			if usecase.inTimeSpan(tl.Date, row.StartDate, row.EndDate) && usecase.isValidDate(tl.Date) {
				for _, assig := range row.Assignees {
					RenderSprint.TeamMembers[assig.Email].TimelinePerc[tl.Key] += math.Round(assig.PercentContribuition)
				}
			}
		}
	}
	/* END: DEFINE TEAM ALOCATION */

	/* DEFINE STATUS MAPPED */
	for _, row := range RenderSprint.Rows {
		row.StatusMapped = "warning"

		if usecase.strInSlice(strings.ToUpper(row.Status), usecase.StatusMapping.Done) {
			if row.PercentComplete < 100 {
				row.StatusMapped = "warning"
				continue
			}
			row.StatusMapped = "success"
			continue
		} else {
			if time.Now().After(row.EndDate) {
				row.StatusMapped = "danger"
				continue
			}
		}

		if usecase.strInSlice(strings.ToUpper(row.Status), usecase.StatusMapping.Executing) {
			if row.PercentComplete >= 100 {
				row.StatusMapped = "warning"
				continue
			}
			if time.Now().After(row.EndDate) {
				row.StatusMapped = "danger"
				continue
			}
			if time.Now().Before(row.StartDate) {
				row.StatusMapped = "success"
				continue
			}
			row.StatusMapped = "normal"
			continue
		}
		if usecase.strInSlice(strings.ToUpper(row.Status), usecase.StatusMapping.Stopped) {
			row.StatusMapped = "warning"
			continue
		}
	}
	/* END: DEFINE STATUS MAPPED */

	for i, row := range RenderSprint.Rows {
		RenderSprint.Rows[i] = row
	}

	return usecase.RenderHtmlService.Parse(RenderSprint).HtmlContent
}

func (usecase *generateGanttHtml) createTimeline(minDate time.Time, maxDate time.Time, rows map[string]*DomainService.RenderHtmlServiceRow) []*DomainService.RenderHtmlServiceTimelineStep {
	weekParse := map[string]string{}
	weekParse["Sunday"] = "Dom"
	weekParse["Monday"] = "Seg"
	weekParse["Tuesday"] = "Ter"
	weekParse["Wednesday"] = "Qua"
	weekParse["Thursday"] = "Qui"
	weekParse["Friday"] = "Sex"
	weekParse["Saturday"] = "Sáb"

	days := maxDate.Sub(minDate).Hours() / 24
	timelines := map[string]*DomainService.RenderHtmlServiceTimelineStep{}
	dt := minDate.AddDate(0, 0, -1)
	for i := 0; i <= int(days); i++ {
		dt = dt.AddDate(0, 0, 1)
		if dt.After(maxDate) {
			break
		}
		timelines[dt.Format("2006-01-02")] = &DomainService.RenderHtmlServiceTimelineStep{
			Key:              dt.Format("2006-01-02"),
			Date:             dt,
			Locked:           !usecase.isValidDate(dt),
			DisplayShortName: weekParse[dt.Weekday().String()] + " " + dt.Format("02/01"),
			DisplayFullName:  dt.Format("2006-01-02T15:04:05.000Z"),
			Today:            dt.Format("2006-01-02") == time.Now().Format("2006-01-02"),
		}
	}

	for _, row := range rows {
		for tk, tl := range timelines {
			if usecase.inTimeSpan(tl.Date, row.StartDate, row.EndDate) {
				timelines[tk].CrossTaskKeys = append(timelines[tk].CrossTaskKeys, row.TaskKey)
			}
		}
	}

	retTimelines := []*DomainService.RenderHtmlServiceTimelineStep{}
	for _, t := range timelines {
		retTimelines = append(retTimelines, t)
	}
	sort.Slice(retTimelines, func(i, j int) bool {
		return retTimelines[i].Date.Before(retTimelines[j].Date)
	})

	return retTimelines
}

func (usecase *generateGanttHtml) strInSlice(text string, slice []string) bool {
	for _, t := range slice {
		if t == text {
			return true
		}
	}
	return false
}

func (usecase *generateGanttHtml) isValidDate(dt time.Time) bool {

	if dt.Weekday().String() == "Sunday" {
		return false
	}

	if dt.Weekday().String() == "Saturday" {
		return false
	}

	for _, dof := range usecase.Dayoffs {
		ds, _ := time.Parse("2006-01-02", dof.Format("2006-01-02"))
		de, _ := time.Parse("2006-01-02", dof.Format("2006-01-02"))
		de = de.AddDate(0, 0, 1).Add(time.Second * -1)
		if usecase.inTimeSpan(dt, ds, de) {
			return false
		}
	}

	return true
}

func (usecase *generateGanttHtml) fixValidDate(dt time.Time) time.Time {

	if !usecase.isValidDate(dt) {
		return usecase.fixValidDate(usecase.sumDays(dt, 1))
	}

	return dt
}

func (usecase *generateGanttHtml) sumDays(dt time.Time, days int) time.Time {
	for i := 1; i <= days; i++ {
		dt = usecase.fixValidDate(dt.AddDate(0, 0, 1))
	}
	return dt
}

func (usecase *generateGanttHtml) inTimeSpan(check, start, end time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return !start.After(check) || !end.Before(check)
}
