package service

import (
	stdcontext "context"
	"io"
	"time"

	"github.com/google/uuid"

	"git.ykonkov.com/ykonkov/survey-bot/internal/context"
	"git.ykonkov.com/ykonkov/survey-bot/internal/entity"
)

type (
	ResultsFilter struct {
		From *time.Time
		To   *time.Time
	}

	UserSurveyState struct {
		UserGUID  uuid.UUID
		Survey    entity.Survey
		State     entity.State
		IsCurrent bool
	}

	UserListResponse struct {
		Users []UserReport `json:"users"`
		Total int          `json:"total"`
	}

	UserReport struct {
		GUID              uuid.UUID `json:"guid"`
		NickName          string    `json:"nick_name"`
		CompletedTests    int       `json:"completed_tests"`
		AnsweredQuestions int       `json:"answered_questions"`
		RegisteredAt      time.Time `json:"registered_at"`
		LastActivity      time.Time `json:"last_activity"`
	}

	Service interface {
		HandleResultsCommand(ctx context.Context, f ResultsFilter) error
		HandleStartCommand(ctx context.Context) error
		HandleSurveyCommand(ctx context.Context, surveyID int64) error
		HandleListCommand(ctx context.Context) error
		HandleAnswer(ctx context.Context, msg string) error

		GetCompletedSurveys(ctx stdcontext.Context, userID int64) ([]entity.SurveyStateReport, error)
		GetUsersList(ctx stdcontext.Context, limit, offset int, search string) (UserListResponse, error)

		SaveFinishedSurveys(ctx stdcontext.Context, tx DBTransaction, w io.Writer, f ResultsFilter, batchSize int) (int, error)
		CreateSurvey(ctx stdcontext.Context, s entity.Survey) (entity.Survey, error)

		// Updates "name", "questions" and "calculations_type" fields.
		UpdateSurvey(ctx stdcontext.Context, s entity.Survey) error
	}

	TelegramRepo interface {
		SendSurveyList(ctx context.Context, states []UserSurveyState) error
		SendSurveyQuestion(ctx context.Context, question entity.Question) error
		SendMessage(ctx context.Context, msg string) error
		SendFile(ctx context.Context, path string) error
	}

	DBRepo interface {
		BeginTx(ctx stdcontext.Context) (DBTransaction, error)

		CreateSurvey(ctx stdcontext.Context, exec DBTransaction, s entity.Survey) error
		DeleteSurvey(ctx stdcontext.Context, exec DBTransaction, surveyGUID uuid.UUID) error
		GetSurveysList(ctx stdcontext.Context, exec DBTransaction) ([]entity.Survey, error)
		GetSurvey(ctx stdcontext.Context, exec DBTransaction, surveGUID uuid.UUID) (entity.Survey, error)
		GetSurveyByID(ctx stdcontext.Context, exec DBTransaction, surveyID int64) (entity.Survey, error)
		UpdateSurvey(ctx stdcontext.Context, exec DBTransaction, s entity.Survey) error

		GetUserByGUID(ctx stdcontext.Context, exec DBTransaction, userGUID uuid.UUID) (entity.User, error)
		GetUserByID(ctx stdcontext.Context, exec DBTransaction, userID int64) (entity.User, error)
		CreateUser(ctx stdcontext.Context, exec DBTransaction, user entity.User) error
		UpdateUserCurrentSurvey(ctx stdcontext.Context, exec DBTransaction, userGUID uuid.UUID, surveyGUID uuid.UUID) error
		UpdateUserLastActivity(ctx stdcontext.Context, exec DBTransaction, userGUID uuid.UUID) error
		SetUserCurrentSurveyToNil(ctx stdcontext.Context, exec DBTransaction, userGUID uuid.UUID) error
		GetCompletedSurveys(ctx stdcontext.Context, exec DBTransaction, userGUID uuid.UUID) ([]entity.SurveyStateReport, error)
		GetUsersList(ctx stdcontext.Context, exec DBTransaction, limit, offset int, search string) (UserListResponse, error)

		GetFinishedSurveys(ctx stdcontext.Context, exec DBTransaction, f ResultsFilter, batchSize int, offset int) ([]entity.SurveyStateReport, error)
		GetUserSurveyStates(ctx stdcontext.Context, exec DBTransaction, userGUID uuid.UUID, states []entity.State) ([]entity.SurveyState, error)
		GetUserSurveyState(ctx stdcontext.Context, exec DBTransaction, userGUID uuid.UUID, surveyGUID uuid.UUID, states []entity.State) (entity.SurveyState, error)

		CreateUserSurveyState(ctx stdcontext.Context, exec DBTransaction, state entity.SurveyState) error
		UpdateActiveUserSurveyState(ctx stdcontext.Context, exec DBTransaction, state entity.SurveyState) error
		DeleteUserSurveyState(ctx stdcontext.Context, exec DBTransaction, userGUID uuid.UUID, surveyGUID uuid.UUID) error
	}

	DBTransaction interface {
		Commit() error
		Rollback() error
	}
)
