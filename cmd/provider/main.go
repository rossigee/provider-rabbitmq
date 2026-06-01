/*
Copyright 2025 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"os"
	"path/filepath"
	goruntime "runtime"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"

	"github.com/rossigee/provider-rabbitmq/apis"
	"github.com/rossigee/provider-rabbitmq/internal/controller"
	"github.com/rossigee/provider-rabbitmq/internal/features"
	"github.com/rossigee/provider-rabbitmq/internal/health"
	"github.com/rossigee/provider-rabbitmq/internal/version"
)

func main() {
	var (
		app                      = kingpin.New(filepath.Base(os.Args[0]), "RabbitMQ Crossplane provider").DefaultEnvars()
		debug                    = app.Flag("debug", "Run with debug logging.").Short('d').Bool()
		pollInterval             = app.Flag("poll", "Poll interval controls how often an individual resource should be checked for drift.").Default("1m").Duration()
		leaderElection           = app.Flag("leader-election", "Use leader election for the controller manager.").Short('l').Default("false").OverrideDefaultFromEnvar("LEADER_ELECTION").Bool()
		maxReconcileRate         = app.Flag("max-reconcile-rate", "The global maximum rate per second at which resources may be checked for drift from the desired state.").Default("100").Int()
		metricsAddr              = app.Flag("metrics-addr", "The address the metrics endpoint binds to.").Default(":8080").String()
		healthAddr               = app.Flag("health-addr", "The address the health endpoint binds to.").Default(":8081").String()
		enableManagementPolicies = app.Flag("enable-management-policies", "Enable support for Management Policies.").Default("true").Bool()
	)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	zl := zap.New(zap.UseDevMode(*debug))
	log := logging.NewLogrLogger(zl.WithName("provider-rabbitmq"))

	if *debug {
		ctrl.SetLogger(zl)
	} else {
		ctrl.SetLogger(zl.WithValues("source", "controller-runtime").V(1))
	}

	log.Info("Provider starting up",
		"provider", "provider-rabbitmq",
		"version", version.Version,
		"go-version", goruntime.Version(),
		"platform", goruntime.GOOS+"/"+goruntime.GOARCH,
		"poll-interval", pollInterval.String(),
		"max-reconcile-rate", *maxReconcileRate,
		"leader-election", *leaderElection,
		"management-policies", *enableManagementPolicies,
	)

	ns, err := getWatchNamespace()
	kingpin.FatalIfError(err, "Cannot get watch namespace")

	cfg, err := ctrl.GetConfig()
	kingpin.FatalIfError(err, "Cannot get config")

	cacheOpts := cache.Options{}
	if ns != "" {
		cacheOpts.DefaultNamespaces = map[string]cache.Config{ns: {}}
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Cache:  cacheOpts,
		Metrics: server.Options{
			BindAddress: *metricsAddr,
		},
		LeaderElection:         *leaderElection,
		LeaderElectionID:       "crossplane-provider-rabbitmq",
		HealthProbeBindAddress: *healthAddr,
	})
	kingpin.FatalIfError(err, "Cannot create controller manager")

	kingpin.FatalIfError(apis.AddToScheme(mgr.GetScheme()), "Cannot add RabbitMQ APIs to scheme")

	featureFlags := &feature.Flags{}
	if *enableManagementPolicies {
		featureFlags.Enable(features.EnableAlphaManagementPolicies)
	}

	o := xpcontroller.Options{
		Logger:                  log,
		MaxConcurrentReconciles: *maxReconcileRate,
		PollInterval:            *pollInterval,
		GlobalRateLimiter:       ratelimiter.NewGlobal(*maxReconcileRate),
		Features:                featureFlags,
	}

	kingpin.FatalIfError(controller.Setup(mgr, o), "Cannot setup RabbitMQ controllers")

	healthChecker := health.NewHealthChecker(mgr.GetClient(), nil)
	kingpin.FatalIfError(mgr.AddHealthzCheck("rabbitmq-provider", healthChecker.HealthzCheck), "Cannot add healthz check")
	kingpin.FatalIfError(mgr.AddReadyzCheck("rabbitmq-provider", healthChecker.ReadyzCheck), "Cannot add readyz check")

	log.Info("Starting manager")
	kingpin.FatalIfError(mgr.Start(ctrl.SetupSignalHandler()), "Cannot start controller manager")
}

func getWatchNamespace() (string, error) {
	ns, _ := os.LookupEnv("WATCH_NAMESPACE")
	return ns, nil
}
