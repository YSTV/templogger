package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var (
		username, password, database, host, port, device, accessToken string
		secure                                                        bool
	)

	flag.StringVar(&username, "user", "", "InfluxDB username")
	flag.StringVar(&password, "pass", "", "InfluxDB password")
	flag.BoolVar(&secure, "secure", false, "InfluxDB use https?")
	flag.StringVar(&host, "host", "localhost", "InfluxDB hostname")
	flag.StringVar(&port, "port", "8086", "InfluxDB port")
	flag.StringVar(&database, "db", "", "InfluxDB database")
	flag.StringVar(&device, "device", "", "Particle device ID")
	flag.StringVar(&accessToken, "token", "", "Particle API access token")

	flag.Parse()

	if database == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	res, err := http.Get(fmt.Sprintf("https://api.particle.io/v1/devices/%s/temp?access_token=%s", device, accessToken))
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		log.Fatalf("Non-200 response from particle API: %s", http.StatusText(res.StatusCode))
	}

	var resJSON struct {
		Result float32
	}

	dec := json.NewDecoder(res.Body)
	if err = dec.Decode(&resJSON); err != nil {
		log.Fatal(err)
	}

	proto := "http"

	if secure {
		proto = "https"
	}

	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     fmt.Sprintf("%s://%s:%s", proto, host, port),
		Username: username,
		Password: password,
	})
	check(err)

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  database,
		Precision: "s",
	})
	check(err)

	tags := map[string]string{"location": "control-room-shelf"}
	fields := map[string]interface{}{
		"celcius": resJSON.Result,
	}

	pt, err := client.NewPoint("temperature", tags, fields, time.Now())
	check(err)
	bp.AddPoint(pt)

	check(c.Write(bp))
}
