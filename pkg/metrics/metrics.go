package metrics

import (
	"context"
	"fmt"
	"github.com/hramov/dbouncer/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

var (
	statMap          = make(map[string]int)
	queriesProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "queries_processed_all",
		Help: "The total number of processed queries",
	})
)

func collectStat(ctx context.Context, statCh chan map[string]int) {
	for {
		select {
		case <-ctx.Done():
			return
		case stat, ok := <-statCh:
			if !ok {
				return
			}
			queriesAll := utils.MergeStatMapsAll(statMap, stat)

			if queriesAll > 0 {
				queriesProcessed.Add(float64(queriesAll))
			}
		}
	}
}

func Collect(ctx context.Context, port int, statCh chan map[string]int) {
	go collectStat(ctx, statCh)

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("metrics server started on %d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Printf("cannot start metrics server: %v\n", err)
		return
	}
}
