package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/crowdmob/goamz/autoscaling"
	"github.com/crowdmob/goamz/aws"
	"github.com/jacobstr/confer"
	"github.com/marpaia/graphite-golang"
	"os"
	"strconv"
	"time"
)

// Passed in by build script
var (
	buildDate string
	version   string
)

var (
	config *Config
)

type Config struct {
	DebugMode bool
	Hostname  string
	AwsRegion aws.Region
	Graphite  *Graphite
}

type Graphite struct {
	Host      string
	Port      int
	Namespace string
}

func init() {

	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stderr)
	log.SetLevel(log.DebugLevel)

	app := confer.NewConfig()

	paths := []string{"config/application.yml"}

	if err := app.ReadPaths(paths...); err != nil {
		check_err(err)
	}

	// Set defaults
	app.SetDefault("app.graphite.port", 2003)
	app.SetDefault("app.graphite.name_space", "/autoscaling/logstash-indexer")

	config = &Config{
		AwsRegion: aws.USEast,
		DebugMode: app.GetBool("app.debug"),
		Hostname:  getHostname(),
		Graphite: &Graphite{
			Host:      app.GetString("app.graphite.host"),
			Port:      app.GetInt("app.graphite.port"),
			Namespace: app.GetString("app.graphite.name_space"),
		},
	}

	log.WithFields(log.Fields{
		"version":            version,
		"buildDate":          buildDate,
		"hostname":           config.Hostname,
		"aws_region":         config.AwsRegion.Name,
		"graphite_host":      config.Graphite.Host,
		"graphite_port":      config.Graphite.Port,
		"graphite_namespace": config.Graphite.Namespace,
	}).Debug("Config loaded")
}

func check_err(err error) {
	if err != nil {
		if config.DebugMode == true {
			log.Panic(err)
		} else {
			log.Error(err)
			os.Exit(2)
		}

	}
}
func getHostname() string {
	host, err := os.Hostname()
	check_err(err)
	return host
}

func main() {

	creds, err := aws.GetAuth("", "", "", time.Time{})
	check_err(err)

	g, err := graphite.NewGraphite(config.Graphite.Host, config.Graphite.Port)
	check_err(err)

	asg := autoscaling.New(creds, config.AwsRegion)
	resp, err := asg.DescribeAutoScalingGroups(nil)
	check_err(err)

	coordinates := fmt.Sprintf(config.Hostname + config.Graphite.Namespace)

	for _, i := range resp.AutoScalingGroups {
		if i.AutoScalingGroupName == "logstash-indexer-auto" {

			log.WithFields(log.Fields{
				"indexer_count":    strconv.Itoa(len(i.Instances)),
				"max_size":         strconv.FormatInt(i.MaxSize, 10),
				"min_size":         strconv.FormatInt(i.MinSize, 10),
				"desired_capacity": strconv.FormatInt(i.DesiredCapacity, 10),
				"hostname":         config.Hostname,
			}).Info("Sending Updates")

			g.SimpleSend(coordinates+"/indexer_count", strconv.Itoa(len(i.Instances)))
			g.SimpleSend(coordinates+"/max_size", strconv.FormatInt(i.MaxSize, 10))
			g.SimpleSend(coordinates+"/min_size", strconv.FormatInt(i.MinSize, 10))
			g.SimpleSend(coordinates+"/desired_capacity", strconv.FormatInt(i.DesiredCapacity, 10))

		}
		log.Debug("Update complete")
	}

}
