package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"zombiezen.com/go/postgrestest"

	"git.ykonkov.com/ykonkov/survey-bot/internal/entity"
	"git.ykonkov.com/ykonkov/survey-bot/internal/service"
)

type repisotoryTestSuite struct {
	suite.Suite

	repo service.DBRepo
	db   *sqlx.DB
	f    func()
}

func (suite *repisotoryTestSuite) SetupSuite() {
	dbConnect, f, err := connectToPG()
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.repo = New(dbConnect)
	suite.db = dbConnect
	suite.f = f
}

func (suite *repisotoryTestSuite) AfterTest(suiteName, testName string) {
	// truncate all tables here
	_, err := suite.db.Exec("TRUNCATE TABLE users, surveys, survey_states")
	suite.NoError(err)
}

func (suite *repisotoryTestSuite) TestCreateUser() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	u := entity.User{
		GUID:          uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		UserID:        1,
		ChatID:        1,
		CurrentSurvey: nil,
	}

	err := suite.repo.CreateUser(context.Background(), nil, u)
	suite.NoError(err)

	var got []user
	err = suite.db.Select(&got, "SELECT * FROM users")
	suite.NoError(err)

	suite.equalUsers(
		[]user{
			{
				GUID:          uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
				UserID:        1,
				ChatID:        1,
				CurrentSurvey: nil,
				LastActivity:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				CreatedAt:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		got,
	)
}

func (suite *repisotoryTestSuite) TestCreateUserFailAlreadyExists() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	u := entity.User{
		GUID:          uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		UserID:        1,
		ChatID:        1,
		CurrentSurvey: nil,
	}

	err := suite.repo.CreateUser(context.Background(), nil, u)
	suite.NoError(err)

	err = suite.repo.CreateUser(context.Background(), nil, u)
	suite.Error(err)
	suite.Equal(service.ErrAlreadyExists, err)
}

func (suite *repisotoryTestSuite) TestCreateUserFailAlreadyExistsWithSameUserID() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	u := entity.User{
		GUID:          uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		UserID:        1,
		ChatID:        1,
		CurrentSurvey: nil,
	}

	err := suite.repo.CreateUser(context.Background(), nil, u)
	suite.NoError(err)

	u.GUID = uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD")

	err = suite.repo.CreateUser(context.Background(), nil, u)
	suite.Error(err)
	suite.Equal(service.ErrAlreadyExists, err)
}

func (suite *repisotoryTestSuite) TestUpdateCurrentUserSurvey() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	u := entity.User{
		GUID:          uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		UserID:        1,
		ChatID:        1,
		CurrentSurvey: nil,
	}

	err := suite.repo.CreateUser(context.Background(), nil, u)
	suite.NoError(err)

	_, err = suite.db.Exec("INSERT INTO surveys (guid, id, name, calculations_type, description, questions, created_at, updated_at) VALUES ($1, $2, '', '', '', $3, $4, $5)",
		uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		1,
		"{}",
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	suite.NoError(err)

	now = func() time.Time {
		return time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	err = suite.repo.UpdateUserCurrentSurvey(context.Background(), nil, u.GUID, uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"))
	suite.NoError(err)

	var got []user
	err = suite.db.Select(&got, "SELECT * FROM users")
	suite.NoError(err)

	var surveyGUID = uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD")

	suite.equalUsers(
		[]user{
			{
				GUID:          uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
				UserID:        1,
				ChatID:        1,
				CurrentSurvey: &surveyGUID,
				CreatedAt:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
				LastActivity:  time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		got,
	)
}

func (suite *repisotoryTestSuite) TestUpdateCurrentUserSurveyFailSurveyNotExists() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	u := entity.User{
		GUID:          uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		UserID:        1,
		ChatID:        1,
		CurrentSurvey: nil,
	}

	err := suite.repo.CreateUser(context.Background(), nil, u)
	suite.NoError(err)

	err = suite.repo.UpdateUserCurrentSurvey(context.Background(), nil, u.GUID, uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"))
	suite.Error(err)
}

func (suite *repisotoryTestSuite) TestGetUserByID() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	u := entity.User{
		GUID:         uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		UserID:       1,
		ChatID:       1,
		LastActivity: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	err := suite.repo.CreateUser(context.Background(), nil, u)
	suite.NoError(err)

	got, err := suite.repo.GetUserByID(context.Background(), nil, 1)
	suite.NoError(err)

	suite.Equal(now().Unix(), got.LastActivity.Unix())
	got.LastActivity = now()

	suite.Equal(u, got)
}

func (suite *repisotoryTestSuite) TestGetSurvey() {
	s := survey{
		GUID:      uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		ID:        1,
		Questions: []byte("[]"),
	}

	_, err := suite.db.Exec("INSERT INTO surveys (guid, id, name, calculations_type, description, questions, created_at, updated_at) VALUES ($1, $2, '', '', '', $3, $4, $5)",
		s.GUID,
		s.ID,
		s.Questions,
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	suite.NoError(err)

	got, err := suite.repo.GetSurvey(context.Background(), nil, s.GUID)
	suite.NoError(err)
	suite.Equal(got, entity.Survey{
		GUID:      s.GUID,
		ID:        s.ID,
		Questions: []entity.Question{},
	})
}

func (suite *repisotoryTestSuite) TestGetSurveyFailNotFound() {
	s := survey{
		GUID:      uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		ID:        1,
		Questions: []byte("[]"),
	}

	_, err := suite.db.Exec("INSERT INTO surveys (guid, id, name, calculations_type, description, questions, created_at, updated_at) VALUES ($1, $2, '', '', '', $3, $4, $5)",
		s.GUID,
		s.ID,
		s.Questions,
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	suite.NoError(err)

	_, err = suite.repo.GetSurvey(context.Background(), nil, uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"))
	suite.Error(err)
	suite.Equal(service.ErrNotFound, err)
}

func (suite *repisotoryTestSuite) TestGetSurveysList() {
	s := survey{
		GUID:      uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		ID:        1,
		Questions: []byte(`[{"text":"abc"}]`),
	}

	_, err := suite.db.Exec("INSERT INTO surveys (guid, id, name, calculations_type, description, questions, created_at, updated_at) VALUES ($1, $2, '', '', '', $3, $4, $5)",
		s.GUID,
		s.ID,
		s.Questions,
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	suite.NoError(err)

	s = survey{
		GUID:      uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADF"),
		ID:        1,
		Questions: []byte(`[{"text":"abcd"}]`),
	}

	_, err = suite.db.Exec("INSERT INTO surveys (guid, id, name, calculations_type, description, questions, created_at, updated_at) VALUES ($1, $2, '', '', '', $3, $4, $5)",
		s.GUID,
		s.ID,
		s.Questions,
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	suite.NoError(err)

	got, err := suite.repo.GetSurveysList(context.Background(), nil)
	suite.NoError(err)
	suite.Equal([]entity.Survey{
		{
			GUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
			ID:   1,
			Questions: []entity.Question{
				{
					Text: "abc",
				},
			},
		},
		{
			GUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADF"),
			ID:   1,
			Questions: []entity.Question{
				{
					Text: "abcd",
				},
			},
		},
	}, got)
}

func (suite *repisotoryTestSuite) TestGetSurveyByID() {
	s := survey{
		GUID:      uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		ID:        1,
		Questions: []byte(`[{"text":"abc"}]`),
	}

	_, err := suite.db.Exec("INSERT INTO surveys (guid, id, name, calculations_type, description, questions, created_at, updated_at) VALUES ($1, $2, '', '', '', $3, $4, $5)",
		s.GUID,
		s.ID,
		s.Questions,
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	suite.NoError(err)

	got, err := suite.repo.GetSurveyByID(context.Background(), nil, 1)
	suite.NoError(err)
	suite.Equal(entity.Survey{
		GUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		ID:   1,
		Questions: []entity.Question{
			{
				Text: "abc",
			},
		},
	}, got)
}

func (suite *repisotoryTestSuite) TestGetSurveyByIDFailNotFound() {
	s := survey{
		GUID:      uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		ID:        1,
		Questions: []byte(`[{"text":"abc"}]`),
	}

	_, err := suite.db.Exec("INSERT INTO surveys (guid, id, name, calculations_type, description, questions, created_at, updated_at) VALUES ($1, $2, '', '', '', $3, $4, $5)",
		s.GUID,
		s.ID,
		s.Questions,
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	suite.NoError(err)

	_, err = suite.repo.GetSurveyByID(context.Background(), nil, 2)
	suite.Error(err)
	suite.Equal(service.ErrNotFound, err)
}

func (suite *repisotoryTestSuite) TestCreateUserSurveyState() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	var u = user{
		GUID:          uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		UserID:        1,
		ChatID:        1,
		CurrentSurvey: nil,
		CreatedAt:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		LastActivity:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err := suite.db.Exec("INSERT INTO users (guid, user_id, chat_id, current_survey, created_at, updated_at, last_activity) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		u.GUID,
		u.UserID,
		u.ChatID,
		u.CurrentSurvey,
		u.CreatedAt,
		u.UpdatedAt,
		u.LastActivity,
	)
	suite.NoError(err)

	var s = survey{
		GUID:             uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		Name:             "abc",
		ID:               1,
		CalculationsType: "abc",
		Questions:        []byte(`[{"text":"abc"}]`),
		CreatedAt:        time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err = suite.db.Exec("INSERT INTO surveys (guid, id, name, description, questions, calculations_type, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		s.GUID,
		s.ID,
		s.Name,
		"Survey description", // Provide a non-null value for the description column
		s.Questions,
		s.CalculationsType,
		s.CreatedAt,
		s.UpdatedAt,
	)
	suite.NoError(err)

	err = suite.repo.CreateUserSurveyState(context.Background(), nil, entity.SurveyState{
		State:      entity.ActiveState,
		UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		Answers: []entity.Answer{
			{
				Type: entity.AnswerTypeSegment,
				Data: []int{1},
			},
		},
	})
	suite.NoError(err)

	var got []surveyState
	err = suite.db.Select(&got, "SELECT * FROM survey_states")
	suite.NoError(err)

	suite.equalSurveyStates(
		[]surveyState{
			{
				State:      entity.ActiveState,
				UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
				SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
				Answers:    []byte(`[{"data": [1], "type": "segment"}]`),
				CreatedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		got,
	)
}

func (suite *repisotoryTestSuite) TestCreateUserSurveyStateWithResults() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	var u = user{
		GUID:          uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		UserID:        1,
		ChatID:        1,
		CurrentSurvey: nil,
		CreatedAt:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		LastActivity:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err := suite.db.Exec("INSERT INTO users (guid, user_id, chat_id, current_survey, created_at, updated_at, last_activity) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		u.GUID,
		u.UserID,
		u.ChatID,
		u.CurrentSurvey,
		u.CreatedAt,
		u.UpdatedAt,
		u.LastActivity,
	)
	suite.NoError(err)

	var s = survey{
		GUID:             uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		Name:             "abc",
		ID:               1,
		CalculationsType: "abc",
		Questions:        []byte(`[{"text":"abc"}]`),
		CreatedAt:        time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err = suite.db.Exec("INSERT INTO surveys (guid, id, name, description, questions, calculations_type, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		s.GUID,
		s.ID,
		s.Name,
		"Survey description", // Provide a non-null value for the description column
		s.Questions,
		s.CalculationsType,
		s.CreatedAt,
		s.UpdatedAt,
	)
	suite.NoError(err)

	err = suite.repo.CreateUserSurveyState(context.Background(), nil, entity.SurveyState{
		State:      entity.ActiveState,
		UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		Answers: []entity.Answer{
			{
				Type: entity.AnswerTypeSegment,
				Data: []int{1},
			},
		},
		Results: &entity.Results{
			Text: "abc",
			Metadata: entity.ResultsMetadata{
				Raw: map[string]interface{}{
					"a": "de",
					"f": 10,
				},
			},
		},
	})
	suite.NoError(err)

	var got []surveyState
	err = suite.db.Select(&got, "SELECT * FROM survey_states")
	suite.NoError(err)

	results := []byte(`{"Text": "abc", "Metadata": {"Raw": {"a": "de", "f": 10}}}`)
	suite.equalSurveyStates(
		[]surveyState{
			{
				State:      entity.ActiveState,
				UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
				SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
				Answers:    []byte(`[{"data": [1], "type": "segment"}]`),
				Results:    &results,
				CreatedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		got,
	)
}

func (suite *repisotoryTestSuite) TestGetUserSurveyStateWithResults() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	var u = user{
		GUID:          uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		UserID:        1,
		ChatID:        1,
		CurrentSurvey: nil,
		CreatedAt:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		LastActivity:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err := suite.db.Exec("INSERT INTO users (guid, user_id, chat_id, current_survey, created_at, updated_at, last_activity) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		u.GUID,
		u.UserID,
		u.ChatID,
		u.CurrentSurvey,
		u.CreatedAt,
		u.UpdatedAt,
		u.LastActivity,
	)
	suite.NoError(err)

	var s = survey{
		GUID:      uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		ID:        1,
		Questions: []byte(`[{"text":"abc"}]`),
		CreatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err = suite.db.Exec("INSERT INTO surveys (guid, id, name, calculations_type, description, questions, created_at, updated_at) VALUES ($1, $2, '', '', '', $3, $4, $5)",
		s.GUID,
		s.ID,
		s.Questions,
		s.CreatedAt,
		s.UpdatedAt,
	)
	suite.NoError(err)

	results := []byte(`{"Text": "abc", "Metadata": {"Raw": {"a": "de", "f": 10}}}`)
	var ss = surveyState{
		State:      entity.ActiveState,
		UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		Answers:    []byte(`[{"data": [1], "type": "segment"}]`),
		Results:    &results,
		CreatedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err = suite.db.Exec("INSERT INTO survey_states (state, user_guid, survey_guid, answers, results, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		ss.State,
		ss.UserGUID,
		ss.SurveyGUID,
		ss.Answers,
		ss.Results,
		ss.CreatedAt,
		ss.UpdatedAt,
	)
	suite.NoError(err)

	_, err = suite.db.Exec("INSERT INTO survey_states (state, user_guid, survey_guid, answers, results, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		entity.FinishedState,
		ss.UserGUID,
		ss.SurveyGUID,
		ss.Answers,
		ss.Results,
		ss.CreatedAt,
		ss.UpdatedAt,
	)
	suite.NoError(err)

	got, err := suite.repo.GetUserSurveyStates(context.Background(), nil, ss.UserGUID, []entity.State{entity.ActiveState})
	suite.NoError(err)

	suite.T().Log(got)

	suite.Equal([]entity.SurveyState{
		{
			State:      entity.ActiveState,
			UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
			SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
			Answers: []entity.Answer{
				{
					Type: entity.AnswerTypeSegment,
					Data: []int{1},
				},
			},
			Results: &entity.Results{
				Text: "abc",
				Metadata: entity.ResultsMetadata{
					Raw: map[string]interface{}{
						"a": "de",
						"f": 10.0,
					},
				},
			},
		},
	}, got)
}

func (suite *repisotoryTestSuite) TestGetUserSurveyStates() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	var u = user{
		GUID:          uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		UserID:        1,
		ChatID:        1,
		CurrentSurvey: nil,
		CreatedAt:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		LastActivity:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err := suite.db.Exec("INSERT INTO users (guid, user_id, chat_id, current_survey, created_at, updated_at, last_activity) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		u.GUID,
		u.UserID,
		u.ChatID,
		u.CurrentSurvey,
		u.CreatedAt,
		u.UpdatedAt,
		u.LastActivity,
	)
	suite.NoError(err)

	var s = survey{
		GUID:      uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		ID:        1,
		Questions: []byte(`[{"text":"abc"}]`),
		CreatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err = suite.db.Exec("INSERT INTO surveys (guid, id, name, calculations_type, description, questions, created_at, updated_at) VALUES ($1, $2, '', '', '', $3, $4, $5)",
		s.GUID,
		s.ID,
		s.Questions,
		s.CreatedAt,
		s.UpdatedAt,
	)
	suite.NoError(err)

	var ss = surveyState{
		State:      entity.ActiveState,
		UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		Answers:    []byte(`[{"data": [1], "type": "segment"}]`),
		CreatedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err = suite.db.Exec("INSERT INTO survey_states (state, user_guid, survey_guid, answers, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		ss.State,
		ss.UserGUID,
		ss.SurveyGUID,
		ss.Answers,
		ss.CreatedAt,
		ss.UpdatedAt,
	)
	suite.NoError(err)

	_, err = suite.db.Exec("INSERT INTO survey_states (state, user_guid, survey_guid, answers, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		entity.FinishedState,
		ss.UserGUID,
		ss.SurveyGUID,
		ss.Answers,
		ss.CreatedAt,
		ss.UpdatedAt,
	)
	suite.NoError(err)

	got, err := suite.repo.GetUserSurveyStates(context.Background(), nil, ss.UserGUID, []entity.State{entity.ActiveState})
	suite.NoError(err)
	suite.Equal([]entity.SurveyState{
		{
			State:      entity.ActiveState,
			UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
			SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
			Answers: []entity.Answer{
				{
					Type: entity.AnswerTypeSegment,
					Data: []int{1},
				},
			},
		},
	}, got)
}

func (suite *repisotoryTestSuite) TestCreateUserSurveyStateFailNotFound() {
	err := suite.repo.CreateUserSurveyState(context.Background(), nil, entity.SurveyState{
		State:      entity.ActiveState,
		UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		Answers: []entity.Answer{
			{
				Type: entity.AnswerTypeSegment,
				Data: []int{1},
			},
		},
	})
	suite.Error(err)
}

func (suite *repisotoryTestSuite) TestGetUserSurveyState() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	var u = user{
		GUID:          uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		UserID:        1,
		ChatID:        1,
		CurrentSurvey: nil,
		CreatedAt:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		LastActivity:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err := suite.db.Exec("INSERT INTO users (guid, user_id, chat_id, current_survey, created_at, updated_at, last_activity) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		u.GUID,
		u.UserID,
		u.ChatID,
		u.CurrentSurvey,
		u.CreatedAt,
		u.UpdatedAt,
		u.LastActivity,
	)
	suite.NoError(err)

	var s = survey{
		GUID:      uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		ID:        1,
		Questions: []byte(`[{"text":"abc"}]`),
		CreatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err = suite.db.Exec("INSERT INTO surveys (guid, id, name, calculations_type, description, questions, created_at, updated_at) VALUES ($1, $2, '', '', '', $3, $4, $5)",
		s.GUID,
		s.ID,
		s.Questions,
		s.CreatedAt,
		s.UpdatedAt,
	)
	suite.NoError(err)

	var ss = surveyState{
		State:      entity.ActiveState,
		UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		Answers:    []byte(`[{"data": [1], "type": "segment"}]`),
		CreatedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err = suite.db.Exec("INSERT INTO survey_states (state, user_guid, survey_guid, answers, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		ss.State,
		ss.UserGUID,
		ss.SurveyGUID,
		ss.Answers,
		ss.CreatedAt,
		ss.UpdatedAt,
	)
	suite.NoError(err)

	got, err := suite.repo.GetUserSurveyState(context.Background(), nil, ss.UserGUID, ss.SurveyGUID, []entity.State{entity.ActiveState})
	suite.NoError(err)
	suite.Equal(entity.SurveyState{
		State:      entity.ActiveState,
		UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		Answers: []entity.Answer{
			{
				Type: entity.AnswerTypeSegment,
				Data: []int{1},
			},
		},
	}, got)
}

func (suite *repisotoryTestSuite) TestGetUserSurveyStateFailNotFound() {
	_, err := suite.repo.GetUserSurveyState(
		context.Background(),
		nil,
		uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		[]entity.State{entity.ActiveState},
	)
	suite.Error(err)
	suite.Equal(service.ErrNotFound, err)
}

func (suite *repisotoryTestSuite) TestUpdateUserSurveyState() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	var u = user{
		GUID:          uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		UserID:        1,
		ChatID:        1,
		CurrentSurvey: nil,
		CreatedAt:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		LastActivity:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err := suite.db.Exec("INSERT INTO users (guid, user_id, chat_id, current_survey, created_at, updated_at, last_activity) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		u.GUID,
		u.UserID,
		u.ChatID,
		u.CurrentSurvey,
		u.CreatedAt,
		u.UpdatedAt,
		u.LastActivity,
	)
	suite.NoError(err)

	var s = survey{
		GUID:      uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		ID:        1,
		Questions: []byte(`[{"text":"abc"}]`),
		CreatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err = suite.db.Exec("INSERT INTO surveys (guid, id, name, calculations_type, description, questions, created_at, updated_at) VALUES ($1, $2, '', '', '', $3, $4, $5)",
		s.GUID,
		s.ID,
		s.Questions,
		s.CreatedAt,
		s.UpdatedAt,
	)
	suite.NoError(err)

	var ss = surveyState{
		State:      entity.ActiveState,
		UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		Answers:    []byte(`[{"data": [1], "type": "segment"}]`),
		CreatedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	_, err = suite.db.NamedExec("INSERT INTO survey_states (state, user_guid, survey_guid, answers, created_at, updated_at) VALUES (:state, :user_guid, :survey_guid, :answers, :created_at, :updated_at)", ss)
	suite.NoError(err)

	err = suite.repo.UpdateActiveUserSurveyState(context.Background(), nil, entity.SurveyState{
		State:      entity.ActiveState,
		UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		Answers: []entity.Answer{
			{
				Type: entity.AnswerTypeSegment,
				Data: []int{2},
			},
		},
	})
	suite.NoError(err)

	var got []surveyState
	err = suite.db.Select(&got, "SELECT * FROM survey_states")
	suite.NoError(err)

	suite.equalSurveyStates(
		[]surveyState{
			{
				State:      entity.ActiveState,
				UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
				SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
				Answers:    []byte(`[{"data": [2], "type": "segment"}]`),
				CreatedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		got,
	)
}

func (suite *repisotoryTestSuite) TestUpdateUserSurveyStateWithResults() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	var u = user{
		GUID:          uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		UserID:        1,
		ChatID:        1,
		CurrentSurvey: nil,
		CreatedAt:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err := suite.db.Exec("INSERT INTO users (guid, user_id, chat_id, current_survey, created_at, updated_at, last_activity) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		u.GUID,
		u.UserID,
		u.ChatID,
		u.CurrentSurvey,
		u.CreatedAt,
		u.UpdatedAt,
		u.LastActivity,
	)
	suite.NoError(err)

	var s = survey{
		GUID:      uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		ID:        1,
		Questions: []byte(`[{"text":"abc"}]`),
		CreatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err = suite.db.Exec("INSERT INTO surveys (guid, id, name, calculations_type, description, questions, created_at, updated_at) VALUES ($1, $2, '', '', '', $3, $4, $5)",
		s.GUID,
		s.ID,
		s.Questions,
		s.CreatedAt,
		s.UpdatedAt,
	)
	suite.NoError(err)

	var ss = surveyState{
		State:      entity.ActiveState,
		UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		Answers:    []byte(`[{"data": [1], "type": "segment"}]`),
		CreatedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	_, err = suite.db.NamedExec("INSERT INTO survey_states (state, user_guid, survey_guid, answers, created_at, updated_at) VALUES (:state, :user_guid, :survey_guid, :answers, :created_at, :updated_at)", ss)
	suite.NoError(err)

	err = suite.repo.UpdateActiveUserSurveyState(context.Background(), nil, entity.SurveyState{
		State:      entity.ActiveState,
		UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		Answers: []entity.Answer{
			{
				Type: entity.AnswerTypeSegment,
				Data: []int{2},
			},
		},
		Results: &entity.Results{
			Text: "abc",
			Metadata: entity.ResultsMetadata{
				Raw: map[string]interface{}{
					"a": "de",
					"f": 10,
				},
			},
		},
	})
	suite.NoError(err)

	var got []surveyState
	err = suite.db.Select(&got, "SELECT * FROM survey_states")
	suite.NoError(err)

	results := []byte(`{"Text": "abc", "Metadata": {"Raw": {"a": "de", "f": 10}}}`)
	suite.equalSurveyStates(
		[]surveyState{
			{
				State:      entity.ActiveState,
				UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
				SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
				Answers:    []byte(`[{"data": [2], "type": "segment"}]`),
				Results:    &results,
				CreatedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		got,
	)
}

func (suite *repisotoryTestSuite) TestCreateSurvey() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	err := suite.repo.CreateSurvey(context.Background(), nil, entity.Survey{
		GUID:             uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		CalculationsType: "test_1",
		ID:               1,
		Name:             "abc",
		Questions:        []entity.Question{},
	})
	suite.NoError(err)

	var got []survey
	err = suite.db.Select(&got, "SELECT * FROM surveys")
	suite.NoError(err)

	suite.equalSurveys(
		[]survey{
			{
				GUID:             uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
				ID:               1,
				Name:             "abc",
				CalculationsType: "test_1",
				Questions:        []byte("[]"),
				CreatedAt:        time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:        time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		got,
	)
}

func (suite *repisotoryTestSuite) TestDeleteSurvey() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	err := suite.repo.CreateSurvey(context.Background(), nil, entity.Survey{
		GUID:      uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		ID:        1,
		Questions: []entity.Question{},
	})
	suite.NoError(err)

	err = suite.repo.DeleteSurvey(context.Background(), nil, uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"))
	suite.NoError(err)

	var got []survey
	err = suite.db.Select(&got, "SELECT * FROM surveys")
	suite.NoError(err)

	deletedAt := now()

	suite.equalSurveys(
		[]survey{
			{
				GUID:      uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
				ID:        1,
				Questions: []byte("[]"),
				CreatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				DeletedAt: &deletedAt,
			},
		},
		got,
	)
}

func (suite *repisotoryTestSuite) TestGetUserByGUID() {
	u := user{
		GUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		UserID: 1,
		ChatID: 1,
	}

	_, err := suite.db.Exec("INSERT INTO users (guid, user_id, chat_id, current_survey, created_at, updated_at, last_activity) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		u.GUID,
		u.UserID,
		u.ChatID,
		u.CurrentSurvey,
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	suite.NoError(err)

	got, err := suite.repo.GetUserByGUID(context.Background(), nil, u.GUID)
	suite.NoError(err)

	suite.Equal(got.LastActivity.Unix(), time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).Unix())
	got.LastActivity = time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	suite.Equal(got, entity.User{
		GUID:          u.GUID,
		UserID:        u.UserID,
		ChatID:        u.ChatID,
		CurrentSurvey: nil,
		LastActivity:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		Nickname:      "unknown",
	})
}

func (suite *repisotoryTestSuite) TestGetUserByGUIDFailNotFound() {
	_, err := suite.repo.GetUserByGUID(context.Background(), nil, uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"))
	suite.Error(err)
}

func (suite *repisotoryTestSuite) TestDeleteUserSurveyState() {
	// insert user
	u := user{
		GUID:          uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		UserID:        1,
		ChatID:        1,
		CurrentSurvey: nil,
		LastActivity:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	_, err := suite.db.Exec("INSERT INTO users (guid, user_id, chat_id, current_survey, created_at, updated_at, last_activity) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		u.GUID,
		u.UserID,
		u.ChatID,
		u.CurrentSurvey,
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		u.LastActivity,
	)
	suite.NoError(err)

	// insert survey
	s := survey{
		GUID:      uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADB"),
		ID:        1,
		Questions: []byte(`[]`),
	}
	_, err = suite.db.Exec("INSERT INTO surveys (guid, id, name, calculations_type, description, questions, created_at, updated_at) VALUES ($1, $2, '', 'test', 'Survey description', $3, $4, $5)",
		s.GUID,
		s.ID,
		s.Questions,
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	suite.NoError(err)

	// insert survey state
	ss := surveyState{
		State:      entity.ActiveState,
		UserGUID:   u.GUID,
		SurveyGUID: s.GUID,
		Answers:    []byte(`[]`),
	}
	_, err = suite.db.NamedExec("INSERT INTO survey_states (state, user_guid, survey_guid, answers, created_at, updated_at) VALUES (:state, :user_guid, :survey_guid, :answers, :created_at, :updated_at)", ss)
	suite.NoError(err)

	// delete survey state
	err = suite.repo.DeleteUserSurveyState(context.Background(), nil, u.GUID, s.GUID)
	suite.NoError(err)

	// asert survey state row deleted
	var got []surveyState
	err = suite.db.Select(&got, "SELECT * FROM survey_states")
	suite.NoError(err)
	suite.Equal([]surveyState(nil), got)
}

func (suite *repisotoryTestSuite) TestGetSurveysListEmpty() {
	got, err := suite.repo.GetSurveysList(context.Background(), nil)
	suite.NoError(err)
	suite.Equal([]entity.Survey(nil), got)
}

func (suite *repisotoryTestSuite) TestGetFinishedSurveysWithFilters() {
	now = func() time.Time {
		return time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	users := []user{
		{
			GUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
			UserID: 1,
		},
		{
			GUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADF"),
			UserID: 2,
		},
		{
			GUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B113ADF"),
			UserID: 3,
		},
	}
	for _, u := range users {
		expUser := u.Export()
		err := suite.repo.CreateUser(context.Background(), nil, expUser)
		suite.NoError(err)
	}

	survey := survey{
		GUID:      uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		Questions: []byte(`[{"text":"abc"}]`),
		Name:      "abc",
	}
	expSurvey, err := survey.Export()
	suite.NoError(err)
	err = suite.repo.CreateSurvey(context.Background(), nil, expSurvey)
	suite.NoError(err)

	results := []byte(`{"Text": "abc", "Metadata": {"Raw": {"a": "de", "f": 10}}}`)

	states := []surveyState{
		{
			State:      entity.ActiveState,
			UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
			SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
			Answers:    []byte(`[{"data": [1], "type": "segment"}]`),
		},
		{
			State:      entity.FinishedState,
			UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADF"),
			SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
			Answers:    []byte(`[{"data": [1], "type": "segment"}]`),
			Results:    &results,
		},
		{
			State:      entity.FinishedState,
			UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B113ADF"),
			SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
			Answers:    []byte(`[{"data": [1], "type": "segment"}]`),
			Results:    &results,
		},
	}

	now = func() time.Time {
		return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	expState, err := states[0].Export()
	suite.NoError(err)
	err = suite.repo.CreateUserSurveyState(context.Background(), nil, expState)
	suite.NoError(err)

	now = func() time.Time {
		return time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	expState, err = states[1].Export()
	suite.NoError(err)
	err = suite.repo.CreateUserSurveyState(context.Background(), nil, expState)
	suite.NoError(err)

	now = func() time.Time {
		return time.Date(2017, 2, 1, 0, 0, 0, 0, time.UTC)
	}

	expState, err = states[2].Export()
	suite.NoError(err)
	err = suite.repo.CreateUserSurveyState(context.Background(), nil, expState)
	suite.NoError(err)

	from := time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC)

	got, err := suite.repo.GetFinishedSurveys(context.Background(), nil, service.ResultsFilter{
		From: &from,
		To:   &to,
	}, 3, 0)
	suite.NoError(err)
	suite.equalSurveyStateReports([]entity.SurveyStateReport{
		{
			SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
			UserGUID:   "ae2b602c-f255-47e5-b661-a3f17b113adf",
			UserID:     3,
			SurveyName: "abc",
			StartedAt:  time.Date(2017, 2, 1, 0, 0, 0, 0, time.UTC),
			FinishedAt: time.Date(2017, 2, 1, 0, 0, 0, 0, time.UTC),
			Answers: []entity.Answer{
				{
					Type: entity.AnswerTypeSegment,
					Data: []int{1},
				},
			},
			Results: &entity.Results{
				Text: "abc",
				Metadata: entity.ResultsMetadata{
					Raw: map[string]any{
						"a": "de",
						"f": 10.0,
					},
				},
			},
		},
		{
			SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
			UserGUID:   "ae2b602c-f255-47e5-b661-a3f17b163adf",
			UserID:     2,
			SurveyName: "abc",
			StartedAt:  time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
			FinishedAt: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
			Answers: []entity.Answer{
				{
					Type: entity.AnswerTypeSegment,
					Data: []int{1},
				},
			},
			Results: &entity.Results{
				Text: "abc",
				Metadata: entity.ResultsMetadata{
					Raw: map[string]any{
						"a": "de",
						"f": 10.0,
					},
				},
			},
		},
	}, got)
}

func (suite *repisotoryTestSuite) TestGetFinishedSurveys() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	users := []user{
		{
			GUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
			UserID: 1,
		},
		{
			GUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADF"),
			UserID: 2,
		},
		{
			GUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B113ADF"),
			UserID: 3,
		},
	}
	for _, u := range users {
		expUser := u.Export()
		err := suite.repo.CreateUser(context.Background(), nil, expUser)
		suite.NoError(err)
	}

	survey := survey{
		GUID:      uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		Questions: []byte(`[{"text":"abc"}]`),
		Name:      "abc",
	}
	expSurvey, err := survey.Export()
	suite.NoError(err)
	err = suite.repo.CreateSurvey(context.Background(), nil, expSurvey)
	suite.NoError(err)

	results := []byte(`{"Text": "abc", "Metadata": {"Raw": {"a": "de", "f": 10}}}`)

	states := []surveyState{
		{
			State:      entity.ActiveState,
			UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
			SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
			Answers:    []byte(`[{"data": [1], "type": "segment"}]`),
		},
		{
			State:      entity.FinishedState,
			UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADF"),
			SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
			Answers:    []byte(`[{"data": [1], "type": "segment"}]`),
		},
		{
			State:      entity.FinishedState,
			UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B113ADF"),
			SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
			Answers:    []byte(`[{"data": [1], "type": "segment"}]`),
			Results:    &results,
		},
	}
	for _, state := range states {
		expState, err := state.Export()
		suite.NoError(err)
		err = suite.repo.CreateUserSurveyState(context.Background(), nil, expState)
		suite.NoError(err)
	}

	got, err := suite.repo.GetFinishedSurveys(context.Background(), nil, service.ResultsFilter{}, 2, 1)
	suite.NoError(err)
	suite.equalSurveyStateReports([]entity.SurveyStateReport{
		{
			SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
			UserGUID:   "ae2b602c-f255-47e5-b661-a3f17b113adf",
			UserID:     3,
			SurveyName: "abc",
			StartedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			FinishedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			Answers: []entity.Answer{
				{
					Type: entity.AnswerTypeSegment,
					Data: []int{1},
				},
			},
			Results: &entity.Results{
				Text: "abc",
				Metadata: entity.ResultsMetadata{
					Raw: map[string]interface{}{
						"a": "de",
						"f": 10.0,
					},
				},
			},
		},
	}, got)
}

func (suite *repisotoryTestSuite) TestUpdateSurvey() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	// insert survey
	s := survey{
		GUID:             uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		ID:               1,
		Name:             "Initial Survey",
		Questions:        []byte(`[{"text":"Question 1"}]`),
		CalculationsType: "type1",
	}
	_, err := suite.db.Exec("INSERT INTO surveys (guid, name, id, questions, calculations_type, description, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, '', $6, $7)",
		s.GUID,
		s.Name,
		s.ID,
		s.Questions,
		s.CalculationsType,
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	suite.NoError(err)

	var updatedSurvey entity.Survey
	updatedSurvey.GUID = s.GUID
	updatedSurvey.Name = "Updated Survey"
	updatedSurvey.Questions = []entity.Question{
		{
			Text: "Updated Question 1",
		},
	}
	updatedSurvey.CalculationsType = "type2"

	err = suite.repo.UpdateSurvey(context.Background(), nil, updatedSurvey)
	suite.NoError(err)

	// get survey
	var got survey
	err = suite.db.Get(&got, "SELECT * FROM surveys WHERE guid = $1", s.GUID)
	suite.NoError(err)

	suite.T().Log(string(got.Questions))

	// assert
	suite.equalSurvey(survey{
		GUID:             uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		ID:               1,
		Name:             "Updated Survey",
		Questions:        []byte(`[{"text": "Updated Question 1", "answer_type": "", "answers_text": null, "possible_answers": null}]`),
		CalculationsType: "type2",
		CreatedAt:        time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:        now(),
	}, got)
}

func (suite *repisotoryTestSuite) TestBeginTx() {
	tx, err := suite.repo.BeginTx(context.Background())
	suite.NoError(err)
	suite.NotNil(tx)
}

func (suite *repisotoryTestSuite) TestGetCompletedSurveys() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	var u = user{
		GUID:          uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		UserID:        1,
		ChatID:        1,
		CurrentSurvey: nil,
		CreatedAt:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		LastActivity:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err := suite.db.Exec("INSERT INTO users (guid, user_id, chat_id, current_survey, created_at, updated_at, last_activity) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		u.GUID,
		u.UserID,
		u.ChatID,
		u.CurrentSurvey,
		u.CreatedAt,
		u.UpdatedAt,
		u.LastActivity,
	)
	suite.NoError(err)

	var s = survey{
		GUID:             uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		Name:             "Survey 1",
		ID:               1,
		CalculationsType: "type1",
		Questions:        []byte(`[{"text":"Question 1"}]`),
		CreatedAt:        time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err = suite.db.Exec("INSERT INTO surveys (guid, id, name, questions, calculations_type, description, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, '', $6, $7)",
		s.GUID,
		s.ID,
		s.Name,
		s.Questions,
		s.CalculationsType,
		s.CreatedAt,
		s.UpdatedAt,
	)
	suite.NoError(err)

	var ss = surveyState{
		State:      entity.FinishedState,
		UserGUID:   uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
		SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
		Answers:    []byte(`[{"data": [1], "type": "segment"}]`),
		Results:    nil,
		CreatedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err = suite.db.Exec("INSERT INTO survey_states (state, user_guid, survey_guid, answers, results, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		ss.State,
		ss.UserGUID,
		ss.SurveyGUID,
		ss.Answers,
		ss.Results,
		ss.CreatedAt,
		ss.UpdatedAt,
	)
	suite.NoError(err)

	got, err := suite.repo.GetCompletedSurveys(context.Background(), nil, u.GUID)
	suite.NoError(err)

	suite.equalSurveyStateReports([]entity.SurveyStateReport{
		{
			SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
			SurveyName: "Survey 1",
			StartedAt:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			FinishedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			UserGUID:   "ae2b602c-f255-47e5-b661-a3f17b163adc",
			UserID:     1,
			Answers: []entity.Answer{
				{
					Type: entity.AnswerTypeSegment,
					Data: []int{1},
				},
			},
			Results: nil,
		},
	}, got)
}

func TestRepisotoryTestSuite(t *testing.T) {
	suite.Run(t, new(repisotoryTestSuite))
}

func (suite *repisotoryTestSuite) TestGetUsersList() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	users := []user{
		{
			GUID:         uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
			UserID:       1,
			Nickname:     "user1",
			CreatedAt:    now(),
			UpdatedAt:    now(),
			LastActivity: now(),
		},
		{
			GUID:         uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADF"),
			UserID:       2,
			Nickname:     "user2",
			CreatedAt:    now(),
			UpdatedAt:    now(),
			LastActivity: now(),
		},
	}

	for _, u := range users {
		_, err := suite.db.Exec("INSERT INTO users (guid, user_id, chat_id, nickname, created_at, updated_at, last_activity) VALUES ($1, $2, 0, $3, $4, $5, $6)",
			u.GUID, u.UserID, u.Nickname, u.CreatedAt, u.UpdatedAt, u.LastActivity)
		suite.NoError(err)
	}

	surveys := []survey{
		{
			GUID:             uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
			ID:               1,
			Name:             "Survey 1",
			Questions:        []byte(`[{"text":"Question 1"}]`),
			CalculationsType: "type1",
			CreatedAt:        now(),
			UpdatedAt:        now(),
		},
		{
			GUID:             uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADE"),
			ID:               2,
			Name:             "Survey 2",
			Questions:        []byte(`[{"text":"Question 2"}]`),
			CalculationsType: "type2",
			CreatedAt:        now(),
			UpdatedAt:        now(),
		},
	}

	for _, s := range surveys {
		_, err := suite.db.Exec("INSERT INTO surveys (guid, id, name, questions, calculations_type, description, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, '', $6, $7)",
			s.GUID, s.ID, s.Name, s.Questions, s.CalculationsType, s.CreatedAt, s.UpdatedAt)
		suite.NoError(err)
	}

	surveyStates := []surveyState{
		{
			State:      entity.FinishedState,
			UserGUID:   users[0].GUID,
			SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADD"),
			Answers:    []byte(`[{"data": [1], "type": "segment"}]`),
			CreatedAt:  now(),
			UpdatedAt:  now(),
		},
		{
			State:      entity.ActiveState,
			UserGUID:   users[1].GUID,
			SurveyGUID: uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADE"),
			Answers:    []byte(`[{"data": [2], "type": "segment"}]`),
			CreatedAt:  now(),
			UpdatedAt:  now(),
		},
	}

	for _, ss := range surveyStates {
		_, err := suite.db.Exec("INSERT INTO survey_states (state, user_guid, survey_guid, answers, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
			ss.State, ss.UserGUID, ss.SurveyGUID, ss.Answers, ss.CreatedAt, ss.UpdatedAt)
		suite.NoError(err)
	}

	got, err := suite.repo.GetUsersList(context.Background(), nil, 10, 0, "user")
	suite.NoError(err)

	suite.Equal(now().Unix(), got.Users[0].RegisteredAt.Unix())
	suite.Equal(now().Unix(), got.Users[0].LastActivity.Unix())
	suite.Equal(now().Unix(), got.Users[1].RegisteredAt.Unix())
	suite.Equal(now().Unix(), got.Users[1].LastActivity.Unix())

	got.Users[0].RegisteredAt = now()
	got.Users[0].LastActivity = now()
	got.Users[1].RegisteredAt = now()
	got.Users[1].LastActivity = now()

	expected := service.UserListResponse{
		Users: []service.UserReport{
			{
				GUID:              users[0].GUID,
				NickName:          "user1",
				RegisteredAt:      now(),
				LastActivity:      now(),
				CompletedTests:    1,
				AnsweredQuestions: 1,
			},
			{
				GUID:              users[1].GUID,
				NickName:          "user2",
				RegisteredAt:      now(),
				LastActivity:      now(),
				CompletedTests:    0,
				AnsweredQuestions: 1,
			},
		},
		Total: 2,
	}

	suite.Equal(expected, got)
}

func (suite *repisotoryTestSuite) TestGetUsersListEmpty() {
	got, err := suite.repo.GetUsersList(context.Background(), nil, 10, 0, "user")
	suite.NoError(err)
	suite.Equal(service.UserListResponse{Users: []service.UserReport{}, Total: 0}, got)
}

func (suite *repisotoryTestSuite) TestGetUsersListWithEmptySearch() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	users := []user{
		{
			GUID:         uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADC"),
			UserID:       1,
			Nickname:     "user1",
			CreatedAt:    now(),
			UpdatedAt:    now(),
			LastActivity: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			GUID:         uuid.MustParse("AE2B602C-F255-47E5-B661-A3F17B163ADF"),
			UserID:       2,
			Nickname:     "user2",
			CreatedAt:    now(),
			UpdatedAt:    now(),
			LastActivity: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, u := range users {
		_, err := suite.db.Exec("INSERT INTO users (guid, user_id, chat_id, nickname, created_at, updated_at, last_activity) VALUES ($1, $2, 0, $3, $4, $5, $6)",
			u.GUID, u.UserID, u.Nickname, u.CreatedAt, u.UpdatedAt, u.LastActivity)
		suite.NoError(err)
	}

	got, err := suite.repo.GetUsersList(context.Background(), nil, 10, 0, "")
	suite.NoError(err)

	suite.Equal(now().Unix(), got.Users[0].RegisteredAt.Unix())
	suite.Equal(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).Unix(), got.Users[1].LastActivity.Unix())
	suite.Equal(now().Unix(), got.Users[1].RegisteredAt.Unix())
	suite.Equal(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC).Unix(), got.Users[0].LastActivity.Unix())

	got.Users[0].RegisteredAt = now()
	got.Users[0].LastActivity = now()
	got.Users[1].RegisteredAt = now()
	got.Users[1].LastActivity = now()

	expected := service.UserListResponse{
		Users: []service.UserReport{
			{
				GUID:              users[1].GUID,
				NickName:          "user2",
				RegisteredAt:      now(),
				LastActivity:      now(),
				CompletedTests:    0,
				AnsweredQuestions: 0,
			},
			{
				GUID:              users[0].GUID,
				NickName:          "user1",
				RegisteredAt:      now(),
				LastActivity:      now(),
				CompletedTests:    0,
				AnsweredQuestions: 0,
			},
		},
		Total: 2,
	}

	suite.Equal(expected, got)
}

func (suite *repisotoryTestSuite) equalUsers(expected, actual []user) {
	suite.Len(actual, len(expected))
	for i := range expected {
		suite.equalUser(expected[i], actual[i])
	}
}

func (suite *repisotoryTestSuite) equalUser(expected, actual user) {
	// check created_at, updated_at, deleted_at if exists
	suite.Equal(expected.CreatedAt.Unix(), actual.CreatedAt.Unix())
	suite.Equal(expected.UpdatedAt.Unix(), actual.UpdatedAt.Unix())
	suite.Equal(expected.LastActivity.Unix(), actual.LastActivity.Unix())

	// remove created_at, updated_at, last_activity
	expected.CreatedAt = actual.CreatedAt
	expected.UpdatedAt = actual.UpdatedAt
	expected.LastActivity = actual.LastActivity

	suite.Equal(expected, actual)
}

func (suite *repisotoryTestSuite) equalSurveyStates(expected, actual []surveyState) {
	suite.Len(actual, len(expected))
	for i := range expected {
		suite.equalSurveyState(expected[i], actual[i])
	}
}

func (suite *repisotoryTestSuite) equalSurveyState(expected, actual surveyState) {
	// check created_at, updated_at, deleted_at if exists
	suite.Equal(expected.CreatedAt.Unix(), actual.CreatedAt.Unix())
	suite.Equal(expected.UpdatedAt.Unix(), actual.UpdatedAt.Unix())

	// remove created_at, updated_at, deleted_at
	expected.CreatedAt = actual.CreatedAt
	expected.UpdatedAt = actual.UpdatedAt

	suite.Equal(expected, actual)
}

func (suite *repisotoryTestSuite) equalSurveyStateReports(expected, actual []entity.SurveyStateReport) {
	suite.Len(actual, len(expected))
	for i := range expected {
		suite.equalSurveyStateReport(expected[i], actual[i])
	}
}

func (suite *repisotoryTestSuite) equalSurveyStateReport(expected, actual entity.SurveyStateReport) {
	// check created_at, updated_at, deleted_at if exists
	suite.Equal(expected.StartedAt.Unix(), actual.StartedAt.Unix())
	suite.Equal(expected.FinishedAt.Unix(), actual.FinishedAt.Unix())

	// remove created_at, updated_at, deleted_at
	expected.StartedAt = actual.StartedAt
	expected.FinishedAt = actual.FinishedAt

	suite.Equal(expected, actual)
}

func (suite *repisotoryTestSuite) equalSurveys(expected, actual []survey) {
	suite.Len(actual, len(expected))
	for i := range expected {
		suite.equalSurvey(expected[i], actual[i])
	}
}

func (suite *repisotoryTestSuite) equalSurvey(expected, actual survey) {
	// check created_at, updated_at, deleted_at if exists
	suite.Equal(expected.CreatedAt.Unix(), actual.CreatedAt.Unix())
	suite.Equal(expected.UpdatedAt.Unix(), actual.UpdatedAt.Unix())

	switch {
	case expected.DeletedAt == nil && actual.DeletedAt != nil:
		suite.Fail("expected deleted_at to be nil, but got %v", actual.DeletedAt)
	case expected.DeletedAt != nil && actual.DeletedAt == nil:
		suite.Fail("expected deleted_at to be %v, but got nil", expected.DeletedAt)
	case expected.DeletedAt != nil && actual.DeletedAt != nil:
		suite.Equal(expected.DeletedAt.Unix(), actual.DeletedAt.Unix())
	}

	// remove created_at, updated_at, deleted_at
	expected.CreatedAt = actual.CreatedAt
	expected.UpdatedAt = actual.UpdatedAt
	expected.DeletedAt = actual.DeletedAt

	suite.Equal(expected, actual)
}

func connectToPG() (*sqlx.DB, func(), error) {
	ctx := context.Background()
	srv, err := postgrestest.Start(ctx)
	if err != nil {
		return nil, func() {}, fmt.Errorf("Failed to start container: %w", err)
	}

	db, err := srv.NewDatabase(ctx)
	if err != nil {
		srv.Cleanup()
		return nil, func() {}, fmt.Errorf("Failed to create database: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		srv.Cleanup()
		return nil, func() {}, fmt.Errorf("Failed to create driver: %w", err)
	}

	d, err := iofs.New(fs, "sql")
	if err != nil {
		srv.Cleanup()
		return nil, func() {}, fmt.Errorf("Failed to create iofs source: %w", err)
	}

	m, err := migrate.NewWithInstance(
		"file",
		d,
		"postgres",
		driver,
	)
	if err != nil {
		srv.Cleanup()
		return nil, func() {}, fmt.Errorf("Failed to create migrate instance: %w", err)
	}

	err = m.Up()
	if err != nil {
		srv.Cleanup()
		return nil, func() {}, fmt.Errorf("Failed to migrate: %w", err)
	}

	dbx := sqlx.NewDb(db, "postgres")

	return dbx, func() { srv.Cleanup() }, nil
}
