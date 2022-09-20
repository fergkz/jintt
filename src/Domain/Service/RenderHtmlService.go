package DomainService

import "time"

type RenderHtmlServiceRow struct {
	TaskKey            string
	TaskName           string
	EpicKey            string
	EpicName           string
	StartDate          time.Time
	EndDate            time.Time
	CreateDate         time.Time
	DurationHours      int
	DurationDays       int
	PercentComplete    float64
	Status             string
	LinkPublicUrl      string
	MainAssignee       *RenderHtmlServiceTeamMember
	DependencyKeys     []string
	Assignees          []*RenderHtmlServiceTeamMember
	TimelineStartKey   string
	TimelineEndKey     string
	TimelineStartIndex int
	TimelineEndIndex   int
}

type RenderHtmlServiceTeamMember struct {
	DisplayName          string
	PublicImageUrl       string
	Email                string
	PercentContribuition float64
}

type RenderHtmlServiceTimelineStep struct {
	Key              string
	Date             time.Time
	Locked           bool
	Today            bool
	DisplayShortName string
	DisplayFullName  string
	CrossTaskKeys    []string
}

type RenderHtmlServiceSprintTeam struct {
	TeamMember   RenderHtmlServiceTeamMember
	TimelinePerc map[string]float64
	TaskKeys     []string
}

type RenderHtmlServiceSprint struct {
	Title       string
	DateStart   time.Time
	DateEnd     time.Time
	Rows        []*RenderHtmlServiceRow
	Timeline    []*RenderHtmlServiceTimelineStep
	TeamMembers map[string]*RenderHtmlServiceSprintTeam
}

type RenderHtmlServiceHtmlRendered struct {
	HtmlContent string
}

type RenderHtmlService interface {
	Parse(sprint RenderHtmlServiceSprint) RenderHtmlServiceHtmlRendered
}
