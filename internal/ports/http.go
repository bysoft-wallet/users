package ports

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bysoft-wallet/users/internal/app"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type HttpServer struct {
	app *app.Application
}

const DEFAULT_PORT = "8088"

func NewHttpServer(app *app.Application) *HttpServer {
	return &HttpServer{app: app}
}

func (h HttpServer) Start() {
	port := DEFAULT_PORT
	if os.Getenv("APP_PORT") != "" {
		port = os.Getenv("APP_PORT")
	}

	r := chi.NewRouter()

	h.registerMiddlewares(r)
	h.registerRoutes(r)
	
	http.ListenAndServe(fmt.Sprintf(":%s", port), r)
}

func (h HttpServer) registerMiddlewares(r *chi.Mux) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(render.SetContentType(render.ContentTypeJSON))
}

func (h HttpServer) registerRoutes(r *chi.Mux) {
	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, map[string]string{"status": "ok"})
	})
}
