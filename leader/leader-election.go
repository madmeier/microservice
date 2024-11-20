package leader

// leader selection based using kubernetes
// go get k8s.io/client-go@latest
// go get sigs.k8s.io/controller-runtime@latest

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	// lockName and lockNamespace need to be shared across all running instances
	lockNamespace = "blue-orb"
	leader        = false
)

type ElectionHandler interface {
	StartLeading(ctx context.Context)
	StopLeading()
	ElectedLeader(identity string)
}

func IsLeader() bool {
	return leader
}

// RunElection runs the leader election process.
// leaseDuration is the duration that non-leader candidates will wait to force acquire leadership (e.g. 15)
// renewDeadline is the duration that the acting leader will retry refreshing leadership before giving up (e.g. 10)
// retryPeriod is the duration the LeaderElector clients should wait between tries of actions (e.g. 2)
func RunElection(
	log *slog.Logger,
	serviceName string,
	serviceUID uuid.UUID,
	handler ElectionHandler,
	leaseDuration int,
	renewDeadline int,
	retryPeriod int,
) {
	lockName := serviceName + "-lock"
	serviceIdentity := serviceUID.String()

	cfg, err := ctrl.GetConfig()
	if err != nil {
		log.Error("failed to get config", "error", err)
	}
	if cfg == nil {
		log.Info("cfg is nil")
		handler.StartLeading(context.Background())
		leader = true
		return
	}

	log.Info("creating lease", "service", serviceName, "identity", serviceIdentity)
	// Create a new lock. This will be used to create a Lease resource in the cluster.
	l, err := resourcelock.NewFromKubeconfig(
		resourcelock.LeasesResourceLock,
		lockNamespace,
		lockName,
		resourcelock.ResourceLockConfig{
			Identity: serviceIdentity,
		},
		cfg,
		time.Second*time.Duration(leaseDuration),
	)
	if err != nil {
		panic(err)
	}
	if l == nil {
		log.Info("lock is nil", "service", serviceName, "identity", serviceIdentity)
	}

	log.Info("starting leader election", "service", serviceName, "identity", serviceIdentity)
	// Create a new leader election configuration with a 15 second lease duration.
	// Visit https://pkg.go.dev/k8s.io/client-go/tools/leaderelection#LeaderElectionConfig
	// for more information on the LeaderElectionConfig struct fields
	el, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          l,
		LeaseDuration: time.Second * time.Duration(leaseDuration), // the guarantee of the lease
		RenewDeadline: time.Second * time.Duration(renewDeadline), // waiting period before retries
		RetryPeriod:   time.Second * time.Duration(retryPeriod),   // retry if master fails
		Name:          lockName,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				leader = true
				handler.StartLeading(ctx)
			},
			OnStoppedLeading: func() {
				handler.StopLeading()
				leader = false
			},
			OnNewLeader: handler.ElectedLeader,
		},
	})
	if err != nil {
		panic(err)
	}

	// Begin the leader election process.
	// This will block.
	// returns when lead is lost
	go el.Run(context.Background())
	return
}
