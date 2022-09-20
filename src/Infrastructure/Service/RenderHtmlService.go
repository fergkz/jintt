package InfrastructureService

import (
	"bytes"
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strings"

	DomainService "github.com/fergkz/jintt/src/Domain/Service"
	"github.com/tyler-sommer/stick"
)

type renderHtmlService struct {
	templateFilename string
}

func NewRenderHtmlService(
	templateFilename string,
) *renderHtmlService {
	service := new(renderHtmlService)
	service.templateFilename = templateFilename
	return service
}

func (service renderHtmlService) Parse(RenderSprint DomainService.RenderHtmlServiceSprint) DomainService.RenderHtmlServiceHtmlRendered {

	fullPath, _ := filepath.Abs("public/template.twig")

	dirname, filename := filepath.Split(fullPath)

	render := stick.New(stick.NewFilesystemLoader(dirname))

	/* Normalize task percent */
	for i, row := range RenderSprint.Rows {
		RenderSprint.Rows[i].PercentComplete = math.Round(row.PercentComplete)
	}

	/* Normalize task assignee order */
	for x := range RenderSprint.Rows {
		sort.Slice(RenderSprint.Rows[x].Assignees, func(i, j int) bool {
			if RenderSprint.Rows[x].Assignees[i].PercentContribuition == RenderSprint.Rows[x].Assignees[j].PercentContribuition {
				return RenderSprint.Rows[x].Assignees[i].DisplayName < RenderSprint.Rows[x].Assignees[j].DisplayName
			}
			return RenderSprint.Rows[x].Assignees[i].PercentContribuition > RenderSprint.Rows[x].Assignees[j].PercentContribuition
		})
		for _, assig := range RenderSprint.Rows[x].Assignees {
			assig.PercentContribuition = math.Round(assig.PercentContribuition)
		}
	}

	nParams := map[string]stick.Value{}
	nParams["RenderSprint"] = RenderSprint
	nParams["TimelineLen"] = len(RenderSprint.Timeline)
	nParams["RenderSprintDateStart"] = RenderSprint.DateStart.Format("02/01/2006")
	nParams["RenderSprintDateEnd"] = RenderSprint.DateEnd.Format("02/01/2006")
	TimelineLenValid := 0
	for _, tl := range RenderSprint.Timeline {
		if !tl.Locked {
			TimelineLenValid += 1
		}
	}
	nParams["TimelineLenValid"] = TimelineLenValid

	/* Define Percs */
	RowAssigneesPerc := make(map[string]map[string]int)
	for _, row := range RenderSprint.Rows {
		if _, ok := RowAssigneesPerc[row.TaskKey]; !ok {
			RowAssigneesPerc[row.TaskKey] = map[string]int{}
		}
		for _, assig := range row.Assignees {
			RowAssigneesPerc[row.TaskKey][assig.Email] = int(assig.PercentContribuition)
		}
	}
	nParams["RowAssigneesPerc"] = RowAssigneesPerc

	/* Define Short Names */
	RowAssigneesShort := make(map[string]map[string]string)
	for _, row := range RenderSprint.Rows {
		if _, ok := RowAssigneesShort[row.TaskKey]; !ok {
			RowAssigneesShort[row.TaskKey] = map[string]string{}
		}
		for _, assig := range row.Assignees {
			RowAssigneesShort[row.TaskKey][assig.Email] = strings.Split(assig.DisplayName, " ")[0]
		}
	}
	nParams["RowAssigneesShort"] = RowAssigneesShort

	/* Define TeamMembers */
	TeamMembers := []*DomainService.RenderHtmlServiceSprintTeam{}
	TeamMembersKeysJoin := map[string]string{}
	TeamMembersTotal := map[string]float64{}
	for _, t := range RenderSprint.TeamMembers {
		TeamMembers = append(TeamMembers, t)
		TeamMembersTotal[t.TeamMember.Email] = 0
		for tk := range t.TimelinePerc {
			t.TimelinePerc[tk] = math.Round(t.TimelinePerc[tk])
			TeamMembersTotal[t.TeamMember.Email] += t.TimelinePerc[tk]
		}
		TeamMembersTotal[t.TeamMember.Email] = math.Round(TeamMembersTotal[t.TeamMember.Email])
		TeamMembersKeysJoin[t.TeamMember.Email] = strings.Join(t.TaskKeys, ",")
	}
	sort.Slice(TeamMembers, func(i, j int) bool {
		return TeamMembers[i].TeamMember.DisplayName < TeamMembers[j].TeamMember.DisplayName
	})
	nParams["TeamMembers"] = TeamMembers
	nParams["TeamMembersKeysJoin"] = TeamMembersKeysJoin
	nParams["TeamMembersTotal"] = TeamMembersTotal

	var b bytes.Buffer
	err := render.Execute(filename, &b, nParams)
	if err != nil {
		fmt.Println(err)
	}

	rendered := DomainService.RenderHtmlServiceHtmlRendered{}
	rendered.HtmlContent = b.String()

	return rendered
}
