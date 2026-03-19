package console

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Back-Seat-Boy/kumpul-be/internal/config"
	"github.com/Back-Seat-Boy/kumpul-be/internal/delivery"
	"github.com/Back-Seat-Boy/kumpul-be/internal/model"
	"github.com/Back-Seat-Boy/kumpul-be/internal/repository"
	"github.com/Back-Seat-Boy/kumpul-be/internal/usecase"
	"github.com/kumparan/go-connect"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "server",
	Short: "run server",
	Long:  `This subcommand start the server`,
	Run:   run,
}

func init() {
	RootCmd.AddCommand(runCmd)
}

func run(_ *cobra.Command, _ []string) {
	e := echo.New()
	e.Pre(middleware.AddTrailingSlash())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	initializeCockroachConn()
	cacheKeeper := newCacheKeeper()
	db := connect.CockroachDB

	userRepo := repository.NewUserRepository(db, cacheKeeper)
	sessionRepo := repository.NewSessionRepository(cacheKeeper)

	sessionUC := usecase.NewSessionUsecase(sessionRepo)
	userUC := usecase.NewUserUsecase(userRepo)

	authCfg := model.AuthConfig{
		ClientID:     config.GoogleClientID(),
		ClientSecret: config.GoogleClientSecret(),
		RedirectURL:  config.GoogleRedirectURL(),
		Scopes:       []string{"openid", "email", "profile"},
	}
	authUC := usecase.NewAuthUsecase(authCfg, userRepo, sessionUC)

	apiHandler := delivery.NewAPIHandler(authUC, sessionUC, userUC)
	apiHandler.Routes(e)

	sigCh := make(chan os.Signal, 1)
	errCh := make(chan error, 1)
	quitCh := make(chan bool, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		for {
			select {
			case <-sigCh:
				log.Info("received an interrupt")
				gracefulShutdownHTTPServer(e)
				quitCh <- true
				return
			case err := <-errCh:
				log.Error(err)
				gracefulShutdownHTTPServer(e)
				quitCh <- true
			}
		}
	}()

	go func() {
		errCh <- e.Start(":" + config.Port())
	}()

	<-quitCh
	log.Info("shutdown")
}
