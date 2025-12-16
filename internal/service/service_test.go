package service_test

import (
	stdcontext "context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"git.ykonkov.com/ykonkov/survey-bot/internal/context"
	"git.ykonkov.com/ykonkov/survey-bot/internal/entity"
	"git.ykonkov.com/ykonkov/survey-bot/internal/responses"
	"git.ykonkov.com/ykonkov/survey-bot/internal/service"
	mocks "git.ykonkov.com/ykonkov/survey-bot/internal/service/mocks"
)

type ServiceTestSuite struct {
	suite.Suite

	telegramRepo *mocks.TelegramRepo
	dbRepo       *mocks.DBRepo
	resultsProc  *mocks.ResultsProcessor
	logger       *mocks.Logger

	svc service.Service
}

func (suite *ServiceTestSuite) SetupTest() {
	suite.dbRepo = mocks.NewDBRepo(suite.T())
	suite.telegramRepo = mocks.NewTelegramRepo(suite.T())
	suite.resultsProc = mocks.NewResultsProcessor(suite.T())
	suite.logger = mocks.NewLogger(suite.T())

	suite.svc = service.New(suite.telegramRepo, suite.dbRepo, suite.resultsProc, suite.logger)
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (suite *ServiceTestSuite) TestHandleStartCommand() {
	ctx := newTestContext(stdcontext.Background(), 10, 33, []string{"start"})

	service.UUIDProvider = func() uuid.UUID {
		return uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947")
	}

	tx := mocks.NewDBTransaction(suite.T())
	suite.dbRepo.On(
		"BeginTx",
		ctx,
	).Return(tx, nil)

	suite.dbRepo.On(
		"GetUserByID",
		ctx,
		tx,
		int64(10),
	).Return(entity.User{}, service.ErrNotFound)

	suite.dbRepo.On(
		"CreateUser",
		ctx,
		tx,
		entity.User{
			UserID:   10,
			GUID:     uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
			ChatID:   33,
			Nickname: "nickname",
		},
	).Return(nil)

	suite.dbRepo.On(
		"UpdateUserLastActivity",
		ctx,
		tx,
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
	).Return(nil)

	suite.dbRepo.On("GetSurveysList", ctx, tx).Return(
		suite.generateTestSurveyList(),
		nil,
	)

	suite.dbRepo.On(
		"GetUserSurveyStates",
		ctx, tx, uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
		[]entity.State{entity.ActiveState},
	).Return(
		suite.generateSurveyStates(),
		nil,
	)

	suite.telegramRepo.On(
		"SendSurveyList",
		ctx,
		suite.generateTestUserSurveyList(),
	).Return(nil)

	tx.On("Commit").Return(nil)

	err := suite.svc.HandleStartCommand(ctx)
	suite.NoError(err)
}

func (suite *ServiceTestSuite) TestHandleStartCommand_UserAlreadyExists() {
	ctx := newTestContext(stdcontext.Background(), 10, 33, []string{"start"})

	service.UUIDProvider = func() uuid.UUID {
		return uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947")
	}

	tx := mocks.NewDBTransaction(suite.T())
	suite.dbRepo.On(
		"BeginTx",
		ctx,
	).Return(tx, nil)

	suite.dbRepo.On(
		"GetUserByID",
		ctx,
		tx,
		int64(10),
	).Return(entity.User{
		GUID:          uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
		UserID:        10,
		ChatID:        33,
		CurrentSurvey: nil,
	}, nil)

	suite.logger.On("Infof", ctx, "user already pressed start command")

	suite.dbRepo.On("GetSurveysList", ctx, tx).Return(
		suite.generateTestSurveyList(),
		nil,
	)

	suite.dbRepo.On(
		"UpdateUserLastActivity",
		ctx,
		tx,
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
	).Return(nil)

	suite.dbRepo.On("GetSurveysList", ctx, tx).Return(
		suite.generateTestSurveyList(),
		nil,
	)

	suite.dbRepo.On(
		"GetUserSurveyStates",
		ctx,
		tx,
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
		[]entity.State{entity.ActiveState},
	).Return(
		suite.generateSurveyStates(),
		nil,
	)

	suite.telegramRepo.On(
		"SendSurveyList",
		ctx,
		suite.generateTestUserSurveyList(),
	).Return(nil)

	tx.On("Commit").Return(nil)

	err := suite.svc.HandleStartCommand(ctx)
	suite.NoError(err)
}

func (suite *ServiceTestSuite) TestHandleListCommand() {
	ctx := newTestContext(stdcontext.Background(), 10, 33, []string{"start"})

	tx := mocks.NewDBTransaction(suite.T())
	suite.dbRepo.On(
		"BeginTx",
		ctx,
	).Return(tx, nil)

	suite.dbRepo.On("GetUserByID", ctx, tx, int64(10)).Return(entity.User{
		UserID: 10,
		GUID:   uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
	}, nil)

	suite.dbRepo.On(
		"UpdateUserLastActivity",
		ctx,
		tx,
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
	).Return(nil)

	suite.dbRepo.On("GetSurveysList", ctx, tx).Return(
		suite.generateTestSurveyList(),
		nil,
	)

	suite.dbRepo.On("GetSurveysList", ctx, tx).Return(
		suite.generateTestSurveyList(),
		nil,
	)

	suite.dbRepo.On(
		"GetUserSurveyStates",
		ctx,
		tx,
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
		[]entity.State{entity.ActiveState},
	).Return(
		suite.generateSurveyStates(),
		nil,
	)

	suite.telegramRepo.On(
		"SendSurveyList",
		ctx,
		suite.generateTestUserSurveyList(),
	).Return(nil)

	suite.telegramRepo.On(
		"SendSurveyList",
		ctx,
		suite.generateTestUserSurveyList(),
	).Return(nil)

	tx.On("Commit").Return(nil)

	err := suite.svc.HandleListCommand(ctx)
	suite.NoError(err)
}

func (suite *ServiceTestSuite) TestHandleSurveyCommand() {
	ctx := newTestContext(stdcontext.Background(), 10, 33, []string{"start"})

	tx := mocks.NewDBTransaction(suite.T())
	suite.dbRepo.On(
		"BeginTx",
		ctx,
	).Return(tx, nil)

	suite.dbRepo.On("GetUserByID", ctx, tx, int64(10)).Return(entity.User{
		UserID: 10,
		GUID:   uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
	}, nil)

	suite.dbRepo.On("GetSurveyByID", ctx, tx, int64(1)).Return(
		suite.generateTestSurveyList()[0],
		nil,
	)

	suite.dbRepo.On(
		"GetUserSurveyState",
		ctx,
		tx,
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
		[]entity.State{entity.ActiveState},
	).Return(
		entity.SurveyState{
			SurveyGUID: uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
			UserGUID:   uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
			State:      entity.NotStartedState,
		},
		nil,
	)

	suite.dbRepo.On(
		"UpdateUserCurrentSurvey",
		ctx,
		tx,
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
	).Return(nil)

	suite.dbRepo.On(
		"UpdateUserLastActivity",
		ctx,
		tx,
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
	).Return(nil)

	suite.telegramRepo.On(
		"SendSurveyQuestion",
		ctx,
		entity.Question{
			Text:            "Question 1",
			AnswerType:      entity.AnswerTypeSelect,
			PossibleAnswers: []int{1, 2, 3, 4},
			AnswersText:     []string{"variant 1", "variant 2", "variant 3", "variant 4"},
		},
	).Return(nil)

	tx.On("Commit").Return(nil)

	err := suite.svc.HandleSurveyCommand(ctx, 1)
	suite.NoError(err)
}

func (suite *ServiceTestSuite) TestHandleSurveyCommand_UserCreatedIfNotFound() {
	ctx := newTestContext(stdcontext.Background(), 10, 33, []string{"start"})

	// Fix UUID for deterministic expectations
	service.UUIDProvider = func() uuid.UUID {
		return uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947")
	}

	tx := mocks.NewDBTransaction(suite.T())
	suite.dbRepo.On(
		"BeginTx",
		ctx,
	).Return(tx, nil)

	// User not found -> create user
	suite.dbRepo.On("GetUserByID", ctx, tx, int64(10)).Return(entity.User{}, service.ErrNotFound)

	suite.dbRepo.On(
		"CreateUser",
		ctx,
		tx,
		entity.User{
			UserID:   10,
			GUID:     uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
			ChatID:   33,
			Nickname: "nickname",
		},
	).Return(nil)

	suite.dbRepo.On("GetSurveyByID", ctx, tx, int64(1)).Return(
		suite.generateTestSurveyList()[0],
		nil,
	)

	// Assume no state yet -> will be created
	suite.dbRepo.On(
		"GetUserSurveyState",
		ctx,
		tx,
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
		[]entity.State{entity.ActiveState},
	).Return(
		entity.SurveyState{},
		service.ErrNotFound,
	)

	suite.dbRepo.On(
		"CreateUserSurveyState",
		ctx,
		tx,
		entity.SurveyState{
			SurveyGUID: uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
			UserGUID:   uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
			State:      entity.ActiveState,
		},
	).Return(nil)

	suite.dbRepo.On(
		"UpdateUserCurrentSurvey",
		ctx,
		tx,
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
	).Return(nil)

	suite.dbRepo.On(
		"UpdateUserLastActivity",
		ctx,
		tx,
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
	).Return(nil)

	suite.telegramRepo.On(
		"SendSurveyQuestion",
		ctx,
		entity.Question{
			Text:            "Question 1",
			AnswerType:      entity.AnswerTypeSelect,
			PossibleAnswers: []int{1, 2, 3, 4},
			AnswersText:     []string{"variant 1", "variant 2", "variant 3", "variant 4"},
		},
	).Return(nil)

	tx.On("Commit").Return(nil)

	err := suite.svc.HandleSurveyCommand(ctx, 1)
	suite.NoError(err)
}

func (suite *ServiceTestSuite) TestHandleSurveyCommand_SurveyAlreadFinished() {
	ctx := newTestContext(stdcontext.Background(), 10, 33, []string{"start"})

	tx := mocks.NewDBTransaction(suite.T())
	suite.dbRepo.On(
		"BeginTx",
		ctx,
	).Return(tx, nil)

	suite.dbRepo.On("GetUserByID", ctx, tx, int64(10)).Return(entity.User{
		UserID: 10,
		GUID:   uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
	}, nil)

	suite.dbRepo.On("GetSurveyByID", ctx, tx, int64(1)).Return(
		suite.generateTestSurveyList()[0],
		nil,
	)

	suite.dbRepo.On(
		"GetUserSurveyState",
		ctx,
		tx,
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
		[]entity.State{entity.ActiveState},
	).Return(
		entity.SurveyState{
			SurveyGUID: uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
			UserGUID:   uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
			State:      entity.FinishedState,
		},
		nil,
	)

	suite.telegramRepo.On(
		"SendMessage",
		ctx,
		responses.SurveyAlreadyFinished,
	).Return(nil)

	tx.On("Rollback").Return(nil)

	err := suite.svc.HandleSurveyCommand(ctx, 1)
	suite.ErrorContains(err, "survey already finished")
}

func (suite *ServiceTestSuite) TestHandleSurveyCommand_SurveyNotYetStarted() {
	ctx := newTestContext(stdcontext.Background(), 10, 33, []string{"start"})

	tx := mocks.NewDBTransaction(suite.T())
	suite.dbRepo.On(
		"BeginTx",
		ctx,
	).Return(tx, nil)

	suite.dbRepo.On("GetUserByID", ctx, tx, int64(10)).Return(entity.User{
		UserID: 10,
		GUID:   uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
	}, nil)

	suite.dbRepo.On("GetSurveyByID", ctx, tx, int64(1)).Return(
		suite.generateTestSurveyList()[0],
		nil,
	)

	suite.dbRepo.On(
		"GetUserSurveyState",
		ctx,
		tx,
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
		[]entity.State{entity.ActiveState},
	).Return(
		entity.SurveyState{},
		service.ErrNotFound,
	)

	suite.dbRepo.On(
		"CreateUserSurveyState",
		ctx,
		tx,
		entity.SurveyState{
			SurveyGUID: uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
			UserGUID:   uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
			State:      entity.ActiveState,
		},
	).Return(nil)

	suite.dbRepo.On(
		"UpdateUserCurrentSurvey",
		ctx,
		tx,
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
	).Return(nil)

	suite.dbRepo.On(
		"UpdateUserLastActivity",
		ctx,
		tx,
		uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
	).Return(nil)

	suite.telegramRepo.On(
		"SendSurveyQuestion",
		ctx,
		entity.Question{
			Text:            "Question 1",
			AnswerType:      entity.AnswerTypeSelect,
			PossibleAnswers: []int{1, 2, 3, 4},
			AnswersText:     []string{"variant 1", "variant 2", "variant 3", "variant 4"},
		},
	).Return(nil)

	tx.On("Commit").Return(nil)

	err := suite.svc.HandleSurveyCommand(ctx, 1)
	suite.NoError(err)
}

func (suite *ServiceTestSuite) generateSurveyStates() []entity.SurveyState {
	surveys := suite.generateTestSurveyList()
	return []entity.SurveyState{
		{
			SurveyGUID: surveys[0].GUID,
			UserGUID:   uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
			State:      entity.NotStartedState,
		},
	}
}

func (suite *ServiceTestSuite) generateTestUserSurveyList() []service.UserSurveyState {
	surveys := suite.generateTestSurveyList()
	return []service.UserSurveyState{
		{
			Survey:    surveys[0],
			UserGUID:  uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
			State:     entity.NotStartedState,
			IsCurrent: false,
		},
		{
			Survey:    surveys[1],
			UserGUID:  uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
			State:     entity.NotStartedState,
			IsCurrent: false,
		},
		{
			Survey:    surveys[2],
			UserGUID:  uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
			State:     entity.NotStartedState,
			IsCurrent: false,
		},
	}
}

func (suite *ServiceTestSuite) generateTestSurveyList() []entity.Survey {
	return []entity.Survey{
		{
			ID:   1,
			GUID: uuid.MustParse("EDDF980A-73E9-458B-926D-13B79BC2E947"),
			Questions: []entity.Question{
				{
					Text:            "Question 1",
					AnswerType:      entity.AnswerTypeSelect,
					PossibleAnswers: []int{1, 2, 3, 4},
					AnswersText:     []string{"variant 1", "variant 2", "variant 3", "variant 4"},
				},
				{
					Text:            "Question 2",
					AnswerType:      entity.AnswerTypeSegment,
					PossibleAnswers: []int{1, 5},
				},
				{
					Text:            "Question 3",
					AnswerType:      entity.AnswerTypeSelect,
					PossibleAnswers: []int{1, 2, 3, 4},
					AnswersText:     []string{"variant 1", "variant 2", "variant 3", "variant 4"},
				},
			},
		},
		{
			ID:   2,
			GUID: uuid.MustParse("91DEF2EA-829D-443E-BCBF-FA2EF8283214"),
			Questions: []entity.Question{
				{
					Text:            "Question 1",
					AnswerType:      entity.AnswerTypeSelect,
					PossibleAnswers: []int{1, 2, 3, 4},
					AnswersText:     []string{"variant 1", "variant 2", "variant 3", "variant 4"},
				},
				{
					Text:            "Question 2",
					AnswerType:      entity.AnswerTypeSegment,
					PossibleAnswers: []int{1, 5},
				},
				{
					Text:            "Question 3",
					AnswerType:      entity.AnswerTypeSelect,
					PossibleAnswers: []int{1, 2, 3, 4},
					AnswersText:     []string{"variant 1", "variant 2", "variant 3", "variant 4"},
				},
			},
		},
		{
			ID:   3,
			GUID: uuid.MustParse("01046E61-F452-47F6-A149-46EC267BBF2F"),
			Questions: []entity.Question{
				{
					Text:            "Question 1",
					AnswerType:      entity.AnswerTypeSelect,
					PossibleAnswers: []int{1, 2, 3, 4},
					AnswersText:     []string{"variant 1", "variant 2", "variant 3", "variant 4"},
				},
				{
					Text:            "Question 2",
					AnswerType:      entity.AnswerTypeSegment,
					PossibleAnswers: []int{1, 5},
				},
				{
					Text:            "Question 3",
					AnswerType:      entity.AnswerTypeSelect,
					PossibleAnswers: []int{1, 2, 3, 4},
					AnswersText:     []string{"variant 1", "variant 2", "variant 3", "variant 4"},
				},
			},
		},
	}
}

type testContext struct {
	userID int64
	chatID int64
	msg    []string
	stdcontext.Context
}

func newTestContext(ctx stdcontext.Context, userID int64, chatID int64, msg []string) context.Context {
	return &testContext{
		userID:  userID,
		chatID:  chatID,
		msg:     msg,
		Context: ctx,
	}
}

func (c *testContext) Send(msg interface{}, options ...interface{}) error {
	return nil
}

func (c *testContext) UserID() int64 {
	return c.userID
}

func (c *testContext) ChatID() int64 {
	return c.chatID
}

func (c *testContext) SetStdContext(ctx stdcontext.Context) {
	c.Context = ctx
}

func (c *testContext) Nickname() string {
	return "nickname"
}
