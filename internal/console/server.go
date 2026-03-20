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
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/go-playground/validator/v10"
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

	validator := validator.New()
	e.Validator = &model.CustomValidator{Validator: validator}

	initializeCockroachConn()
	cacheKeeper := newCacheKeeper()
	db := connect.CockroachDB

	// Initialize Cloudinary
	cld, err := cloudinary.NewFromParams(
		config.CloudinaryCloudName(),
		config.CloudinaryAPIKey(),
		config.CloudinaryAPISecret(),
	)
	if err != nil {
		log.Fatal("failed to initialize cloudinary: ", err)
	}

	userRepo := repository.NewUserRepository(db, cacheKeeper)
	sessionRepo := repository.NewSessionRepository(cacheKeeper)
	venueRepo := repository.NewVenueRepository(db)
	eventRepo := repository.NewEventRepository(db)
	eventOptionRepo := repository.NewEventOptionRepository(db)
	voteRepo := repository.NewVoteRepository(db)
	participantRepo := repository.NewParticipantRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	paymentRecordRepo := repository.NewPaymentRecordRepository(db)

	sessionUC := usecase.NewSessionUsecase(sessionRepo)
	userUC := usecase.NewUserUsecase(userRepo)
	venueUC := usecase.NewVenueUsecase(venueRepo)
	eventUC := usecase.NewEventUsecase(eventRepo)
	eventOptionUC := usecase.NewEventOptionUsecase(eventOptionRepo)
	voteUC := usecase.NewVoteUsecase(voteRepo)
	participantUC := usecase.NewParticipantUsecase(participantRepo)
	paymentUC := usecase.NewPaymentUsecase(paymentRepo, participantRepo)
	paymentRecordUC := usecase.NewPaymentRecordUsecase(paymentRecordRepo)
	uploadUC := usecase.NewUploadUsecase(cld)

	authCfg := model.AuthConfig{
		ClientID:     config.GoogleClientID(),
		ClientSecret: config.GoogleClientSecret(),
		RedirectURL:  config.GoogleRedirectURL(),
		Scopes:       []string{"openid", "email", "profile"},
	}
	authUC := usecase.NewAuthUsecase(authCfg, userRepo, sessionUC)

	apiHandler := delivery.NewAPIHandler(
		authUC,
		sessionUC,
		userUC,
		venueUC,
		eventUC,
		eventOptionUC,
		voteUC,
		participantUC,
		paymentUC,
		paymentRecordUC,
		uploadUC,
	)
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
