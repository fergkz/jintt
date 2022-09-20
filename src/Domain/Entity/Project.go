package DomainEntity

import (
	"time"
)

type ProjectEpic struct {
	Key     string
	Summary string
	Status  string
}

type ProjectUser struct {
	Id        string
	Email     string
	Name      string
	AvatarUrl string
}

type ProjectComment struct {
	Body      string
	CreatedAt time.Time
	Public    bool
	User      ProjectUser
}

type ProjectSprintId int

type ProjectSprint struct {
	Id           ProjectSprintId
	Name         string
	State        string
	StartDate    time.Time
	EndDate      time.Time
	CompleteDate time.Time
}

type ProjectSubTask struct {
	Key      string
	Status   string
	Done     bool
	Assignee ProjectUser
}

type ProjectTask struct {
	Key               string
	PercentComplete   float64
	Summary           string
	Type              string
	Status            string
	Done              bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Planned           bool
	TimeEstimateHours int
	TimeSpentHours    int
	StartEstimate     time.Time
	PublicHtmlUrl     string
	Epic              ProjectEpic
	Assignee          ProjectUser
	Reporter          ProjectUser
	Subtasks          []ProjectSubTask
	DependenciesKeys  []string
}
