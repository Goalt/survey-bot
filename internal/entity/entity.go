package entity

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/exp/utf8string"
)

const (
	FinishedState   State = "finished"
	ActiveState     State = "active"
	NotStartedState State = "not_started"

	AnswerTypeSegment     AnswerType = "segment"
	AnswerTypeSelect      AnswerType = "select"
	AnswerTypeMultiSelect AnswerType = "multiselect"
)

var (
	ErrParseAnswerTypeSegment     = errors.New("can't parse answer to segment")
	ErrParseAnswerTypeSelect      = errors.New("can't parse answer to select")
	ErrParseAnswerTypeMultiSelect = errors.New("can't parse answer to multiselect")
	ErrAnswerOutOfRange           = errors.New("answer is out of range")
	ErrAnswerNotFound             = errors.New("answer not found")
)

type (
	State      string
	AnswerType string

	User struct {
		GUID          uuid.UUID
		UserID        int64
		ChatID        int64
		Nickname      string
		CurrentSurvey *uuid.UUID
		LastActivity  time.Time
	}

	Survey struct {
		GUID             uuid.UUID
		CalculationsType string
		ID               int64
		Name             string
		Description      string
		Questions        []Question
	}

	SurveyState struct {
		State      State
		UserGUID   uuid.UUID
		SurveyGUID uuid.UUID
		Answers    []Answer

		// not nil if state is finished
		Results *Results
	}

	SurveyStateReport struct {
		SurveyGUID  uuid.UUID
		SurveyName  string
		Description string
		StartedAt   time.Time
		FinishedAt  time.Time
		UserGUID    string
		UserID      int64
		Answers     []Answer
		Results     *Results
	}

	Question struct {
		Text       string     `json:"text"`
		AnswerType AnswerType `json:"answer_type"`

		// PossibleAnswers equals to
		// 1. [min(int), max(int)] if Type == AnswerTypeSegment
		// 2. [answer_1(int), answer_2(int), ...] if Type == AnswerTypeMultiSelect || Type == AnswerTypeSelect
		PossibleAnswers []int    `json:"possible_answers"`
		AnswersText     []string `json:"answers_text"`
	}

	Answer struct {
		Type AnswerType `json:"type"`

		// Data equals to
		// 1. [value(int)] if Type == AnswerTypeSegment
		// 2. [answer_1(int)] if Type == AnswerTypeSelect
		// 3. [answer_1(int), answer_2(int), ...] if Type == AnswerTypeMultiSelect
		Data []int `json:"data"`
	}

	ResultsMetadata struct {
		Raw map[string]interface{}
	}

	Results struct {
		Text     string
		Metadata ResultsMetadata
	}

	ResultsProcessor interface {
		GetResults(survey Survey, answers []Answer) (Results, error)
		Validate(survey Survey) error
	}
)

func (q Question) GetAnswer(answerRaw string) (Answer, error) {
	switch q.AnswerType {
	case AnswerTypeSegment:
		return q.getAnswerSegment(answerRaw)
	case AnswerTypeSelect:
		return q.getAnswerSelect(answerRaw)
	case AnswerTypeMultiSelect:
		return q.getAnswerMultiSelect(answerRaw)
	default:
		return Answer{}, errors.New("unknown answer type")
	}
}

func (q Question) getAnswerMultiSelect(answerRaw string) (Answer, error) {
	args := strings.Split(answerRaw, ",")
	var answers []int
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		number, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			return Answer{}, fmt.Errorf("failed to parse argument, %w: %w", err, ErrParseAnswerTypeMultiSelect)
		}

		if !slices.Contains(q.PossibleAnswers, int(number)) {
			return Answer{}, ErrAnswerNotFound
		}

		answers = append(answers, int(number))
	}

	return Answer{
		Type: AnswerTypeMultiSelect,
		Data: answers,
	}, nil
}

func (q Question) getAnswerSelect(answerRaw string) (Answer, error) {
	number, err := strconv.ParseInt(answerRaw, 10, 64)
	if err != nil {
		return Answer{}, fmt.Errorf("failed to parse argument, %w: %w", err, ErrParseAnswerTypeSelect)
	}

	if slices.Contains(q.PossibleAnswers, int(number)) {
		return Answer{
			Type: AnswerTypeSelect,
			Data: []int{int(number)},
		}, nil
	}

	return Answer{}, ErrAnswerNotFound
}

func (q Question) getAnswerSegment(answerRaw string) (Answer, error) {
	number, err := strconv.ParseInt(answerRaw, 10, 64)
	if err != nil {
		return Answer{}, fmt.Errorf("failed to parse argument, %w: %w", err, ErrParseAnswerTypeSegment)
	}

	if number < int64(q.PossibleAnswers[0]) || number > int64(q.PossibleAnswers[1]) {
		return Answer{}, ErrAnswerOutOfRange
	}

	return Answer{
		Type: AnswerTypeSegment,
		Data: []int{int(number)},
	}, nil
}

func (q Question) Validate() error {
	text := utf8string.NewString(q.Text)

	if q.Text == "" {
		return errors.New("empty question text")
	}
	if !unicode.IsUpper(text.At(0)) {
		return errors.New("question text should start with uppercase letter")
	}

	lastSymbol := text.At(text.RuneCount() - 1)
	if lastSymbol != '?' && lastSymbol != '.' {
		return errors.New("question text should end with '?' or '.'")
	}

	for _, answerText := range q.AnswersText {
		if answerText == "" {
			return errors.New("empty answer text")
		}
	}

	switch q.AnswerType {
	case AnswerTypeSegment:
		if len(q.PossibleAnswers) != 2 {
			return errors.New("possible answers length should be 2")
		}
	case AnswerTypeSelect, AnswerTypeMultiSelect:
		if len(q.PossibleAnswers) == 0 {
			return errors.New("empty possible answers")
		}
		if len(q.PossibleAnswers) != len(q.AnswersText) {
			return errors.New("possible answers and answers text length mismatch")
		}

		for _, possibleAnswer := range q.PossibleAnswers {
			if possibleAnswer < 0 || (len(q.AnswersText) < possibleAnswer) {
				return errors.New("possible answer is out of range")
			}
		}
	default:
		return errors.New("unknown answer type")
	}

	return nil
}

func (s Survey) Validate() error {
	if s.Name == "" {
		return errors.New("empty survey name")
	}

	if s.Description == "" {
		return errors.New("empty survey description")
	}

	if s.CalculationsType == "" {
		return errors.New("empty calculations type")
	}

	if len(s.Questions) == 0 {
		return errors.New("empty questions")
	}

	for i, question := range s.Questions {
		if err := question.Validate(); err != nil {
			return fmt.Errorf("failed to validate question %d, %w", i, err)
		}
	}

	return nil
}

func (ss SurveyStateReport) ToCSV() ([]string, error) {
	var (
		text     string
		metadata []byte
		err      error
	)

	if ss.Results != nil {
		metadata, err = json.Marshal(ss.Results.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}

		text = ss.Results.Text
	}

	var result = []string{
		ss.SurveyGUID.String(),
		ss.SurveyName,
		ss.Description,
		ss.UserGUID,
		strconv.Itoa(int(ss.UserID)),
		text,
		string(metadata),
		ss.StartedAt.Format(time.RFC3339),
		ss.FinishedAt.Format(time.RFC3339),
	}

	for _, answer := range ss.Answers {
		switch answer.Type {
		case AnswerTypeSegment, AnswerTypeSelect:
			result = append(result, fmt.Sprintf("%d", answer.Data[0]))
		default:
			result = append(result, strings.Join(toStringSlice(answer.Data), " "))
		}
	}

	return result, nil
}

func toStringSlice(m []int) []string {
	var result []string

	for _, v := range m {
		result = append(result, strconv.Itoa(v))
	}

	return result
}
