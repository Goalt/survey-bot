package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	initdata "github.com/telegram-mini-apps/init-data-golang"

	"git.ykonkov.com/ykonkov/survey-bot/internal/logger"
	"git.ykonkov.com/ykonkov/survey-bot/internal/service"
)

type userIDKey string

const userIDKeyType = userIDKey("userID")

type apiServer struct {
	server *http.Server
	svc    service.Service
	log    logger.Logger

	telegramToken  string
	allowedOrigins string
	adminUserIDs   []int64
}

func NewAPIServer(
	port int,
	telegramToken string,
	allowedOrigins string,
	adminUserIDs []int64,
	svc service.Service,
	log logger.Logger,
) *apiServer {
	sentryHandler := sentryhttp.New(sentryhttp.Options{})

	server := &apiServer{
		svc:            svc,
		telegramToken:  telegramToken,
		allowedOrigins: allowedOrigins,
		adminUserIDs:   adminUserIDs,
		log:            log,
		server: &http.Server{
			Addr: ":" + fmt.Sprintf("%d", port),
		},
	}

	handler := http.NewServeMux()
	handler.Handle("/api/surveys", sentryHandler.HandleFunc(server.optionsMiddleware(server.authMiddleware(server.handleCompletedSurveys))))
	handler.Handle("/api/is-admin", sentryHandler.HandleFunc(server.optionsMiddleware(server.authMiddleware(server.handleIsAdmin))))
	handler.Handle("/api/admin/users", sentryHandler.HandleFunc(
		server.optionsMiddleware(server.authMiddleware(server.authAdminMiddleware(server.handleUsersList)))),
	)

	server.server.Handler = handler

	return server
}

func (s *apiServer) Start() error {
	return s.server.ListenAndServe()
}

func (s *apiServer) Stop() error {
	return s.server.Shutdown(context.Background())
}

func (s *apiServer) optionsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			s.optionsResponse(w)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (s *apiServer) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span := sentry.StartSpan(
			r.Context(),
			"authMiddleware",
			sentry.ContinueFromHeaders(r.Header.Get(sentry.SentryTraceHeader), r.Header.Get(sentry.SentryBaggageHeader)),
		)
		defer span.Finish()

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.log.Errorf(r.Context(), "no auth header")
			s.writeError(r.Context(), w, Error{Code: http.StatusUnauthorized, Description: "No auth header"})

			return
		}

		initData := strings.TrimPrefix(authHeader, "tma ")
		if initData == "" {
			s.log.Errorf(r.Context(), "no token")
			s.writeError(r.Context(), w, Error{Code: http.StatusUnauthorized, Description: "Token empty"})

			return
		}

		expIn := 24 * time.Hour

		if err := initdata.Validate(initData, s.telegramToken, expIn); err != nil {
			s.log.Errorf(r.Context(), "invalid token: %v", err)
			s.writeError(r.Context(), w, Error{Code: http.StatusUnauthorized, Description: "Invalid token"})

			return
		}

		data, err := initdata.Parse(initData)
		if err != nil {
			s.log.Errorf(r.Context(), "failed to parse token: %v", err)
			s.writeError(r.Context(), w, Error{Code: http.StatusUnauthorized, Description: "Failed to parse token"})

			return
		}

		span.SetData("userID", fmt.Sprintf("%d", data.User.ID))

		ctx := context.WithValue(r.Context(), userIDKeyType, data.User.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (s *apiServer) authAdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span := sentry.StartSpan(
			r.Context(),
			"authAdminMiddleware",
		)
		defer span.Finish()

		userID, ok := r.Context().Value(userIDKeyType).(int64)
		if !ok {
			s.log.Errorf(r.Context(), "failed to get userID from context")
			s.writeError(r.Context(), w, Error{Code: http.StatusInternalServerError, Description: "Internal Server Error"})
			return
		}

		adminUserIDs := s.adminUserIDs
		isAdmin := false
		for _, id := range adminUserIDs {
			if id == userID {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			s.log.Errorf(r.Context(), "user %d is not admin", userID)
			s.writeError(r.Context(), w, Error{Code: http.StatusForbidden, Description: "Forbidden"})
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (s *apiServer) handleCompletedSurveys(w http.ResponseWriter, r *http.Request) {
	span := sentry.StartSpan(r.Context(), "handleCompletedSurveys")
	defer span.Finish()

	if r.Method != http.MethodGet {
		s.log.Errorf(r.Context(), "method not allowed")
		s.writeError(r.Context(), w, Error{Code: http.StatusMethodNotAllowed, Description: "Method Not Allowed"})

		return
	}

	userID, ok := r.Context().Value(userIDKeyType).(int64)
	if !ok {
		s.log.Errorf(r.Context(), "failed to get userID from context")
		s.writeError(r.Context(), w, Error{Code: http.StatusInternalServerError, Description: "Internal Server Error"})

		return
	}

	surveysResults, err := s.svc.GetCompletedSurveys(r.Context(), userID)
	if err != nil {
		s.log.Errorf(r.Context(), "failed to get completed surveys: %v", err)
		s.writeError(r.Context(), w, Error{Code: http.StatusInternalServerError, Description: "Internal Server Error"})

		return
	}

	var response CompletedSurveys
	for i, survey := range surveysResults {
		response.CompetetedSurveys = append(response.CompetetedSurveys, CompletedSurvey{
			ID:          fmt.Sprintf("%d", i),
			Name:        survey.SurveyName,
			Description: survey.Description,
			Results:     survey.Results.Text,
		})
	}

	// frontend expects an empty array if there are no completed surveys
	if response.CompetetedSurveys == nil {
		response.CompetetedSurveys = []CompletedSurvey{}
	}

	s.writeResponse(r.Context(), w, response)
}

func (s *apiServer) handleIsAdmin(w http.ResponseWriter, r *http.Request) {
	span := sentry.StartSpan(r.Context(), "handleIsAdmin")
	defer span.Finish()

	if r.Method != http.MethodGet {
		s.log.Errorf(r.Context(), "method not allowed")
		s.writeError(r.Context(), w, Error{Code: http.StatusMethodNotAllowed, Description: "Method Not Allowed"})

		return
	}

	userID, ok := r.Context().Value(userIDKeyType).(int64)
	if !ok {
		s.log.Errorf(r.Context(), "failed to get userID from context")
		s.writeError(r.Context(), w, Error{Code: http.StatusInternalServerError, Description: "Internal Server Error"})

		return
	}

	adminUserIDs := s.adminUserIDs
	isAdmin := false
	for _, id := range adminUserIDs {
		if id == userID {
			isAdmin = true
			break
		}
	}

	response := map[string]bool{"is_admin": isAdmin}

	s.writeResponse(r.Context(), w, response)
}

func (s *apiServer) handleUsersList(w http.ResponseWriter, r *http.Request) {
	span := sentry.StartSpan(r.Context(), "handleUsersList")
	defer span.Finish()

	if r.Method != http.MethodGet {
		s.log.Errorf(r.Context(), "method not allowed")
		s.writeError(r.Context(), w, Error{Code: http.StatusMethodNotAllowed, Description: "Method Not Allowed"})

		return
	}

	var (
		limit  = 10
		offset = 0
		search = ""
		err    error
	)

	if r.URL.Query().Get("limit") != "" {
		limit, err = strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			s.log.Errorf(r.Context(), "failed to parse limit: %v", err)
			s.writeError(r.Context(), w, Error{Code: http.StatusBadRequest, Description: "Bad Request"})

			return
		}
	}

	if r.URL.Query().Get("offset") != "" {
		offset, err = strconv.Atoi(r.URL.Query().Get("offset"))
		if err != nil {
			s.log.Errorf(r.Context(), "failed to parse offset: %v", err)
			s.writeError(r.Context(), w, Error{Code: http.StatusBadRequest, Description: "Bad Request"})

			return
		}
	}

	if r.URL.Query().Get("search") != "" {
		search = r.URL.Query().Get("search")
	}

	usersList, err := s.svc.GetUsersList(r.Context(), limit, offset, search)
	if err != nil {
		s.log.Errorf(r.Context(), "failed to get users list: %v", err)
		s.writeError(r.Context(), w, Error{Code: http.StatusInternalServerError, Description: "Internal Server Error"})

		return
	}

	s.writeResponse(r.Context(), w, usersList)
}

type CompletedSurveys struct {
	CompetetedSurveys []CompletedSurvey `json:"surveys"`
}

type CompletedSurvey struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Results     string `json:"results"`
}

type Error struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

func (s *apiServer) optionsResponse(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Origin", s.allowedOrigins)
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, sentry-trace, baggage")
}

func (s *apiServer) writeError(ctx context.Context, w http.ResponseWriter, err Error) {
	span := sentry.StartSpan(ctx, "writeError")
	defer span.Finish()

	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Origin", s.allowedOrigins)
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, sentry-trace, baggage")

	s.log.Errorf(ctx, "response error: %+v", err)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(err); err != nil {
		s.log.Errorf(ctx, "failed to write error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *apiServer) writeResponse(ctx context.Context, w http.ResponseWriter, data any) {
	span := sentry.StartSpan(ctx, "writeResponse")
	defer span.Finish()

	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Origin", s.allowedOrigins)
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, sentry-trace, baggage")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.log.Errorf(ctx, "failed to write response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
