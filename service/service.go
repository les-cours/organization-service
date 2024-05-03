package service

import (
	"database/sql"
	"github.com/les-cours/organization-service/api/orgs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net"
	"net/http"
	"os"
	"runtime"

	"github.com/les-cours/organization-service/api/users"
	"github.com/les-cours/organization-service/resolvers"

	"github.com/les-cours/organization-service/env"
	"google.golang.org/grpc"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	registry       = prometheus.NewRegistry()
	requestCounter = prometheus.NewGauge(prometheus.GaugeOpts{Name: "request_counter", Help: "request counter"})
	memoryUsage    = prometheus.NewGauge(prometheus.GaugeOpts{Name: "memory_usage", Help: "memory usage"})
	goRoutineNum   = prometheus.NewGauge(prometheus.GaugeOpts{Name: "go_routines_num", Help: "the number of go routine "})
)

func monitoringMiddleware(originalHandler http.Handler) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		memoryUsage.Set(float64(m.Alloc))
		goRoutineNum.Set(float64(runtime.NumGoroutine()))
		requestCounter.Inc()
		originalHandler.ServeHTTP(w, r)
	}
}

func loggerInit() *zap.Logger {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(zap.InfoLevel),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0))
	return logger
}
func Start() {
	/*
	  INIT LOGGER ...
	*/
	logger := loggerInit()
	defer logger.Sync()

	/*
	  REGISTRY ...
	*/
	registry.MustRegister(requestCounter, memoryUsage, goRoutineNum)
	promHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	http.HandleFunc("/metrics", monitoringMiddleware(promHandler))
	logger.Info("Starting http server on port " + env.Settings.HttpPort)
	go func() {
		err := http.ListenAndServe(":"+env.Settings.HttpPort, nil)
		if err != nil {
			logger.Error(err.Error())
		}
	}()

	lis, err := net.Listen("tcp", ":"+env.Settings.GrpcPort)
	if err != nil {
		logger.Error(err.Error())
	}
	/*
	   DATABASE
	*/
	var db *sql.DB
	db, err = StartDatabase()
	if err != nil {
		logger.Error(err.Error())
	}

	/*
	   USER-SERVICE CONNECTION ...
	*/
	var userClientConn *grpc.ClientConn
	userClientConn, err = grpc.Dial(env.Settings.UserService.Host+":"+env.Settings.UserService.Port, grpc.WithInsecure())
	if err != nil {
		logger.Error(err.Error())
	}
	defer userClientConn.Close()

	userServiceClient := users.NewUserServiceClient(userClientConn)

	/*
	   CREATE SERVER
	*/
	var s = resolvers.GetInstance(db, userServiceClient, logger)

	grpcServer := grpc.NewServer()
	orgs.RegisterOrgServiceServer(grpcServer, s)

	logger.Info("Starting ORGS grpc server on port " + env.Settings.GrpcPort)
	err = grpcServer.Serve(lis)
	if err != nil {
		logger.Error(err.Error())
	}

}
