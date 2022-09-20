package DomainService

import (
	DomainEntity "github.com/fergkz/jintt/src/Domain/Entity"
)

type TasksRequestService interface {
	GetTasksFromSprints([]DomainEntity.ProjectSprintId) ([]DomainEntity.ProjectTask, []DomainEntity.ProjectSprint)
}
