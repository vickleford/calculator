package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vickleford/calculator/internal/store"
	"github.com/vickleford/calculator/internal/worker"
	"github.com/vickleford/calculator/internal/workqueue"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func main() {
	metricsAddr := flag.String("metrics", ":8081", "prometheus http metrics endpoint")
	etcdAddr := flag.String("etcdAddr", "localhost:2379", "etcd endpoints")
	rabbitAddr := flag.String("rmqAddr", "localhost:5672", "rabbitmq address")
	queueName := flag.String("queue", "calculations", "the workqueue name to use")
	flag.Parse()

	opts := workerOpts{
		metricsAddr: *metricsAddr,
		etcdAddr:    *etcdAddr,
		rabbitAddr:  *rabbitAddr,
		queueName:   *queueName,
	}

	log.Fatal(daemonize(opts))
}

type workerOpts struct {
	metricsAddr string
	etcdAddr    string
	rabbitAddr  string
	queueName   string
}

func (opts workerOpts) RabbitURL() string {
	user := os.Getenv("CALCULATORW_RABBIT_USER")
	pass := os.Getenv("CALCULATORW_RABBIT_PASS")
	return fmt.Sprintf("amqp://%s:%s@%s/", user, pass, opts.rabbitAddr)
}

func daemonize(opts workerOpts) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metricsRegistry := prometheus.NewRegistry()
	metricsRegistry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	workerErr := make(chan error)
	go func() {
		etcdClientMetrics := grpcprom.NewClientMetrics(
			grpcprom.WithClientCounterOptions(
				grpcprom.WithSubsystem("datastore_etcd_client"),
			),
		)
		metricsRegistry.MustRegister(etcdClientMetrics)

		etcdClient, err := clientv3.New(clientv3.Config{
			Endpoints:   strings.Split(opts.etcdAddr, ","),
			DialTimeout: 5 * time.Second,
			DialOptions: []grpc.DialOption{
				grpc.WithChainUnaryInterceptor(etcdClientMetrics.UnaryClientInterceptor()),
				grpc.WithChainStreamInterceptor(etcdClientMetrics.StreamClientInterceptor()),
			},
		})
		if err != nil {
			workerErr <- fmt.Errorf("error dialing etcd at %q: %w", opts.etcdAddr, err)
			return
		}
		defer etcdClient.Close()

		datastore := store.NewCalculationStore(etcdClient)

		rmqConn, err := amqp.Dial(opts.RabbitURL())
		if err != nil {
			workerErr <- fmt.Errorf("error dialing rabbitmq at %q: %w", opts.rabbitAddr, err)
			return
		}
		defer rmqConn.Close()

		fibonacciOfHandler := worker.NewFibOf(datastore)

		consumer := workqueue.NewConsumer(rmqConn,
			fibonacciOfHandler,
			workqueue.WithQueueName[workqueue.AMQP091Consumer](opts.queueName),
		)

		workerErr <- consumer.Start(ctx)
	}()

	metricsErr := make(chan error)
	go func() {
		handler := promhttp.HandlerFor(metricsRegistry,
			promhttp.HandlerOpts{
				Timeout: 30 * time.Second,
			},
		)
		metricsErr <- http.ListenAndServe(opts.metricsAddr, handler)
	}()

	log.Printf("running worker with metrics on %s...", opts.metricsAddr)

	select {
	case err := <-workerErr:
		return fmt.Errorf("error running worker: %w", err)
	case err := <-metricsErr:
		return fmt.Errorf("error running metrics service: %w", err)
	}
}
