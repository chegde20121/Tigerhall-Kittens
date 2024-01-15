package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/chegde20121/Tigerhall-Kittens/internal/app/database"
	"github.com/chegde20121/Tigerhall-Kittens/internal/app/service"
	log "github.com/sirupsen/logrus"
)

type UserHandler struct {
	logger *log.Logger
}

func NewUserHandler(logger *log.Logger) *UserHandler {
	return &UserHandler{logger: logger}
}

func (uh *UserHandler) RegisterUsers(rw http.ResponseWriter, req *http.Request) {
	uh.logger.Info("Registering user.....")
	userService := service.NewUserService(uh.logger, database.GetDB())
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	userService.CreateUser(ctx, rw, req)
}

func (uh *UserHandler) Login(rw http.ResponseWriter, req *http.Request) {
	uh.logger.Info("Login user.....")
	userService := service.NewUserService(uh.logger, database.GetDB())
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	userService.Login(ctx, rw, req)
}

func (uh *UserHandler) Logout(rw http.ResponseWriter, req *http.Request) {
	uh.logger.Info("Log out.....")
	userService := service.NewUserService(uh.logger, database.GetDB())
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	userService.Logout(ctx, rw, req)
}
