package ports

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bysoft-wallet/users/internal/app"
	apperrors "github.com/bysoft-wallet/users/internal/app/errors"
	"github.com/bysoft-wallet/users/internal/app/jwt"
	"github.com/bysoft-wallet/users/internal/app/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type HttpServer struct {
	app          *app.Application
	accessHeader string
	validator    *validator.Validate
}

const DEFAULT_PORT = "8088"

func NewHttpServer(app *app.Application, accessHeader string) *HttpServer {
	validate := validator.New()
	return &HttpServer{
		app:          app,
		accessHeader: accessHeader,
		validator:    validate,
	}
}

func (h *HttpServer) Start() {
	port := DEFAULT_PORT
	if os.Getenv("APP_PORT") != "" {
		port = os.Getenv("APP_PORT")
	}

	r := chi.NewRouter()

	h.registerMiddlewares(r)
	h.registerRoutes(r)

	http.ListenAndServe(fmt.Sprintf(":%s", port), r)
}

func (h *HttpServer) registerMiddlewares(r *chi.Mux) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(render.SetContentType(render.ContentTypeJSON))
}

func (h *HttpServer) registerRoutes(r *chi.Mux) {
	r.Route("/users/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			render.JSON(w, r, map[string]string{"status": "ok"})
		})

		r.Post("/signIn", h.signIn)
		r.Post("/signUp", h.signUp)
		r.Post("/refresh", h.refresh)

		r.Get("/me", h.me)
		r.Put("/settings", h.updateSettings)
	})
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=5"`
}

type RefreshRequest struct {
	Refresh string `json:"refresh" validate:"required"`
}

type SignUpRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=5"`
	Name     string `json:"name" validate:"required,gte=1"`
}

type TokenPairResponse struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

type UserResponse struct {
	UUID     uuid.UUID       `json:"uuid"`
	Email    string          `json:"email"`
	Name     string          `json:"name"`
	Settings SettingsPayload `json:"settings"`
}

type SettingsPayload struct {
	Currency string `json:"currency" validate:"required"`
}

func (e *TokenPairResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, 200)
	return nil
}

func (e *UserResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, 200)
	return nil
}

func (h *HttpServer) signIn(w http.ResponseWriter, r *http.Request) {
	log.Printf("IP %v", r.RemoteAddr)
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body) // response body is []byte
	if err != nil {
		BadRequest("invalid-input", err, w, r)
		return
	}

	var request LoginRequest
	if err := json.Unmarshal(body, &request); err != nil {
		BadRequest("invalid-input", err, w, r)
		return
	}

	serviceRequest := &service.SignInRequest{
		Email:    request.Email,
		Password: request.Password,
		Ip:       r.RemoteAddr,
	}

	tokens, err := h.app.AuthService.SignIn(r.Context(), serviceRequest)
	if err != nil {
		RespondWithAppError(err, w, r)
		return
	}

	render.Render(w, r, &TokenPairResponse{
		Access:  tokens.Access.Token,
		Refresh: tokens.Refresh.Token,
	})
}

func (h *HttpServer) RespondValidationError(errs []validator.FieldError, w http.ResponseWriter, r *http.Request) {
	for _, err := range errs {
		fmt.Println(err.Namespace())
		fmt.Println(err.Field())
		fmt.Println(err.StructNamespace())
		fmt.Println(err.StructField())
		fmt.Println(err.Tag())
		fmt.Println(err.ActualTag())
		fmt.Println(err.Kind())
		fmt.Println(err.Type())
		fmt.Println(err.Value())
		fmt.Println(err.Param())
	}

	err := errs[0]

	slug := "invalid-input"
	if err.Field() == "Password" && err.Tag() == "gte" {
		slug = "field-password-invalid-length"
	} else if err.Field() == "Email" && err.Tag() == "email" {
		slug = "field-email-invalid"
	} else if err.Field() == "Name" && err.Tag() == "gte" {
		slug = "field-name-invalid-length"
	} else if err.Field() == "Name" && err.Tag() == "required" {
		slug = "field-name-required"
	} else if err.Field() == "Password" && err.Tag() == "required" {
		slug = "field-password-required"
	} else if err.Field() == "Email" && err.Tag() == "required" {
		slug = "field-email-required"
	} else if err.Field() == "Refresh" && err.Tag() == "required" {
		slug = "invalid-token"
	} else if err.Field() == "Currency" && err.Tag() == "required" {
		slug = "field-currency-required"
	}

	BadRequest(slug, err, w, r)
}

func (h *HttpServer) signUp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body) // response body is []byte
	if err != nil {
		BadRequest("invalid-input", err, w, r)
		return
	}

	var request SignUpRequest
	if err := json.Unmarshal(body, &request); err != nil {
		BadRequest("invalid-input", err, w, r)
		return
	}

	err = h.validator.Struct(request)
	if err != nil {
		h.RespondValidationError(err.(validator.ValidationErrors), w, r)
		return
	}

	serviceRequest := &service.SignUpRequest{
		Email:    request.Email,
		Password: request.Password,
		Name:     request.Name,
		Ip:       r.RemoteAddr,
	}

	tokens, err := h.app.AuthService.SignUp(r.Context(), serviceRequest)
	if err != nil {
		RespondWithAppError(err, w, r)
		return
	}

	render.Render(w, r, &TokenPairResponse{
		Access:  tokens.Access.Token,
		Refresh: tokens.Refresh.Token,
	})
}

func (h *HttpServer) updateSettings(w http.ResponseWriter, r *http.Request) {
	access, err := h.getAccessFromHeader(w, r)
	if err != nil {
		Unauthorised("invalid-token", err, w, r)
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body) // response body is []byte
	if err != nil {
		BadRequest("invalid-input", err, w, r)
		return
	}

	var payload SettingsPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		BadRequest("invalid-input", err, w, r)
		return
	}

	err = h.validator.Struct(payload)
	if err != nil {
		h.RespondValidationError(err.(validator.ValidationErrors), w, r)
		return
	}

	serviceRequest := service.UpdateSettingsRequest{
		Currency: payload.Currency,
		UserUUID: access.Claims.UserId,
	}

	user, err := h.app.AuthService.UpdateSettings(r.Context(), &serviceRequest)
	if err != nil {
		RespondWithAppError(err, w, r)
		return
	}

	render.Render(w, r, &UserResponse{
		UUID:  user.UUID,
		Email: user.Email,
		Name:  user.Name,
		Settings: SettingsPayload{
			Currency: user.Settings.Currency.String(),
		},
	})
}

func (h *HttpServer) me(w http.ResponseWriter, r *http.Request) {
	access, err := h.getAccessFromHeader(w, r)
	if err != nil {
		Unauthorised("invalid-token", err, w, r)
		return
	}

	user, err := h.app.AuthService.GetUser(r.Context(), access.Claims.UserId)
	if err != nil {
		RespondWithAppError(err, w, r)
		return
	}

	render.Render(w, r, &UserResponse{
		UUID:  user.UUID,
		Email: user.Email,
		Name:  user.Name,
		Settings: SettingsPayload{
			Currency: user.Settings.Currency.String(),
		},
	})
}

func (h *HttpServer) refresh(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body) // response body is []byte
	if err != nil {
		BadRequest("invalid-input", err, w, r)
		return
	}

	var request RefreshRequest
	if err := json.Unmarshal(body, &request); err != nil {
		BadRequest("invalid-input", err, w, r)
		return
	}

	err = h.validator.Struct(request)
	if err != nil {
		h.RespondValidationError(err.(validator.ValidationErrors), w, r)
		return
	}

	tokens, err := h.app.AuthService.Refresh(r.Context(), request.Refresh, r.RemoteAddr)
	if err != nil {
		RespondWithAppError(err, w, r)
		return
	}

	render.Render(w, r, &TokenPairResponse{
		Access:  tokens.Access.Token,
		Refresh: tokens.Refresh.Token,
	})
}

func (h *HttpServer) getAccessFromHeader(w http.ResponseWriter, r *http.Request) (*jwt.AccessJWT, error) {
	tokenHeader := r.Header.Get("X-API-Token")

	access, err := h.app.JWTService.ValidateAccess(tokenHeader)
	if err != nil {
		return &jwt.AccessJWT{}, err
	}

	return access, nil
}

func InternalError(slug string, err error, w http.ResponseWriter, r *http.Request) {
	httpRespondWithError(err, slug, w, r, "Internal server error", http.StatusInternalServerError)
}

func Unauthorised(slug string, err error, w http.ResponseWriter, r *http.Request) {
	log.Printf("Unathorized error %s", slug)
	httpRespondWithError(err, slug, w, r, "Unauthorised", http.StatusUnauthorized)
}

func BadRequest(slug string, err error, w http.ResponseWriter, r *http.Request) {
	httpRespondWithError(err, slug, w, r, "Bad request", http.StatusBadRequest)
}

func NotFound(slug string, err error, w http.ResponseWriter, r *http.Request) {
	httpRespondWithError(err, slug, w, r, "Not found", http.StatusNotFound)
}

func RespondWithAppError(err error, w http.ResponseWriter, r *http.Request) {
	appError, ok := err.(apperrors.AppError)
	if !ok {
		InternalError("internal-server-error", err, w, r)
		return
	}

	switch appError.ErrorType() {
	case apperrors.ErrorTypeAuthorization:
		Unauthorised(appError.Slug(), appError, w, r)
	case apperrors.ErrorTypeIncorrectInput:
		BadRequest(appError.Slug(), appError, w, r)
	case apperrors.ErrorNotFound:
		NotFound(appError.Slug(), appError, w, r)
	default:
		InternalError(appError.Slug(), appError, w, r)
	}
}

func httpRespondWithError(err error, slug string, w http.ResponseWriter, r *http.Request, logMSg string, status int) {
	log.Printf("HTTP Request error %v", err)
	resp := ErrorResponse{slug, status}

	if err := render.Render(w, r, resp); err != nil {
		panic(err)
	}
}

type ErrorResponse struct {
	Slug       string `json:"slug"`
	httpStatus int
}

func (e ErrorResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.httpStatus)
	return nil
}
