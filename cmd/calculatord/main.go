package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vickleford/calculator/internal/apiserver"
	"github.com/vickleford/calculator/internal/pb"
	"github.com/vickleford/calculator/internal/store"
	"github.com/vickleford/calculator/internal/workqueue"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func main() {
	listenAddr := flag.String("listen", ":8080", "gRPC server listen address")
	metricsAddr := flag.String("metrics", ":8081", "prometheus http metrics endpoint")
	etcdAddr := flag.String("etcdAddr", "localhost:2379", "etcd endpoints")
	rabbitAddr := flag.String("rmqAddr", "localhost:5672", "rabbitmq address")
	queueName := flag.String("queue", "calculations", "the workqueue name to use")
	flag.Parse()

	opts := daemonOpts{
		listenAddr:  *listenAddr,
		metricsAddr: *metricsAddr,
		etcdAddr:    *etcdAddr,
		rabbitAddr:  *rabbitAddr,
		queueName:   *queueName,
	}

	log.Fatal(daemonize(opts))
}

type daemonOpts struct {
	listenAddr  string
	metricsAddr string
	etcdAddr    string
	rabbitAddr  string
	queueName   string
}

func (opts daemonOpts) RabbitURL() string {
	user := os.Getenv("CALCULATORD_RABBIT_USER")
	pass := os.Getenv("CALCULATORD_RABBIT_PASS")
	return fmt.Sprintf("amqp://%s:%s@%s/", user, pass, opts.rabbitAddr)
}

func daemonize(opts daemonOpts) error {
	metricsRegistry := prometheus.NewRegistry()
	metricsRegistry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	listenErr := make(chan error)
	go func() {
		etcdClient, err := clientv3.New(clientv3.Config{
			Endpoints:   strings.Split(opts.etcdAddr, ","),
			DialTimeout: 5 * time.Second,
		})
		if err != nil {
			listenErr <- fmt.Errorf("error dialing etcd at %q: %w", opts.etcdAddr, err)
			return
		}
		defer etcdClient.Close()

		datastore := store.NewCalculationStore(etcdClient)

		rmqConn, err := amqp.Dial(opts.RabbitURL())
		if err != nil {
			listenErr <- fmt.Errorf("error dialing rabbitmq at %q: %w", opts.rabbitAddr, err)
			return
		}
		defer rmqConn.Close()

		producer := workqueue.NewProducer(rmqConn,
			workqueue.WithQueueName[workqueue.Producer](opts.queueName))

		listener, err := net.Listen("tcp", opts.listenAddr)
		if err != nil {
			listenErr <- err
			return
		}

		preferredBuckets := []float64{
			0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120,
		}

		srvMetrics := grpcprom.NewServerMetrics(
			grpcprom.WithServerHandlingTimeHistogram(
				grpcprom.WithHistogramBuckets(preferredBuckets),
			),
		)
		metricsRegistry.MustRegister(srvMetrics)

		grpcOpts := []grpc.ServerOption{
			grpc.ChainUnaryInterceptor(srvMetrics.UnaryServerInterceptor()),
		}
		gRPCServer := grpc.NewServer(grpcOpts...)
		pb.RegisterCalculationsServer(gRPCServer, apiserver.NewCalculations(datastore, producer))

		listenErr <- gRPCServer.Serve(listener)
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

	log.Printf("running on %s with metrics on %s...", opts.listenAddr, opts.metricsAddr)

	select {
	case err := <-listenErr:
		return fmt.Errorf("error running gRPC server: %w", err)
	case err := <-metricsErr:
		return fmt.Errorf("error running metrics service: %w", err)
	}
}
