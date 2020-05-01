package main

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
	"regexp"
	"strconv"
	"time"
)


func matchString(pattern string, string string, logger log.Logger) bool {
	res, err := regexp.MatchString(pattern, string)

	if err != nil {
		level.Info(logger).Log("Error parsing MatchString", err)
	}
	return res
}

func extractTargetFromKey(key string, regex string, logger log.Logger) string {
	re := regexp.MustCompile(regex)
	match := re.FindAllStringSubmatch(key, -1)


	if 0 < len(match){
		if(1 < len(match[0])){
			return match[0][1]
		}
	}
	level.Info(logger).Log("Could not extract device from key", key)
	return "undefined"

}

type collector struct {
	ctx    context.Context
	target string
	redisClient *redis.Client
	logger log.Logger
}

// Describe implements Prometheus.Collector.
func (c collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

// Collect implements Prometheus.Collector.
func (c collector) Collect(ch chan<- prometheus.Metric) {


	// Todo: walkpjlink(c.target, c.pass, &pjSlice, c.logger)


	keys, err := c.redisClient.HGetAll(c.target).Result()
	if err != nil {
		panic(err)
	}

	level.Debug(c.logger).Log("collector", "here")

	for key, val := range keys{
		level.Debug(c.logger).Log(key, val)

		ival, _ := strconv.Atoi(val)

		switch{
		case matchString(`system.power.state`, key, c.logger):
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("avcontrol_system_power_state", "system power state", nil, nil),
				prometheus.GaugeValue, float64(ival))

		case matchString(`system.init`, key, c.logger):
			level.Debug(c.logger).Log("key", "uptime is matching")
			ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("avcontrol_system_uptime", "system uptime in seconds", nil, nil),
			prometheus.GaugeValue, float64(int(time.Now().Unix()) - ival))

		case matchString(`system.power.nightly`, key, c.logger):
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("avcontrol_system_power_nightly", "system nightly shutdown running", nil, nil),
				prometheus.GaugeValue, float64(ival))


		case matchString(`system.firealarm.state`, key, c.logger):
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("avcontrol_system_firealarm_state", "system firealarm mode and locked when value is 1", nil, nil),
				prometheus.GaugeValue, float64(ival))

		case matchString(`system.touchpanel.page`, key, c.logger):
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("avcontrol_system_touchpanel_page", "system selected touchpanel page", nil, nil),
			prometheus.GaugeValue, float64(ival))

		case matchString(`system.connected.[a-z0-9-.]+`, key, c.logger):
			target := extractTargetFromKey(key,`(?m)system\.connected\.(.*?)$`, c.logger)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("avcontrol_system_connected", "system perepherie item connected", []string{"device"}, nil),
				prometheus.GaugeValue, float64(ival), target)
				//Todo: Parse DNS-Name from key

		case matchString(`image.input.select.[a-z0-9-.]+`, key, c.logger):

			target := extractTargetFromKey(key,`(?m)image\.input\.select\.(.*?)$`, c.logger )

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("avcontrol_input_select", "system input selected for device", []string{"target"}, nil),
				prometheus.GaugeValue, float64(ival),target)
				// Todo: Parse targets Name from key

		}



	}

}