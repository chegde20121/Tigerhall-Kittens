package handlers

import (
	"net/http"

	_ "github.com/chegde20121/Tigerhall-Kittens/docs"
	"github.com/chegde20121/Tigerhall-Kittens/internal/app/service"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
)

func RegisterApiHandlers(sm *mux.Router) {
	postMethods := sm.Methods(http.MethodPost).Subrouter()
	postMethods.HandleFunc("/api/v1/register", NewUserHandler(logrus.New()).RegisterUsers)
	postMethods.HandleFunc("/api/v1/login", NewUserHandler(logrus.New()).Login)
	getMethods := sm.Methods(http.MethodGet).Subrouter()
	getMethods.HandleFunc("/api/v1/logout", NewUserHandler(logrus.New()).Logout)
	postMethods.HandleFunc("/api/v1/createTigers", service.AuthMiddleware(NewTigerHanlder(logrus.New()).CreateTiger))
	postMethods.HandleFunc("/api/v1/createSights", service.AuthMiddleware(NewSightingHandler(logrus.New()).CreateSight))
	getMethods.HandleFunc("/api/v1/listTigers", NewTigerHanlder(logrus.New()).ListTigers)
	getMethods.HandleFunc("/api/v1/tigers/{id}/listSightings", NewSightingHandler(logrus.New()).ListAllSightings)
	sm.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
}
