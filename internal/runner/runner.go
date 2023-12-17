package runner

import (
	"context"

	"github.com/gotway/gotway/pkg/log"
	"github.com/hamedetemaad/lineq-operator/internal/config"
	"github.com/hamedetemaad/lineq-operator/pkg/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

type Runner struct {
	ctrl      *controller.Controller
	clientset *kubernetes.Clientset
	config    config.Config
	logger    log.Logger
}

func (r *Runner) Start(ctx context.Context) {
	if r.config.HA.Enabled {
		r.logger.Info("starting HA controller")
		r.runHA(ctx)
	} else {
		r.logger.Info("starting standalone controller")
		r.runSingleNode(ctx)
	}
}

func (r *Runner) runSingleNode(ctx context.Context) {
	if err := r.ctrl.Run(ctx, r.config.NumWorkers); err != nil {
		r.logger.Fatal("error running controller ", err)
	}
}

func (r *Runner) initCfg() {
	cm, _ := r.clientset.CoreV1().ConfigMaps("haproxy-controller").Get(context.TODO(), "haproxy-kubernetes-ingress", metav1.GetOptions{})

	if _, exists := cm.Annotations["initialized"]; exists {
		return
	}

	cm.Annotations["initialized"] = "true"

	config := `
http-request set-var(txn.vwr_path) var(txn.host),concat('.vwr',txn.path),map(/etc/haproxy/maps/path-exact.map)
use_backend %[var(txn.path_match),field(1,.)] if !{ var(txn.vwr_path) -m found }
http-request set-var(txn.has_cookie) req.cook_cnt(sessionid)
http-request set-var(txn.t2) uuid()  if !{ var(txn.has_cookie) -m int gt 0 }
http-request set-var(txn.sessionid) req.cook(sessionid)
http-request set-var(txn.index) var(txn.host),regsub(\.,_,g),concat(,txn.path,),regsub(\/,_,g)
http-request track-sc0 var(txn.index) table room
http-response add-header Set-Cookie "sessionid=%[var(txn.t2)]; path=%[var(txn.path)]" if !{ var(txn.has_cookie) -m int gt 0 }
http-request track-sc1 var(txn.sessionid) table var(txn.index) if { var(txn.has_cookie) -m int gt 0 }
http-request track-sc1 var(txn.t2) table var(txn.index) if !{ var(txn.has_cookie) -m int gt 0 }
acl has_slot sc_get_gpc1(1) eq 1
acl free_slot sc_get_gpc0(0) gt 0
http-request sc-inc-gpc1(1) if free_slot !has_slot
use_backend %[var(txn.path_match),field(1,.)] if has_slot || free_slot
default_backend lineq

`

	cm.Data = map[string]string{
		"frontend-config-snippet": config,
	}

	r.clientset.CoreV1().ConfigMaps("haproxy-controller").Update(context.TODO(), cm, metav1.UpdateOptions{})
}

func (r *Runner) initAuxCfg() {
	auxCm, _ := r.clientset.CoreV1().ConfigMaps("haproxy-controller").Get(context.TODO(), "haproxy-auxiliary-configmap", metav1.GetOptions{})

	if _, exists := auxCm.Annotations["initialized"]; exists {
		return
	}

	auxCm.Annotations = map[string]string{
		"initialized": "true",
	}

	config := `
peers lineq
  server local
  server lineq lineq-tcp.lineq.svc:11111
backend room
  stick-table type string size 2 expire 1d store gpc0 peers lineq
backend lineq
  mode http
  server lineq lineq-http.lineq.svc:8060

`

	auxCm.Data = map[string]string{
		"haproxy-auxiliary.cfg": config,
	}
	r.clientset.CoreV1().ConfigMaps("haproxy-controller").Update(context.TODO(), auxCm, metav1.UpdateOptions{})
}

func (r *Runner) runHA(ctx context.Context) {
	if r.config.HA == (config.HA{}) || !r.config.HA.Enabled {
		r.logger.Fatal("HA config not set or not enabled")
	}

	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      r.config.HA.LeaseLockName,
			Namespace: r.config.Namespace,
		},
		Client: r.clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: r.config.HA.NodeId,
		},
	}
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   r.config.HA.LeaseDuration,
		RenewDeadline:   r.config.HA.RenewDeadline,
		RetryPeriod:     r.config.HA.RetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				r.logger.Info("start leading")
				r.initAuxCfg()
				r.initCfg()
				r.runSingleNode(ctx)
			},
			OnStoppedLeading: func() {
				r.logger.Info("stopped leading")
			},
			OnNewLeader: func(identity string) {
				if identity == r.config.HA.NodeId {
					r.logger.Info("obtained leadership")
					return
				}
				r.logger.Infof("leader elected: '%s'", identity)
			},
		},
	})
}

func NewRunner(
	ctrl *controller.Controller,
	clientset *kubernetes.Clientset,
	config config.Config,
	logger log.Logger,
) *Runner {
	return &Runner{
		ctrl:      ctrl,
		clientset: clientset,
		config:    config,
		logger:    logger,
	}
}
