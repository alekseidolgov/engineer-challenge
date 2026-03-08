package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexdolgov/auth-service/internal/config"
	"github.com/alexdolgov/auth-service/internal/identity/application/command"
	"github.com/alexdolgov/auth-service/internal/identity/infrastructure/hasher"
	"github.com/alexdolgov/auth-service/internal/identity/infrastructure/outbox"
	"github.com/alexdolgov/auth-service/internal/identity/infrastructure/postgres"
	"github.com/alexdolgov/auth-service/internal/identity/infrastructure/ratelimit"
	"github.com/alexdolgov/auth-service/internal/identity/infrastructure/token"
	authgrpc "github.com/alexdolgov/auth-service/internal/identity/port/grpc"
	pb "github.com/alexdolgov/auth-service/internal/identity/port/grpc/pb/auth/v1"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(config.DBMaxOpenConns)
	db.SetMaxIdleConns(config.DBMaxIdleConns)
	db.SetConnMaxLifetime(config.DBConnMaxLifetime)

	if err := db.Ping(); err != nil {
		logger.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	logger.Info("database connected")

	userRepo := postgres.NewUserRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)
	resetRepo := postgres.NewResetTokenRepository(db)

	argon2Hasher := hasher.NewArgon2Hasher()

	// В production ключ загружается из Vault/KMS/файла, а не генерируется при старте.
	privateKey, err := token.GenerateKey()
	if err != nil {
		logger.Error("failed to generate JWT signing key", "error", err)
		os.Exit(1)
	}
	jwtIssuer := token.NewJWTIssuer(privateKey, cfg.JWTTTL)
	logger.Info("JWT issuer initialized", "algorithm", "ES256")

	mockOutbox := outbox.NewMockOutbox(logger)

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		logger.Error("failed to parse redis URL", "error", err)
		os.Exit(1)
	}
	redisClient := redis.NewClient(redisOpts)
	defer redisClient.Close()
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	logger.Info("redis connected")

	loginLimiter := ratelimit.NewRedisLimiter(redisClient, cfg.RateLimitLogin, cfg.RateLimitWindow, "login")
	resetLimiter := ratelimit.NewRedisLimiter(redisClient, cfg.RateLimitReset, cfg.RateLimitWindow, "reset")

	registerHandler := command.NewRegisterHandler(userRepo, userRepo, argon2Hasher, mockOutbox)
	loginHandler := command.NewLoginHandler(userRepo, sessionRepo, argon2Hasher, jwtIssuer)
	logoutHandler := command.NewLogoutHandler(sessionRepo)
	requestResetHandler := command.NewRequestResetHandler(userRepo, resetRepo, resetRepo, mockOutbox)
	confirmResetHandler := command.NewConfirmResetHandler(userRepo, userRepo, resetRepo, resetRepo, argon2Hasher, mockOutbox)
	refreshTokenHandler := command.NewRefreshTokenHandler(sessionRepo, sessionRepo, userRepo, jwtIssuer)

	authServer := authgrpc.NewAuthServer(
		registerHandler, loginHandler, logoutHandler,
		requestResetHandler, confirmResetHandler,
		refreshTokenHandler,
		loginLimiter, resetLimiter,
	)

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(authgrpc.LoggingInterceptor(logger)),
	)
	pb.RegisterAuthServiceServer(srv, authServer)

	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("auth.v1.AuthService", healthpb.HealthCheckResponse_SERVING)

	reflection.Register(srv)

	lis, err := net.Listen("tcp", cfg.GRPCAddr())
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	go func() {
		logger.Info(fmt.Sprintf("gRPC server listening on %s", cfg.GRPCAddr()))
		if err := srv.Serve(lis); err != nil {
			logger.Error("gRPC server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down gracefully")
	srv.GracefulStop()
}
