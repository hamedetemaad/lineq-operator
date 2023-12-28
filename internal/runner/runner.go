package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gotway/gotway/pkg/log"
	"github.com/hamedetemaad/lineq-operator/internal/config"
	"github.com/hamedetemaad/lineq-operator/pkg/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

type ResponseBody struct {
	Status               string `json:"status"`
	Message              string `json:"message"`
	RoomTableName        string `json:"lineq_room_table"`
	UserTableName        string `json:"lineq_user_table"`
	LineqSessionDuration int    `json:"lineq_session_duration"`
}

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
	if err := r.ctrl.Run(ctx, r.config.NumWorkers, r.config); err != nil {
		r.logger.Fatal("error running controller ", err)
	}
}

func (r *Runner) getCfg() {
	url := fmt.Sprintf("http://%s:%d/getConfig", r.config.LineqHttpAddr, r.config.LineqHttpPort)

	response, err := http.Get(url)
	if err != nil {
		r.logger.Infof("Error sending request: %v", err)
		return
	}
	defer response.Body.Close()

	var responseBody ResponseBody

	err = json.NewDecoder(response.Body).Decode(&responseBody)
	if err != nil {
		r.logger.Infof("Error decoding JSON response: '%v'", err)
		return
	}

	r.config.LineqSessionDuration = responseBody.LineqSessionDuration
	r.config.RoomTableName = responseBody.RoomTableName
	r.config.UserTableName = responseBody.UserTableName

	r.logger.Infof("Response Status: '%s'", responseBody.Status)
	r.logger.Infof("Response Message: '%s'", responseBody.Message)
}

func (r *Runner) initCfg() {
	cm, _ := r.clientset.CoreV1().ConfigMaps("haproxy-controller").Get(context.TODO(), "haproxy-kubernetes-ingress", metav1.GetOptions{})

	if _, exists := cm.Annotations["initialized"]; exists {
		return
	}

	cm.Annotations["initialized"] = "true"

	config := `

	http-request set-var(txn.vwr_path) var(txn.host),concat('.vwr',txn.path),map(/etc/haproxy/maps/path-exact.map)
	http-request set-var(txn.has_cookie) req.cook_cnt(sessionid) if { var(txn.vwr_path) -m found }
	http-request set-var(txn.t2) uuid()  if { var(txn.vwr_path) -m found } !{ var(txn.has_cookie) -m int gt 0 }
	http-request set-var(txn.sessionid) req.cook(sessionid) if { var(txn.vwr_path) -m found }
	http-request set-var(txn.index) var(txn.host),regsub(\.,_,g),concat(,txn.path,),regsub(\/,_,g) if { var(txn.vwr_path) -m found }
	http-request track-sc0 var(txn.index) table %s if { var(txn.vwr_path) -m found }
	http-response add-header Set-Cookie "sessionid=%%[var(txn.t2)]; path=%%[var(txn.path)]" if { var(txn.vwr_path) -m found } !{ var(txn.has_cookie) -m int gt 0 }
	http-request track-sc1 var(txn.sessionid),concat('@',txn.index) table %s if { var(txn.vwr_path) -m found } { var(txn.has_cookie) -m int gt 0 }
	http-request track-sc1 var(txn.t2),concat('@',txn.index) table %s if { var(txn.vwr_path) -m found } !{ var(txn.has_cookie) -m int gt 0 }
	http-request sc-inc-gpc1(1) if { var(txn.vwr_path) -m found } { sc_get_gpc0(0) gt 0 } !{ sc_get_gpc1(1) eq 1 }
	use_backend %%[var(txn.path_match),field(1,.)] if !{ var(txn.vwr_path) -m found } !{ path_sub /lineq }
	use_backend %%[var(txn.vwr_path),field(1,.)] if { sc_get_gpc1(1) eq 1 } || { sc_get_gpc0(0) gt 0 }
	use_backend lineq

`

	config = fmt.Sprintf(config, r.config.RoomTableName, r.config.UserTableName, r.config.UserTableName)

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
  server lineq %s:%d
backend %s
  stick-table type string size 10 expire 1d store gpc0 peers lineq
backend %s
  stick-table type string len 72 size 100k expire %dm store gpc1 peers lineq
backend lineq
  mode http
  server lineq %s:%d

`

	config = fmt.Sprintf(config, r.config.LineqTcpAddr, r.config.LineqTcpPort, r.config.RoomTableName, r.config.UserTableName, r.config.LineqSessionDuration, r.config.LineqHttpAddr, r.config.LineqHttpPort)

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
				r.getCfg()
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
