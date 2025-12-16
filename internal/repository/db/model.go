package db

import (
	"encoding/json"
	"fmt"
	"time"

	"git.ykonkov.com/ykonkov/survey-bot/internal/entity"
	"github.com/google/uuid"
)

type (
	user struct {
		GUID          uuid.UUID  `db:"guid"`
		UserID        int64      `db:"user_id"`
		ChatID        int64      `db:"chat_id"`
		Nickname      string     `db:"nickname"`
		CurrentSurvey *uuid.UUID `db:"current_survey"`
		LastActivity  time.Time  `db:"last_activity"`

		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}

	survey struct {
		GUID             uuid.UUID `db:"guid"`
		ID               int64     `db:"id"`
		CalculationsType string    `db:"calculations_type"`
		Name             string    `db:"name"`
		Description      string    `db:"description"`
		Questions        []byte    `db:"questions"`

		CreatedAt time.Time  `db:"created_at"`
		UpdatedAt time.Time  `db:"updated_at"`
		DeletedAt *time.Time `db:"deleted_at"`
	}

	surveyState struct {
		State      entity.State `db:"state"`
		UserGUID   uuid.UUID    `db:"user_guid"`
		SurveyGUID uuid.UUID    `db:"survey_guid"`
		Answers    []byte       `db:"answers"`
		CreatedAt  time.Time    `db:"created_at"`
		UpdatedAt  time.Time    `db:"updated_at"`
		Results    *[]byte      `db:"results"`
	}

	surveyStateReport struct {
		SurveyGUID  uuid.UUID `db:"survey_guid"`
		SurveyName  string    `db:"survey_name"`
		Description string    `db:"description"`
		StartedAt   time.Time `db:"created_at"`
		FinishedAt  time.Time `db:"updated_at"`
		UserGUID    string    `db:"user_guid"`
		UserID      int64     `db:"user_id"`
		Answers     []byte    `db:"answers"`
		Results     *[]byte   `db:"results"`
	}

	userListReportResponse struct {
		GUID         uuid.UUID     `db:"guid"`
		NickName     string        `db:"nickname"`
		CreatedAt    time.Time     `db:"created_at"`
		LastActivity time.Time     `db:"last_activity"`
		SurveyState  *entity.State `db:"state"`
		Answers      *[]byte       `db:"answers"`
	}
)

func (u user) Export() entity.User {
	return entity.User{
		GUID:          u.GUID,
		UserID:        u.UserID,
		ChatID:        u.ChatID,
		Nickname:      u.Nickname,
		CurrentSurvey: u.CurrentSurvey,
		LastActivity:  u.LastActivity,
	}
}

func (s survey) Export() (entity.Survey, error) {
	var questions []entity.Question
	if err := json.Unmarshal(s.Questions, &questions); err != nil {
		return entity.Survey{}, fmt.Errorf("failed to unmarshal questions: %w", err)
	}

	return entity.Survey{
		GUID:             s.GUID,
		ID:               s.ID,
		Name:             s.Name,
		Description:      s.Description,
		Questions:        questions,
		CalculationsType: s.CalculationsType,
	}, nil
}

func (s surveyState) Export() (entity.SurveyState, error) {
	var answers []entity.Answer
	if err := json.Unmarshal(s.Answers, &answers); err != nil {
		return entity.SurveyState{}, fmt.Errorf("failed to unmarshal answers: %w", err)
	}

	var results *entity.Results

	if (s.Results != nil) && (len(*s.Results) > 0) {
		if err := json.Unmarshal(*s.Results, &results); err != nil {
			return entity.SurveyState{}, fmt.Errorf("failed to unmarshal results: %w", err)
		}
	}

	return entity.SurveyState{
		State:      s.State,
		UserGUID:   s.UserGUID,
		SurveyGUID: s.SurveyGUID,
		Answers:    answers,
		Results:    results,
	}, nil
}

func (um *user) Load(u entity.User) {
	um.GUID = u.GUID
	um.UserID = u.UserID
	um.ChatID = u.ChatID
	um.Nickname = u.Nickname
	um.CurrentSurvey = u.CurrentSurvey
	um.LastActivity = u.LastActivity
}

func (s *survey) Load(survey entity.Survey) error {
	questions, err := json.Marshal(survey.Questions)
	if err != nil {
		return fmt.Errorf("failed to marshal questions: %w", err)
	}

	s.CalculationsType = survey.CalculationsType
	s.Name = survey.Name
	s.Description = survey.Description
	s.GUID = survey.GUID
	s.ID = survey.ID
	s.Questions = questions

	return nil
}

func (s *surveyState) Load(state entity.SurveyState) error {
	answers, err := json.Marshal(state.Answers)
	if err != nil {
		return fmt.Errorf("failed to marshal answers: %w", err)
	}

	if state.Results != nil {
		results, err := json.Marshal(state.Results)
		if err != nil {
			return fmt.Errorf("failed to marshal results: %w", err)
		}
		s.Results = &results
	}

	s.State = state.State
	s.UserGUID = state.UserGUID
	s.SurveyGUID = state.SurveyGUID
	s.Answers = answers

	return nil
}

func (s *surveyStateReport) Export() (entity.SurveyStateReport, error) {
	var answers []entity.Answer
	if err := json.Unmarshal(s.Answers, &answers); err != nil {
		return entity.SurveyStateReport{}, fmt.Errorf("failed to unmarshal answers: %w", err)
	}

	exported := entity.SurveyStateReport{
		SurveyGUID:  s.SurveyGUID,
		SurveyName:  s.SurveyName,
		Description: s.Description,
		StartedAt:   s.StartedAt,
		FinishedAt:  s.FinishedAt,
		UserGUID:    s.UserGUID,
		UserID:      s.UserID,
		Answers:     answers,
	}

	if (s.Results != nil) && (len(*s.Results) > 0) {
		var results entity.Results
		if err := json.Unmarshal(*s.Results, &results); err != nil {
			return entity.SurveyStateReport{}, fmt.Errorf("failed to unmarshal results: %w", err)
		}

		exported.Results = &results
	}

	return exported, nil
}
