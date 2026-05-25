package jules

import (
	"github.com/SamyRai/go-jules/internal/model"
	"github.com/SamyRai/go-jules/internal/services"
)

type (
	APIError = model.APIError

	SessionState   = model.SessionState
	AutomationMode = model.AutomationMode
	Session        = model.Session
	Output         = model.Output
	PullRequest    = model.PullRequest

	SourceContext     = model.SourceContext
	GithubRepoContext = model.GithubRepoContext
	Source            = model.Source
	GithubRepo        = model.GithubRepo
	Branch            = model.Branch

	ActivityOriginator = model.ActivityOriginator
	Activity           = model.Activity
	PlanGenerated      = model.PlanGenerated
	Plan               = model.Plan
	Step               = model.Step
	PlanApproved       = model.PlanApproved
	UserMessaged       = model.UserMessaged
	AgentMessaged      = model.AgentMessaged
	ProgressUpdated    = model.ProgressUpdated
	SessionCompleted   = model.SessionCompleted
	SessionFailed      = model.SessionFailed

	Artifact   = model.Artifact
	BashOutput = model.BashOutput
	ChangeSet  = model.ChangeSet
	GitPatch   = model.GitPatch
	Media      = model.Media

	CreateSessionRequest = model.CreateSessionRequest
	SendMessageRequest   = model.SendMessageRequest

	SessionsResponse   = model.SessionsResponse
	ActivitiesResponse = model.ActivitiesResponse
	SourcesResponse    = model.SourcesResponse

	ListSessionsOptions   = services.ListSessionsOptions
	ListSourcesOptions    = services.ListSourcesOptions
	ListActivitiesOptions = services.ListActivitiesOptions
	ActivityFilter        = services.ActivityFilter
	ActivitySearchOptions = services.ActivitySearchOptions
	ActivityArtifact      = services.ActivityArtifact
	SessionStatus         = services.SessionStatus

	SessionsService   = services.SessionsService
	SourcesService    = services.SourcesService
	ActivitiesService = services.ActivitiesService
	ArtifactsService  = services.ArtifactsService
	SessionMonitor    = services.SessionMonitor
)

const (
	SessionStateUnspecified          = model.SessionStateUnspecified
	SessionStateQueued               = model.SessionStateQueued
	SessionStatePlanning             = model.SessionStatePlanning
	SessionStateAwaitingPlanApproval = model.SessionStateAwaitingPlanApproval
	SessionStateAwaitingUserFeedback = model.SessionStateAwaitingUserFeedback
	SessionStateInProgress           = model.SessionStateInProgress
	SessionStatePaused               = model.SessionStatePaused
	SessionStateFailed               = model.SessionStateFailed
	SessionStateCompleted            = model.SessionStateCompleted

	AutomationModeUnspecified  = model.AutomationModeUnspecified
	AutomationModeAutoCreatePR = model.AutomationModeAutoCreatePR

	ActivityOriginatorSystem = model.ActivityOriginatorSystem
	ActivityOriginatorAgent  = model.ActivityOriginatorAgent
	ActivityOriginatorUser   = model.ActivityOriginatorUser
)
