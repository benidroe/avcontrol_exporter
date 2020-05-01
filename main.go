package main

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
	"strconv"

	//"os/signal"
	//"syscall"
	"time"
)
import "github.com/go-redis/redis"



func RedisNewClient() *redis.Client {

	db, _ := strconv.Atoi(*redisDB)
	client := redis.NewClient(&redis.Options{
		Addr:     *redisAddress,
		Password: *redisPassword, // no password
		DB:       db,  // use default DB
	})


	return client

}

var (
	listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":2113").String()
	udpListenAddress = kingpin.Flag("udp.listen-address", "Address to listen on for web interface and telemetry.").Default(":2114").String()
	redisAddress = kingpin.Flag("redis.address", "Address:Port of your redis-server.").Default("localhost:6379").String()
	redisPassword = kingpin.Flag("redis.password", "Address:Port of your redis-server.").Default("pass").String()
	redisDB = kingpin.Flag("redis.db", "Redis Database Number.").Default("0").String()
	//configFile    = kingpin.Flag("config.file", "Path to configuration file.").Default("pjlink.yml").String()
	logLevel      = kingpin.Flag("log.level", "LogLevel - Debug, Info, Warn, Error").Default("Debug").String()

	// Metrics about the PJLink exporter itself.
	pjlinkDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "avcontrol_collection_duration_seconds",
			Help: "Duration of collections by the PJLink exporter",
		},
		[]string{"module"},
	)
	pjlinkRequestErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "avcontrol_request_errors_total",
			Help: "Errors in requests to the avcontrol exporter",
		},
	)

	//config = Config{}

	reloadCh chan chan error
)

func init() {
	prometheus.MustRegister(pjlinkDuration)
	prometheus.MustRegister(pjlinkRequestErrors)
	prometheus.MustRegister(version.NewCollector("avcontrol_exporter"))
}

func handler(w http.ResponseWriter, r *http.Request, redisClient *redis.Client,  logger log.Logger) {
	query := r.URL.Query()

	target := query.Get("target")
	if len(query["target"]) != 1 || target == "" {
		http.Error(w, "'target' parameter must be specified once", 400)

		return
	}

	logger = log.With(logger, "target", target)
	level.Debug(logger).Log("msg", "Starting scrape")

	start := time.Now()
	registry := prometheus.NewRegistry()
	collector := collector{ctx: r.Context(), target: target, redisClient: redisClient, logger: logger}
	registry.MustRegister(collector)
	// Delegate http serving to Prometheus client library, which will call collector.Collect.
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	duration := time.Since(start).Seconds()
	pjlinkDuration.WithLabelValues("PJLink").Observe(duration)
	level.Debug(logger).Log("msg", "Finished scrape", "duration_seconds", duration)
}

func main() {

	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	level.Info(logger).Log("msg", "Starting avcontrol_exporter...")

	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	switch lol := *logLevel; lol {
	case "Debug":
		logger = level.NewFilter(logger, level.AllowDebug())
		level.Info(logger).Log("msg", "Starting with loglevel Debug")
	case "Info":
		logger = level.NewFilter(logger, level.AllowInfo())
		level.Info(logger).Log("msg", "Starting with loglevel Info")
	case "Warn":
		logger = level.NewFilter(logger, level.AllowWarn())
		level.Info(logger).Log("msg", "Starting with loglevel Warn")
	case "Error":
		logger = level.NewFilter(logger, level.AllowError())
		level.Info(logger).Log("msg", "Starting with loglevel Error")
	}

	redisClient := RedisNewClient()
	go UdpServer(context.Background(), *udpListenAddress, redisClient)	// Start UdpServer in a go routine


	/*if err := config.readConfig(*configFile); err != nil {
		level.Error(logger).Log("msg", "Error reloading config", "err", err)
	} else {
		level.Info(logger).Log("msg", "Loaded config file "+*configFile)
	}*/
	/* Reload Config not necessary yet
	hup := make(chan os.Signal, 1)
	reloadCh = make(chan chan error)
	signal.Notify(hup, syscall.SIGHUP)
	go func() {
		for {
			select {
			case <-hup:

				if err := config.readConfig(*configFile); err != nil {
					level.Error(logger).Log("msg", "Error reloading config", "err", err)
				} else {
					level.Info(logger).Log("msg", "Reloaded config file")
				}
			case rc := <-reloadCh:
				if err := config.readConfig(*configFile); err != nil {
					level.Error(logger).Log("msg", "Error reloading config", "err", err)
					rc <- err
				} else {
					level.Info(logger).Log("msg", "Reloaded config file")
					rc <- nil
				}
			}
		}
	}()

	*/

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/control", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, redisClient, logger)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
            <head>
            <title>AV Control Exporter</title>
            <style>
            label{
            display:inline-block;
            width:75px;
            }
            form label {
            margin: 10px;
            }
            form input {
            margin: 10px;
            }
            </style>
            </head>
            <body>
            <h1>AV Control Exporter</h1>
            <form action="/control">
            <label>Target:</label> <input type="text" name="target" placeholder="X.X.X.X" value="1.2.3.4"><br>
            <input type="submit" value="Submit">
            </form>
						<p><a href="/config">Config</a></p>
            </body>
            </html>`))
	})

	http.ListenAndServe(*listenAddress, nil)
	level.Info(logger).Log("msg", "Listen And server", "port", *listenAddress)

}