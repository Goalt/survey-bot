package service

import (
	stdcontext "context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/google/uuid"

	"git.ykonkov.com/ykonkov/survey-bot/internal/context"
	"git.ykonkov.com/ykonkov/survey-bot/internal/entity"
	"git.ykonkov.com/ykonkov/survey-bot/internal/logger"
	"git.ykonkov.com/ykonkov/survey-bot/internal/responses"
)

var (
	UUIDProvider = uuid.New

	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type service struct {
	telegramRepo TelegramRepo
	dbRepo       DBRepo
	rsltProc     entity.ResultsProcessor

	logger logger.Logger
}

func New(tr TelegramRepo, dbRepo DBRepo, rsltProc entity.ResultsProcessor, logger logger.Logger) *service {
	return &service{

		telegramRepo: tr,
		dbRepo:       dbRepo,
		logger:       logger,
		rsltProc:     rsltProc,
	}
}

func (s *service) Transact(ctx stdcontext.Context, fn func(exec DBTransaction) error) error {
	tx, err := s.dbRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if err := tx.Rollback(); err != nil {
			s.logger.Errorf(ctx, "failed to rollback transaction: %w", err)
		}

		return fmt.Errorf("failed to exec transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *service) HandleSurveyCommand(ctx context.Context, surveyID int64) error {
	if err := s.Transact(ctx, func(tx DBTransaction) error {
		var (
			user entity.User
			err  error
		)

		user, err = s.dbRepo.GetUserByID(ctx, tx, ctx.UserID())
		switch {
		case errors.Is(err, ErrNotFound):
			user = entity.User{
				GUID:          UUIDProvider(),
				UserID:        ctx.UserID(),
				ChatID:        ctx.ChatID(),
				Nickname:      ctx.Nickname(),
				CurrentSurvey: nil,
			}

			if err := s.dbRepo.CreateUser(ctx, tx, user); err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
		case err != nil:
			return fmt.Errorf("failed to get user: %w", err)
		}

		survey, err := s.dbRepo.GetSurveyByID(ctx, tx, surveyID)
		if err != nil {
			return fmt.Errorf("failed to get survey: %w", err)
		}

		state, err := s.dbRepo.GetUserSurveyState(ctx, tx, user.GUID, survey.GUID, []entity.State{entity.ActiveState})
		switch {
		case errors.Is(err, ErrNotFound):
			state = entity.SurveyState{
				UserGUID:   user.GUID,
				SurveyGUID: survey.GUID,
				State:      entity.ActiveState,
			}

			if err := s.dbRepo.CreateUserSurveyState(ctx, tx, state); err != nil {
				return fmt.Errorf("failed to create user survey state: %w", err)
			}
		case err != nil:
			return fmt.Errorf("failed to get user survey state: %w", err)
		case state.State == entity.FinishedState:
			if err := s.telegramRepo.SendMessage(ctx, responses.SurveyAlreadyFinished); err != nil {
				s.logger.Errorf(ctx, "failed to send error message: %w", err)
			}

			return fmt.Errorf("survey already finished")
		}

		if err := s.dbRepo.UpdateUserCurrentSurvey(ctx, tx, user.GUID, survey.GUID); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		if err := s.dbRepo.UpdateUserLastActivity(ctx, tx, user.GUID); err != nil {
			return fmt.Errorf("failed to update user's last activity: %w", err)
		}

		lastQuestionNumber := len(state.Answers) - 1
		lastQuestion := survey.Questions[lastQuestionNumber+1]
		if err := s.telegramRepo.SendSurveyQuestion(ctx, lastQuestion); err != nil {
			return fmt.Errorf("failed to send survey question: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to transact: %w", err)
	}

	return nil
}

func (s *service) HandleListCommand(ctx context.Context) error {
	if err := s.Transact(ctx, func(tx DBTransaction) error {
		var (
			user entity.User
			err  error
		)

		user, err = s.dbRepo.GetUserByID(ctx, tx, ctx.UserID())
		switch {
		case errors.Is(err, ErrNotFound):
			user = entity.User{
				GUID:          UUIDProvider(),
				UserID:        ctx.UserID(),
				ChatID:        ctx.ChatID(),
				Nickname:      ctx.Nickname(),
				CurrentSurvey: nil,
			}

			if err := s.dbRepo.CreateUser(ctx, tx, user); err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
		case err != nil:
			return fmt.Errorf("failed to get user: %w", err)
		}

		if err := s.dbRepo.UpdateUserLastActivity(ctx, tx, user.GUID); err != nil {
			return fmt.Errorf("failed to update user's last activity: %w", err)
		}

		if err := s.sendUserSurveyList(ctx, tx, user); err != nil {
			return fmt.Errorf("failed to send user survey list: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to transact: %w", err)
	}

	return nil
}

func (s *service) HandleAnswer(ctx context.Context, msg string) error {
	if err := s.Transact(ctx, func(tx DBTransaction) error {
		var (
			user entity.User
			err  error
		)

		user, err = s.dbRepo.GetUserByID(ctx, tx, ctx.UserID())
		switch {
		case errors.Is(err, ErrNotFound):
			user = entity.User{
				GUID:          UUIDProvider(),
				UserID:        ctx.UserID(),
				ChatID:        ctx.ChatID(),
				Nickname:      ctx.Nickname(),
				CurrentSurvey: nil,
			}

			if err := s.dbRepo.CreateUser(ctx, tx, user); err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
		case err != nil:
			return fmt.Errorf("failed to get user: %w", err)
		}

		if err := s.dbRepo.UpdateUserLastActivity(ctx, tx, user.GUID); err != nil {
			return fmt.Errorf("failed to update user's last activity: %w", err)
		}

		if user.CurrentSurvey == nil {
			if err := s.telegramRepo.SendMessage(ctx, responses.ChooseSurvey); err != nil {
				s.logger.Errorf(ctx, "failed to send error message: %w", err)
			}
			return fmt.Errorf("user does not have current survey")
		}

		state, err := s.dbRepo.GetUserSurveyState(ctx, tx, user.GUID, *user.CurrentSurvey, []entity.State{entity.ActiveState})
		if err != nil {
			return fmt.Errorf("failed to get user survey state: %w", err)
		}

		survey, err := s.dbRepo.GetSurvey(ctx, tx, *user.CurrentSurvey)
		if err != nil {
			return fmt.Errorf("failed to get survey: %w", err)
		}

		lastQuestionNumber := len(state.Answers)
		if lastQuestionNumber >= len(survey.Questions) {
			return fmt.Errorf("last question is out of range")
		}
		lastQuestion := survey.Questions[lastQuestionNumber]
		answer, err := lastQuestion.GetAnswer(msg)
		if err != nil {
			s.handleAnswerValidateError(ctx, err)

			return fmt.Errorf("failed to get answer: %w", err)
		}

		state.Answers = append(state.Answers, answer)

		if lastQuestionNumber == len(survey.Questions)-1 {
			// if it was last question
			results, err := s.rsltProc.GetResults(survey, state.Answers)
			if err != nil {
				return fmt.Errorf("failed to get results: %w", err)
			}

			if err := s.dbRepo.SetUserCurrentSurveyToNil(ctx, tx, user.GUID); err != nil {
				return fmt.Errorf("failed to set current user survey to null: %w", err)
			}

			if err := s.telegramRepo.SendMessage(ctx, results.Text); err != nil {
				s.logger.Errorf(ctx, "failed to send results: %w", err)
			}

			state.State = entity.FinishedState
			state.Results = &results
		} else {
			// otherwise send next question
			nextQuestion := survey.Questions[lastQuestionNumber+1]
			if err := s.telegramRepo.SendSurveyQuestion(ctx, nextQuestion); err != nil {
				return fmt.Errorf("failed to send survey question: %w", err)
			}
		}

		if err := s.dbRepo.UpdateActiveUserSurveyState(ctx, tx, state); err != nil {
			return fmt.Errorf("failed to update user survey state: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to transact: %w", err)
	}

	return nil
}

func (s *service) HandleStartCommand(ctx context.Context) error {
	if err := s.Transact(ctx, func(tx DBTransaction) error {
		var (
			user entity.User
			err  error
		)

		user, err = s.dbRepo.GetUserByID(ctx, tx, ctx.UserID())
		switch {
		case errors.Is(err, ErrNotFound):
			user = entity.User{
				GUID:          UUIDProvider(),
				UserID:        ctx.UserID(),
				ChatID:        ctx.ChatID(),
				Nickname:      ctx.Nickname(),
				CurrentSurvey: nil,
			}

			if err := s.dbRepo.CreateUser(ctx, tx, user); err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
		case err != nil:
			return fmt.Errorf("failed to create user: %w", err)
		default:
			s.logger.Infof(ctx, "user already pressed start command")
		}

		if err := s.dbRepo.UpdateUserLastActivity(ctx, tx, user.GUID); err != nil {
			return fmt.Errorf("failed to update user's last activity: %w", err)
		}

		if err := s.sendUserSurveyList(ctx, tx, user); err != nil {
			return fmt.Errorf("failed to send user survey list: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to transact: %w", err)
	}

	return nil
}

func (s *service) getSurveyList(ctx stdcontext.Context, tx DBTransaction) ([]entity.Survey, error) {
	return s.dbRepo.GetSurveysList(ctx, tx)
}

func (s *service) CreateSurvey(ctx stdcontext.Context, survey entity.Survey) (entity.Survey, error) {
	if err := s.Transact(ctx, func(tx DBTransaction) error {
		if err := s.rsltProc.Validate(survey); err != nil {
			return fmt.Errorf("failed to validate survey: %w", err)
		}

		// get surveys list
		surveys, err := s.getSurveyList(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to get surveys list: %w", err)
		}

		// get max survey id
		var maxSurveyID int64
		for _, survey := range surveys {
			if survey.ID > maxSurveyID {
				maxSurveyID = survey.ID
			}
		}

		survey.ID = maxSurveyID + 1
		survey.GUID = UUIDProvider()

		if err := s.dbRepo.CreateSurvey(ctx, tx, survey); err != nil {
			return fmt.Errorf("failed to create survey: %w", err)
		}

		return nil
	}); err != nil {
		return entity.Survey{}, fmt.Errorf("failed to transact: %w", err)
	}

	return survey, nil
}

func (s *service) sendUserSurveyList(ctx context.Context, tx DBTransaction, user entity.User) error {
	all, err := s.dbRepo.GetSurveysList(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to get all surveys: %w", err)
	}

	userStates, err := s.dbRepo.GetUserSurveyStates(ctx, tx, user.GUID, []entity.State{entity.ActiveState})
	if err != nil {
		return fmt.Errorf("failed to get user survey states: %w", err)
	}

	// sort surveys by id
	sort.Slice(all, func(i, j int) bool {
		return all[i].ID < all[j].ID
	})

	var states []UserSurveyState
	for _, survey := range all {
		var (
			state     entity.State
			isFound   bool
			isCurrent bool
		)

		for _, userState := range userStates {
			if userState.SurveyGUID == survey.GUID {
				state = userState.State
				isFound = true
				break
			}
		}

		if !isFound {
			state = entity.NotStartedState
		}

		if (user.CurrentSurvey != nil) && (survey.GUID == *user.CurrentSurvey) {
			isCurrent = true
		}

		states = append(states, UserSurveyState{
			UserGUID:  user.GUID,
			Survey:    survey,
			State:     state,
			IsCurrent: isCurrent,
		})
	}

	if err := s.telegramRepo.SendSurveyList(ctx, states); err != nil {
		return fmt.Errorf("failed to send survey list: %w", err)
	}

	return nil
}

func (s *service) handleAnswerValidateError(ctx context.Context, err error) {
	errText := "failed to parse answer"
	switch {
	case errors.Is(err, entity.ErrParseAnswerTypeSegment):
		errText = responses.AnswerNotANumber
	case errors.Is(err, entity.ErrParseAnswerTypeSelect):
		errText = responses.AnswerNotANumber
	case errors.Is(err, entity.ErrParseAnswerTypeMultiSelect):
		errText = responses.AnswerNotANumber
	case errors.Is(err, entity.ErrAnswerOutOfRange):
		errText = responses.AnswerOutOfRange
	case errors.Is(err, entity.ErrAnswerNotFound):
		errText = responses.AnswerNotFound
	}

	if err := ctx.Send(errText); err != nil {
		s.logger.Errorf(ctx, "failed to send error message: %w", err)
	}
}

// GetUserByGUID returns a user by their GUID.
func (s *service) GetUserByGUID(ctx stdcontext.Context, guid uuid.UUID) (entity.User, error) {
	var user entity.User
	if err := s.Transact(ctx, func(tx DBTransaction) error {
		var err error
		user, err = s.dbRepo.GetUserByGUID(ctx, tx, guid)

		return err
	}); err != nil {
		return entity.User{}, fmt.Errorf("failed to transact: %w", err)
	}

	return user, nil
}

func (s *service) SetUserCurrentSurveyToNil(ctx stdcontext.Context, userGUID uuid.UUID) error {
	if err := s.Transact(ctx, func(tx DBTransaction) error {
		return s.dbRepo.SetUserCurrentSurveyToNil(ctx, tx, userGUID)
	}); err != nil {
		return fmt.Errorf("failed to transact: %w", err)
	}

	return nil
}

// DeleteUserSurvey deletes a user survey.
func (s *service) DeleteUserSurvey(ctx stdcontext.Context, userGUID uuid.UUID, surveyGUID uuid.UUID) error {
	if err := s.Transact(ctx, func(tx DBTransaction) error {
		return s.dbRepo.DeleteUserSurveyState(ctx, tx, userGUID, surveyGUID)
	}); err != nil {
		return fmt.Errorf("failed to transact: %w", err)
	}

	return nil
}

func (s *service) SaveFinishedSurveys(ctx stdcontext.Context, tx DBTransaction, w io.Writer, f ResultsFilter, batchSize int) (int, error) {
	writer := csv.NewWriter(w)

	offset := 0
	total := 0

	if err := writer.Write([]string{"survey_guid", "survey_name", "user_guid", "user_id", "text", "metadata", "started_at", "finished_at", "answers"}); err != nil {
		return 0, fmt.Errorf("failed to write header to csv: %w", err)
	}

	for {
		// get batch of finished surveys
		states, err := s.dbRepo.GetFinishedSurveys(ctx, tx, f, batchSize, offset)
		if err != nil {
			return 0, fmt.Errorf("failed to get finished surveys: %w", err)
		}

		if len(states) == 0 {
			break
		}

		// save results to file
		for _, state := range states {
			columns, err := state.ToCSV()
			if err != nil {
				return 0, fmt.Errorf("failed to convert survey state to csv: %w", err)
			}

			if err := writer.Write(columns); err != nil {
				return 0, fmt.Errorf("failed to write survey to csv: %w", err)
			}
		}

		writer.Flush()
		offset += batchSize
		total += len(states)
	}

	return total, nil
}

func ReadSurveyFromFile(filename string) (entity.Survey, error) {
	file, err := os.Open(filename)
	if err != nil {
		return entity.Survey{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(file)
	if err != nil {
		return entity.Survey{}, fmt.Errorf("failed to read file: %w", err)
	}

	type survey struct {
		Questions        []entity.Question `json:"questions"`
		Name             string            `json:"name"`
		CalculationsType string            `json:"calculations_type"`
		Description      string            `json:"description"`
	}

	var s survey
	if err := json.Unmarshal(data, &s); err != nil {
		return entity.Survey{}, fmt.Errorf("failed to unmarshal file: %w", err)
	}

	return entity.Survey{
		Name:             s.Name,
		Questions:        s.Questions,
		CalculationsType: s.CalculationsType,
		Description:      s.Description,
	}, nil
}

func (s *service) UpdateSurvey(ctx stdcontext.Context, new entity.Survey) error {
	if err := s.Transact(ctx, func(tx DBTransaction) error {
		// find survey by guid
		old, err := s.dbRepo.GetSurvey(ctx, tx, new.GUID)
		if err != nil {
			return fmt.Errorf("failed to find survey: %w", err)
		}

		// if len of questions is not equal, then do not update
		if len(old.Questions) != len(new.Questions) {
			return fmt.Errorf("cannot update survey with different number of questions")
		}

		// update name, questions and calculations_type
		old.Name = new.Name
		old.Questions = new.Questions
		old.CalculationsType = new.CalculationsType
		old.Description = new.Description

		if err := s.dbRepo.UpdateSurvey(ctx, tx, old); err != nil {
			return fmt.Errorf("failed to update survey: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to transact: %w", err)
	}

	return nil
}

func (s *service) HandleResultsCommand(ctx context.Context, f ResultsFilter) error {
	if err := s.Transact(ctx, func(tx DBTransaction) error {
		_, err := s.dbRepo.GetUserByID(ctx, tx, ctx.UserID())
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		s.logger.Infof(ctx, "user pressed results command")

		// save file to temporary directory
		filename := fmt.Sprintf("results-%v.csv", time.Now())
		filePath := filepath.Join(os.TempDir(), filename)

		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer func() { _ = file.Close() }()

		total, err := s.SaveFinishedSurveys(ctx, tx, file, f, 100)
		if err != nil {
			return fmt.Errorf("failed to save finished surveys: %w", err)
		}

		if total == 0 {
			if err := s.telegramRepo.SendMessage(ctx, responses.NoResults); err != nil {
				s.logger.Errorf(ctx, "failed to send error message: %w", err)
			}

			return nil
		}

		if err := s.telegramRepo.SendFile(ctx, filePath); err != nil {
			return fmt.Errorf("failed to send file: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to transact: %w", err)
	}

	return nil
}

func (s *service) GetCompletedSurveys(ctx stdcontext.Context, userID int64) ([]entity.SurveyStateReport, error) {
	var surveys []entity.SurveyStateReport
	if err := s.Transact(ctx, func(tx DBTransaction) error {
		user, err := s.dbRepo.GetUserByID(ctx, tx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		if err := s.dbRepo.UpdateUserLastActivity(ctx, tx, user.GUID); err != nil {
			return fmt.Errorf("failed to update user's last activity: %w", err)
		}

		surveys, err = s.dbRepo.GetCompletedSurveys(ctx, tx, user.GUID)
		return err
	}); err != nil {
		return nil, fmt.Errorf("failed to transact: %w", err)
	}
	return surveys, nil
}

func (s *service) GetUsersList(ctx stdcontext.Context, limit, offset int, search string) (UserListResponse, error) {
	var usersList UserListResponse
	if err := s.Transact(ctx, func(tx DBTransaction) error {
		var err error
		usersList, err = s.dbRepo.GetUsersList(ctx, tx, limit, offset, search)
		return err
	}); err != nil {
		return UserListResponse{}, fmt.Errorf("failed to transact: %w", err)
	}
	return usersList, nil
}
