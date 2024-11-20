package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/blueorb/microservice/api"
	"github.com/blueorb/microservice/leader"
	"github.com/google/uuid"
)

type ElectionHandler struct {
	serviceName string
	serviceUID  uuid.UUID
}

func (h *ElectionHandler) StartLeading(ctx context.Context) {
	slog.Info("I'm the leader", "service", h.serviceName, "uid", h.serviceUID)
}

func (h *ElectionHandler) StopLeading() {
	slog.Info("I'm not the leader", "service", h.serviceName, "uid", h.serviceUID)
}

func (h *ElectionHandler) ElectedLeader(identity string) {
	slog.Info("Elected leader", "leader", identity, "service", h.serviceName, "uid", h.serviceUID)
}

func worker(
	serviceName string,
	serviceIdentity uuid.UUID,
) {
	for {
		if leader.IsLeader() {
			slog.Info("I'm the leader", "service", serviceName, "identity", serviceIdentity)
		}
		time.Sleep(time.Second * 5)
	}
}

func main() {
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelDebug)

	serviceUID := uuid.New()
	serviceName := "microservice"

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting the service", "service", serviceName, "uid", serviceUID)
	http.HandleFunc("/", index)
	http.HandleFunc("/api/echo", api.EchoHandleFunc)
	http.HandleFunc("/api/hello", api.HelloHandleFunc)
	http.HandleFunc("/api/books", api.BooksHandleFunc)
	http.HandleFunc("/is-alive", isAlive)
	http.HandleFunc("/is-ready", isReady)

	slog.Info("Starting the worker", "service", serviceName, "uid", serviceUID)
	go worker(serviceName, serviceUID)

	slog.Info("Starting the election", "service", serviceName, "uid", serviceUID)
	leader.RunElection(logger, serviceName, serviceUID, &ElectionHandler{serviceName: serviceName, serviceUID: serviceUID}, 15, 10, 2)

	slog.Info("Start listening", "service", serviceName, "uid", serviceUID, "port", port())
	http.ListenAndServe(port(), nil)
}

func port() string {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	return ":" + port
}

func index(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Welcome to Cloud Native Go.")
}

func isAlive(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
}

func isReady(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
}
