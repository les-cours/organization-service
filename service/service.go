package service

import (
	"github.com/les-cours/organization-service/api/orgs"
	"log"
	"net"
	"net/http"
	"runtime"

	"github.com/les-cours/organization-service/api/users"
	"github.com/les-cours/organization-service/resolvers"

	"github.com/les-cours/organization-service/database"
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

func monitoring_middleware(originalHandler http.Handler) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		memoryUsage.Set(float64(m.Alloc))
		goRoutineNum.Set(float64(runtime.NumGoroutine()))
		requestCounter.Inc()
		originalHandler.ServeHTTP(w, r)
	})
}

func Start() {
	registry.MustRegister(requestCounter, memoryUsage, goRoutineNum)
	promHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	http.HandleFunc("/metrics", monitoring_middleware(promHandler))
	log.Printf("Starting http server on port " + env.Settings.HttpPort)
	go func() {
		err := http.ListenAndServe(":"+env.Settings.HttpPort, nil)
		if err != nil {
			log.Fatalf("Error http server on port %v: %v", env.Settings.HttpPort, err)
		}
	}()

	lis, err := net.Listen("tcp", ":"+env.Settings.GrpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %v: %v", env.Settings.GrpcPort, err)
	}

	db, err := database.StartDatabase()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("user service connection: " + env.Settings.UserService.Host + ":" + env.Settings.UserService.Port)
	userConnectionService, err := grpc.Dial(env.Settings.UserService.Host+":"+env.Settings.UserService.Port, grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}
	defer userConnectionService.Close()

	userServiceClient := users.NewUserServiceClient(userConnectionService)

	var s = resolvers.GetInstance(db, userServiceClient)

	grpcServer := grpc.NewServer()
	orgs.RegisterOrgServiceServer(grpcServer, s)

	log.Printf("Starting ORGS grpc server on port " + env.Settings.GrpcPort)
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("Failed to start gRPC server on port %v: %v", env.Settings.GrpcPort, err)
	}

}
