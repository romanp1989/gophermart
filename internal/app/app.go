package app

import (
	"github.com/go-chi/chi/v5"
	accrualSer "github.com/romanp1989/gophermart/internal/accrual"
	"github.com/romanp1989/gophermart/internal/api/balance"
	"github.com/romanp1989/gophermart/internal/api/order"
	"github.com/romanp1989/gophermart/internal/api/user"
	balanceSer "github.com/romanp1989/gophermart/internal/balance"
	"github.com/romanp1989/gophermart/internal/config"
	"github.com/romanp1989/gophermart/internal/database"
	"github.com/romanp1989/gophermart/internal/logger"
	middlewares2 "github.com/romanp1989/gophermart/internal/middlewares"
	orderSer "github.com/romanp1989/gophermart/internal/order"
	"github.com/romanp1989/gophermart/internal/server"
	userSer "github.com/romanp1989/gophermart/internal/user"
	"go.uber.org/zap"
	"log"
	"os"
	"time"
)

type handlerFunc func(mux *chi.Mux)

type App struct {
	route *chi.Mux
	log   *zap.Logger
}

func NewApp() (*App, error) {
	zapLogger, err := logger.Initialize("info")
	if err != nil {
		log.Printf("can't initalize logger %s", err)
		return nil, err
	}

	defer func(zLog *zap.Logger) {
		_ = zLog.Sync()
	}(zapLogger)

	err = config.NewConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	db, err := database.NewDB(&database.Config{
		Dsn:             config.Options.FlagDBDsn,
		MaxIdleConn:     1,
		MaxOpenConn:     10,
		MaxLifetimeConn: time.Minute * 1,
	})
	if err != nil {
		zapLogger.Fatal("Database init error: ", zap.String("error", err.Error()))
	}

	userRepository := userSer.NewDBStorage(db)
	userService := userSer.NewService(userRepository, zapLogger)
	userHandler := user.NewUserHandler(userService, zapLogger)

	orderRepository := orderSer.NewDBStorage(db)
	orderValidator := orderSer.NewValidator(orderRepository)
	orderService := orderSer.NewService(orderRepository, orderValidator, zapLogger)
	orderHandler := order.NewOrderHandler(orderService, zapLogger)

	balanceRepository := balanceSer.NewDBStorage(db)
	balanceService := balanceSer.NewService(balanceRepository, orderValidator, zapLogger)
	balanceHandler := balance.NewBalanceHandler(balanceService, zapLogger)

	middlewares := middlewares2.New(config.Options.FlagSecretKey)
	route := server.NewRoutes(userHandler, orderHandler, balanceHandler, middlewares)

	accrualRepository := accrualSer.NewDBStorage(db)
	accrualService := accrualSer.NewService(accrualRepository, zapLogger)

	//Запускаем метод  прослушивания Accrual сервиса
	accrualService.OrderStatusChecker(config.Options.FlagAccrualDuration)

	return &App{
		route: route,
		log:   zapLogger,
	}, nil
}

func (a *App) Run() int {
	httpServer := server.NewServer(a.route, a.log)

	errChannel := make(chan error, 1)
	oss, stop := make(chan os.Signal, 1), make(chan struct{}, 1)

	go func() {
		<-oss

		stop <- struct{}{}
	}()

	go func() {
		if err := httpServer.RunServer(); err != nil {
			errChannel <- err
		}
	}()

	for {
		select {
		case err := <-errChannel:
			a.log.Warn("Can't run application", zap.Error(err))
			return 0
		case <-stop:
			httpServer.Stop()
			return 0
		}
	}
}
