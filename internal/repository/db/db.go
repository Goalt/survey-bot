package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"git.ykonkov.com/ykonkov/survey-bot/internal/entity"
	"git.ykonkov.com/ykonkov/survey-bot/internal/service"
)

var now = time.Now

type dbExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

type repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *repository {
	return &repository{db: db}
}

func (r *repository) BeginTx(ctx context.Context) (service.DBTransaction, error) {
	span := sentry.StartSpan(ctx, "BeginTx")
	defer span.Finish()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return tx, nil
}

func (r *repository) castExec(tx service.DBTransaction) (dbExecutor, error) {
	var exec dbExecutor
	switch tx {
	case nil:
		exec = r.db
	default:
		txSqlx, ok := tx.(*sqlx.Tx)
		if !ok {
			return nil, fmt.Errorf("failed to convert tx to *sqlx.Tx")
		}

		exec = txSqlx
	}

	return exec, nil
}

func (r *repository) GetSurveysList(ctx context.Context, tx service.DBTransaction) ([]entity.Survey, error) {
	span := sentry.StartSpan(ctx, "GetSurveysList")
	defer span.Finish()

	exec, err := r.castExec(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to cast exec: %w", err)
	}

	var models []survey

	query := "SELECT * FROM surveys WHERE deleted_at IS NULL"
	if err := exec.SelectContext(ctx, &models, query); err != nil {
		return nil, fmt.Errorf("failed to exec query: %w", err)
	}

	var surveys []entity.Survey
	for _, model := range models {
		s, err := model.Export()
		if err != nil {
			return nil, fmt.Errorf("failed to export survey: %w", err)
		}

		surveys = append(surveys, s)
	}

	return surveys, nil
}

func (r *repository) CreateSurvey(ctx context.Context, tx service.DBTransaction, s entity.Survey) error {
	span := sentry.StartSpan(ctx, "CreateSurvey")
	defer span.Finish()

	exec, err := r.castExec(tx)
	if err != nil {
		return fmt.Errorf("failed to cast exec: %w", err)
	}

	var model survey
	if err := model.Load(s); err != nil {
		return fmt.Errorf("failed to load survey: %w", err)
	}

	nowTime := now()
	model.CreatedAt = nowTime
	model.UpdatedAt = nowTime

	query := `INSERT INTO surveys (guid, id, name, questions, calculations_type, description, created_at, updated_at)
		VALUES (:guid, :id, :name, :questions, :calculations_type, :description, :created_at, :updated_at)`
	if _, err := exec.NamedExecContext(ctx, query, model); err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}

func (r *repository) DeleteSurvey(ctx context.Context, tx service.DBTransaction, surveyGUID uuid.UUID) error {
	span := sentry.StartSpan(ctx, "DeleteSurvey")
	defer span.Finish()

	exec, err := r.castExec(tx)
	if err != nil {
		return fmt.Errorf("failed to cast exec: %w", err)
	}

	nowTime := now()

	query := `UPDATE surveys SET deleted_at = $1 WHERE guid = $2`
	if _, err := exec.ExecContext(ctx, query, nowTime, surveyGUID); err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}

func (r *repository) GetSurvey(ctx context.Context, tx service.DBTransaction, surveyGUID uuid.UUID) (entity.Survey, error) {
	span := sentry.StartSpan(ctx, "GetSurvey")
	defer span.Finish()

	exec, err := r.castExec(tx)
	if err != nil {
		return entity.Survey{}, fmt.Errorf("failed to cast exec: %w", err)
	}

	var model survey

	query := "SELECT * FROM surveys WHERE guid = $1 AND deleted_at IS NULL"
	if err := exec.GetContext(ctx, &model, query, surveyGUID); err != nil {
		if err == sql.ErrNoRows {
			return entity.Survey{}, service.ErrNotFound
		}

		return entity.Survey{}, fmt.Errorf("failed to exec query: %w", err)
	}

	return model.Export()
}

func (r *repository) GetSurveyByID(ctx context.Context, tx service.DBTransaction, surveyID int64) (entity.Survey, error) {
	span := sentry.StartSpan(ctx, "GetSurveyByID")
	defer span.Finish()

	exec, err := r.castExec(tx)
	if err != nil {
		return entity.Survey{}, fmt.Errorf("failed to cast exec: %w", err)
	}

	var model survey

	query := "SELECT * FROM surveys WHERE id = $1 AND deleted_at IS NULL"
	if err := exec.GetContext(ctx, &model, query, surveyID); err != nil {
		if err == sql.ErrNoRows {
			return entity.Survey{}, service.ErrNotFound
		}

		return entity.Survey{}, fmt.Errorf("failed to exec query: %w", err)
	}

	return model.Export()
}

func (r *repository) GetUserByID(ctx context.Context, tx service.DBTransaction, userID int64) (entity.User, error) {
	span := sentry.StartSpan(ctx, "GetUserByID")
	defer span.Finish()

	exec, err := r.castExec(tx)
	if err != nil {
		return entity.User{}, fmt.Errorf("failed to cast exec: %w", err)
	}

	var model user
	query := "SELECT * FROM users WHERE user_id = $1"
	err = exec.GetContext(ctx, &model, query, userID)
	switch {
	case err != nil && strings.Contains(err.Error(), "sql: no rows in result set"):
		return entity.User{}, service.ErrNotFound
	case err != nil:
		return entity.User{}, fmt.Errorf("failed to exec query: %w", err)
	}

	return model.Export(), nil
}

// CreateUser creates new user in DB and returns ErrAlreadyExists if user already exists
func (r *repository) CreateUser(ctx context.Context, tx service.DBTransaction, u entity.User) error {
	span := sentry.StartSpan(ctx, "CreateUser")
	defer span.Finish()

	exec, err := r.castExec(tx)
	if err != nil {
		return fmt.Errorf("failed to cast exec: %w", err)
	}

	nowTime := now()

	var model user
	model.Load(u)
	model.CreatedAt = nowTime
	model.UpdatedAt = nowTime
	model.LastActivity = nowTime

	query := `INSERT INTO users (guid, user_id, chat_id, nickname, current_survey, created_at, updated_at, last_activity)
        VALUES (:guid, :user_id, :chat_id, :nickname, :current_survey, :created_at, :updated_at, :last_activity)`
	_, err = exec.NamedExecContext(ctx, query, model)
	switch {
	case err != nil && strings.Contains(err.Error(), `pq: duplicate key value violates unique constraint "users_pk"`):
		return service.ErrAlreadyExists
	case err != nil && strings.Contains(err.Error(), `pq: duplicate key value violates unique constraint "user_user_id_key"`):
		return service.ErrAlreadyExists
	case err != nil:
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}

func (r *repository) UpdateUserCurrentSurvey(ctx context.Context, tx service.DBTransaction, userGUID uuid.UUID, surveyGUID uuid.UUID) error {
	span := sentry.StartSpan(ctx, "UpdateUserCurrentSurvey")
	defer span.Finish()

	exec, err := r.castExec(tx)
	if err != nil {
		return fmt.Errorf("failed to cast exec: %w", err)
	}

	nowTime := now()

	query := `UPDATE users SET current_survey = $1, updated_at = $2, last_activity = $3 WHERE guid = $4`
	if _, err := exec.ExecContext(ctx, query, surveyGUID, nowTime, nowTime, userGUID); err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}

// GetUserSurveyState returns user's survey state from DB and returns ErrNotFound if not found
func (r *repository) GetUserSurveyState(ctx context.Context, tx service.DBTransaction, userGUID uuid.UUID, surveyGUID uuid.UUID, states []entity.State) (entity.SurveyState, error) {
	span := sentry.StartSpan(ctx, "GetUserSurveyState")
	defer span.Finish()

	exec, err := r.castExec(tx)
	if err != nil {
		return entity.SurveyState{}, fmt.Errorf("failed to cast exec: %w", err)
	}

	query, args, err := sqlx.In("SELECT * FROM survey_states WHERE user_guid = ? AND survey_guid = ? AND state IN(?)", userGUID, surveyGUID, states)
	if err != nil {
		return entity.SurveyState{}, fmt.Errorf("failed to build query: %w", err)
	}

	query = r.db.Rebind(query)

	var model surveyState
	if err := exec.GetContext(ctx, &model, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return entity.SurveyState{}, service.ErrNotFound
		}

		return entity.SurveyState{}, fmt.Errorf("failed to exec query: %w", err)
	}

	return model.Export()
}

func (r *repository) CreateUserSurveyState(ctx context.Context, tx service.DBTransaction, state entity.SurveyState) error {
	span := sentry.StartSpan(ctx, "CreateUserSurveyState")
	defer span.Finish()

	exec, err := r.castExec(tx)
	if err != nil {
		return fmt.Errorf("failed to cast exec: %w", err)
	}

	var surveyState surveyState
	if err := surveyState.Load(state); err != nil {
		return fmt.Errorf("failed to load survey state: %w", err)
	}

	nowTime := now()
	surveyState.CreatedAt = nowTime
	surveyState.UpdatedAt = nowTime

	query := `INSERT INTO survey_states (user_guid, survey_guid, state, answers, results, created_at, updated_at)
		VALUES (:user_guid, :survey_guid, :state, :answers, :results, :created_at, :updated_at)`
	if _, err := exec.NamedExecContext(ctx, query, surveyState); err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}

func (r *repository) UpdateActiveUserSurveyState(ctx context.Context, tx service.DBTransaction, state entity.SurveyState) error {
	span := sentry.StartSpan(ctx, "UpdateUserSurveyState")
	defer span.Finish()

	exec, err := r.castExec(tx)
	if err != nil {
		return fmt.Errorf("failed to cast exec: %w", err)
	}

	var model surveyState
	if err := model.Load(state); err != nil {
		return fmt.Errorf("failed to load survey state: %w", err)
	}

	nowTime := now()
	model.UpdatedAt = nowTime

	query := `UPDATE survey_states SET state = :state, answers = :answers, results = :results, updated_at = :updated_at
		WHERE user_guid = :user_guid AND survey_guid = :survey_guid AND state = 'active'`
	if _, err := exec.NamedExecContext(ctx, query, model); err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}

func (r *repository) SetUserCurrentSurveyToNil(ctx context.Context, tx service.DBTransaction, userGUID uuid.UUID) error {
	span := sentry.StartSpan(ctx, "SetUserCurrentSurveyToNil")
	defer span.Finish()

	exec, err := r.castExec(tx)
	if err != nil {
		return fmt.Errorf("failed to cast exec: %w", err)
	}

	nowTime := now()

	query := `UPDATE users SET current_survey = NULL, updated_at = $1 WHERE guid = $2`
	if _, err := exec.ExecContext(ctx, query, nowTime, userGUID); err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}

func (r *repository) GetUserSurveyStates(ctx context.Context, tx service.DBTransaction, userGUID uuid.UUID, filterStates []entity.State) ([]entity.SurveyState, error) {
	span := sentry.StartSpan(ctx, "GetUserSurveyStates")
	defer span.Finish()

	exec, err := r.castExec(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to cast exec: %w", err)
	}

	var models []surveyState

	query, args, err := sqlx.In("SELECT * FROM survey_states WHERE user_guid = ? AND state IN(?)", userGUID, filterStates)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	query = r.db.Rebind(query)

	if err := exec.SelectContext(ctx, &models, query, args...); err != nil {
		return nil, fmt.Errorf("failed to exec query: %w", err)
	}

	var states []entity.SurveyState
	for _, model := range models {
		s, err := model.Export()
		if err != nil {
			return nil, fmt.Errorf("failed to export survey state: %w", err)
		}

		states = append(states, s)
	}

	return states, nil
}

func (r *repository) DeleteUserSurveyState(ctx context.Context, tx service.DBTransaction, userGUID uuid.UUID, surveyGUID uuid.UUID) error {
	span := sentry.StartSpan(ctx, "DeleteUserSurveyState")
	defer span.Finish()

	exec, err := r.castExec(tx)
	if err != nil {
		return fmt.Errorf("failed to cast exec: %w", err)
	}

	query := `DELETE FROM survey_states WHERE user_guid = $1 AND survey_guid = $2`
	if _, err := exec.ExecContext(ctx, query, userGUID, surveyGUID); err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}

func (r *repository) GetUserByGUID(ctx context.Context, tx service.DBTransaction, userGUID uuid.UUID) (entity.User, error) {
	span := sentry.StartSpan(ctx, "GetUserByGUID")
	defer span.Finish()

	exec, err := r.castExec(tx)
	if err != nil {
		return entity.User{}, fmt.Errorf("failed to cast exec: %w", err)
	}

	var model user
	query := `SELECT * FROM users WHERE guid = $1`
	if err := exec.GetContext(ctx, &model, query, userGUID); err != nil {
		return entity.User{}, fmt.Errorf("failed to exec query: %w", err)
	}

	return model.Export(), nil
}

func (r *repository) GetFinishedSurveys(ctx context.Context, tx service.DBTransaction, f service.ResultsFilter, batchSize int, offset int) ([]entity.SurveyStateReport, error) {
	span := sentry.StartSpan(ctx, "GetFinishedSurveys")
	defer span.Finish()

	exec, err := r.castExec(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to cast exec: %w", err)
	}

	var models []surveyStateReport

	var (
		args   []any
		filter string
	)

	args = append(args, batchSize, offset)

	switch {
	case f.From == nil && f.To == nil:
		filter = "state = $3"
		args = append(args, entity.FinishedState)
	case f.From == nil && f.To != nil:
		filter = "state = $3 AND updated_at < $4"
		args = append(args, entity.FinishedState, f.To)
	case f.From != nil && f.To == nil:
		filter = "state = $3 AND updated_at >= $5"
		args = append(args, entity.FinishedState, f.From)
	case f.From != nil && f.To != nil:
		filter = "state = $3 AND updated_at >= $4 AND updated_at < $5"
		args = append(args, entity.FinishedState, f.From, f.To)
	}

	query := `
	WITH cte_states AS (
		SELECT answers, 
			user_guid, 
			survey_guid, 
			results, 
			created_at state_created_at, 
			updated_at state_updated_at 
		FROM survey_states WHERE %s ORDER BY updated_at LIMIT $1 OFFSET $2
	),
	cte_users AS (
		SELECT answers,
			user_id,
			user_guid,
			survey_guid,
			results,
			state_created_at,
			state_updated_at
		FROM cte_states LEFT JOIN users ON cte_states.user_guid = users.guid
	),
	cte_users_with_surveys_and_states AS (
		SELECT answers, 
			user_id, 
			user_guid, 
			survey_guid, 
			results, 
			state_created_at created_at, 
			state_updated_at updated_at, 
			"name" survey_name 
		FROM cte_users LEFT JOIN surveys ON cte_users.survey_guid = surveys.guid
	)
	SELECT * FROM cte_users_with_surveys_and_states
	`

	if err := exec.SelectContext(
		ctx,
		&models,
		fmt.Sprintf(query, filter),
		args...,
	); err != nil {
		return nil, fmt.Errorf("failed to exec query: %w", err)
	}

	var states []entity.SurveyStateReport
	for _, model := range models {
		s, err := model.Export()
		if err != nil {
			return nil, fmt.Errorf("failed to export survey state: %w", err)
		}

		states = append(states, s)
	}

	return states, nil
}

func (r *repository) UpdateSurvey(ctx context.Context, tx service.DBTransaction, s entity.Survey) error {
	span := sentry.StartSpan(ctx, "UpdateSurvey")
	defer span.Finish()

	exec, err := r.castExec(tx)
	if err != nil {
		return fmt.Errorf("failed to cast exec: %w", err)
	}

	var model survey
	if err := model.Load(s); err != nil {
		return fmt.Errorf("failed to load survey: %w", err)
	}

	nowTime := now()
	model.UpdatedAt = nowTime

	query := `UPDATE surveys SET name = :name, questions = :questions, calculations_type = :calculations_type, description = :description, updated_at = :updated_at
		WHERE guid = :guid`
	if _, err := exec.NamedExecContext(ctx, query, model); err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}

func (r *repository) GetCompletedSurveys(ctx context.Context, tx service.DBTransaction, userGUID uuid.UUID) ([]entity.SurveyStateReport, error) {
	span := sentry.StartSpan(ctx, "GetCompletedSurveys")
	defer span.Finish()

	exec, err := r.castExec(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to cast exec: %w", err)
	}

	var models []surveyStateReport

	query := `
	SELECT ss.survey_guid, s.name as survey_name, s.description, ss.created_at, ss.updated_at, ss.user_guid, u.user_id, ss.answers, ss.results
	FROM survey_states ss
	JOIN surveys s ON ss.survey_guid = s.guid
	JOIN users u ON ss.user_guid = u.guid
	WHERE ss.user_guid = $1 AND ss.state = $2
	ORDER BY ss.updated_at DESC
	`
	if err := exec.SelectContext(ctx, &models, query, userGUID, entity.FinishedState); err != nil {
		return nil, fmt.Errorf("failed to exec query: %w", err)
	}

	var states []entity.SurveyStateReport
	for _, model := range models {
		s, err := model.Export()
		if err != nil {
			return nil, fmt.Errorf("failed to export survey state: %w", err)
		}

		states = append(states, s)
	}

	return states, nil
}

func (r *repository) UpdateUserLastActivity(ctx context.Context, tx service.DBTransaction, userGUID uuid.UUID) error {
	span := sentry.StartSpan(ctx, "UpdateUserLastActivity")
	defer span.Finish()

	exec, err := r.castExec(tx)
	if err != nil {
		return fmt.Errorf("failed to cast exec: %w", err)
	}

	nowTime := now()

	query := `UPDATE users SET last_activity = $1 WHERE guid = $2`
	if _, err := exec.ExecContext(ctx, query, nowTime, userGUID); err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}

func (r *repository) GetUsersList(ctx context.Context, tx service.DBTransaction, limit, offset int, search string) (service.UserListResponse, error) {
	span := sentry.StartSpan(ctx, "GetUsersList")
	defer span.Finish()

	exec, err := r.castExec(tx)
	if err != nil {
		return service.UserListResponse{}, fmt.Errorf("failed to cast exec: %w", err)
	}

	var users []userListReportResponse
	query := `
	WITH USERS_PART AS (
    SELECT
        *
    FROM
        USERS U
	WHERE U.nickname ILIKE $1
    ORDER BY
        U.last_activity DESC
    LIMIT $2 OFFSET $3
	),
	USERS_PART_SURVEYS_STATES AS (
		SELECT
			UP.guid,
			UP.nickname,
			UP.created_at,
			UP.last_activity,
			SS.answers,
			SS.state
		FROM
			USERS_PART UP
		LEFT JOIN SURVEY_STATES SS ON
			UP.guid = SS.user_guid
	)
	SELECT
		*
	FROM
		USERS_PART_SURVEYS_STATES;
	`
	if err := exec.SelectContext(ctx, &users, query, "%"+search+"%", limit, offset); err != nil {
		return service.UserListResponse{}, fmt.Errorf("failed to exec query: %w", err)
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM users WHERE nickname ILIKE $1`
	if err := exec.GetContext(ctx, &total, countQuery, "%"+search+"%"); err != nil {
		return service.UserListResponse{}, fmt.Errorf("failed to exec count query: %w", err)
	}

	result := make(map[uuid.UUID]service.UserReport)
	for _, u := range users {
		userResult := result[u.GUID]

		if (u.SurveyState != nil) && (*u.SurveyState == entity.FinishedState) {
			userResult.CompletedTests++
		}

		userResult.NickName = u.NickName
		userResult.GUID = u.GUID
		userResult.RegisteredAt = u.CreatedAt
		userResult.LastActivity = u.LastActivity

		if u.Answers != nil {
			var answers []entity.Answer
			if err := json.Unmarshal(*u.Answers, &answers); err != nil {
				return service.UserListResponse{}, fmt.Errorf("failed to unmarshal answers: %w", err)
			}
			userResult.AnsweredQuestions += len(answers)
		}

		result[u.GUID] = userResult
	}

	resultSlice := make([]service.UserReport, 0, len(result))
	for _, v := range result {
		resultSlice = append(resultSlice, v)
	}

	sort.Slice(resultSlice, func(i, j int) bool {
		if resultSlice[i].LastActivity.Equal(resultSlice[j].LastActivity) {
			// If timestamps are equal, sort by GUID for deterministic ordering
			return resultSlice[i].GUID.String() < resultSlice[j].GUID.String()
		}
		return resultSlice[i].LastActivity.After(resultSlice[j].LastActivity)
	})

	return service.UserListResponse{
		Users: resultSlice,
		Total: total,
	}, nil
}
