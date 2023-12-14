package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gotway/gotway/pkg/log"
	"github.com/gotway/gotway/pkg/metrics"

	"github.com/hamedetemaad/lineq-operator/internal/config"
	"github.com/hamedetemaad/lineq-operator/internal/runner"
	"github.com/hamedetemaad/lineq-operator/pkg/controller"
	wrv1alpha1clientset "github.com/hamedetemaad/lineq-operator/pkg/waitingroom/v1alpha1/apis/clientset/versioned"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	config, err := config.GetConfig()
	if err != nil {
		panic(fmt.Errorf("error getting config %v", err))
	}
	logger := getLogger(config)
	logger.Debugf("config %v", config)

	var restConfig *rest.Config
	var errKubeConfig error
	if config.KubeConfig != "" {
		restConfig, errKubeConfig = clientcmd.BuildConfigFromFlags("", config.KubeConfig)
	} else {
		restConfig, errKubeConfig = rest.InClusterConfig()
	}
	if errKubeConfig != nil {
		logger.Fatal("error getting kubernetes config ", err)
	}

	kubeClientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		logger.Fatal("error getting kubernetes client ", err)
	}
	wrv1alpha1ClientSet, err := wrv1alpha1clientset.NewForConfig(restConfig)
	if err != nil {
		logger.Fatal("error creating lineq client ", err)
	}

	ctrl := controller.New(
		kubeClientSet,
		wrv1alpha1ClientSet,
		config.Namespace,
		logger.WithField("type", "controller"),
	)

	ctx, cancel := signal.NotifyContext(context.Background(), []os.Signal{
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGKILL,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	}...)
	defer cancel()

	if config.Metrics.Enabled {
		m := metrics.New(
			metrics.Options{
				Path: config.Metrics.Path,
				Port: config.Metrics.Port,
			},
			logger.WithField("type", "metrics"),
		)
		go m.Start()
		defer m.Stop()
	}

	r := runner.NewRunner(
		ctrl,
		kubeClientSet,
		config,
		logger.WithField("type", "runner"),
	)
	r.Start(ctx)
}

func getLogger(config config.Config) log.Logger {
	logger := log.NewLogger(log.Fields{
		"service": "lineq-operator",
	}, config.Env, config.LogLevel, os.Stdout)
	if config.HA.Enabled {
		return logger.WithField("node", config.HA.NodeId)
	}
	return logger
}
