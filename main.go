package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type Client struct {
	v1api  v1.API
}

func NewClient() *Client {
	client, err := api.NewClient(api.Config{
		Address: os.Getenv("URL"),
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	v1api := v1.NewAPI(client)
	return &Client{v1api}
}

func main() {

	os.Remove("./metrics_cpu.csv")

	_, err := os.Create("./metrics_cpu.csv")
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		os.Exit(1)
	}

	client := NewClient()

	// Used for debug specific time range
	//var loc = time.Now().Local().Location()
	//var t time.Time = time.Date(2022, time.January, 15, 0, 0, 0, 0, loc)
	var t time.Time = time.Now()
	
	for i := 0; i < 365; i++ {
		for j := 0; j < 4; j++ {

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			result, err := client.getMetric(t, ctx)
			if err != nil {
				fmt.Printf("Error querying Prometheus: %v\n", err)
				os.Exit(1)
			}
			//fmt.Printf("Result:\n%v\n", result)
			if result.String() != "" {
				//fmt.Printf("debug: %v\n", result)
				err = printRespCSV(result)
				if err != nil {
					fmt.Printf("Error printing CSV: %v\n", err)
					os.Exit(1)
				}
			}


			//need to not full cache vmselect
			time.Sleep(1 * time.Second)
			// remove 1 hour
			t = t.Add(-time.Hour * 6)

		}
	}

}

func (c *Client) getMetric(t time.Time, ctx context.Context) (model.Value, error) {

	fmt.Printf("time: %v\n", t)
	r := v1.Range{
		Start: t.Add(-time.Hour * 6),
		End:   t,
		Step:  time.Minute,
	}

	result, warnings, err := c.v1api.QueryRange(ctx, "sum(irate(node_cpu_seconds_total{mode=~\"user|system\", job=\"node-exporter\"}[1h30s]))", r, v1.WithTimeout(120*time.Second))
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}

	ctx.Done()
	return result, nil

}

func printRespCSV(result model.Value) (error) {
	var err error

	f, err := os.OpenFile("./metrics_cpu.csv", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return err
	}

	mapData := make(map[model.Time]model.SampleValue)
	for _, val := range result.(model.Matrix)[0].Values {
		mapData[val.Timestamp] = val.Value
	}
	for t, v := range mapData {
		// append in file f
		_, err = fmt.Fprintf(f, "%v;%v\n", t, v)
		if err != nil {
			fmt.Println(err)
			f.Close()
			return err
		}
	}
	
	//fmt.Println("file appended successfully")
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
