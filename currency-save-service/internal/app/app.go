package app

import (
	"context"
	_ "currency-save-service/docs"
	"currency-save-service/internal/config"
	"currency-save-service/internal/generated"
	"currency-save-service/internal/handlers"
	"currency-save-service/internal/loger"
	"currency-save-service/internal/metrics"
	"currency-save-service/internal/repository"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Application struct {
	grpcServer *grpc.Server
	httpServer *http.Server
}

func NewApplication() *Application {
	return &Application{}
}

func (a *Application) StartServer() {
	cfg := config.Config{}
	loger := loger.NewLogerr()

	cfg.LoadEnv()

	appPort := cfg.GetEnv("PORT", "8080")
	metricsPort := cfg.GetEnv("METRICS_PORT", "2112")
	connectionString := cfg.GetEnv("DB_CONNECTION", "")
	grpcPort := cfg.GetEnv("GRPC_PORT", "9090")

	metrics := metrics.NewMetrics()
	repo := repository.NewRepository(connectionString, loger, metrics)
	hand := handlers.NewHandler(repo, loger)

	r := mux.NewRouter()

	r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)).Methods(http.MethodGet)

	r.HandleFunc("/currency/save/{date}", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithDeadline(r.Context(), time.Now().Add(30*time.Second))
		defer cancel()

		hand.SaveCurrencyHandler(w, r.WithContext(ctx), ctx)
	})

	r.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(":"+metricsPort, r); err != nil {
			fmt.Println("Failed to start the metrics server:", err)
		}
	}()

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	a.grpcServer = grpc.NewServer()
	generated.RegisterCurrencyServiceServer(a.grpcServer, &handlers.CurrencyServiceServer{Repo: repo})
	reflection.Register(a.grpcServer)

	go func() {
		if err := a.grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	server := &http.Server{
		Addr:         ":" + appPort,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      r,
	}

	log.Fatal(server.ListenAndServe())
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server shutdown failed: %v", err)
	}

	a.grpcServer.GracefulStop()
	log.Println("Servers stopped gracefully")

}

func shutdown(quit chan os.Signal, logger slog.Logger) {
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	s := <-quit
	logger.Info("caught signal",
		"signal", s.String(),
	)
	os.Exit(0)
}
