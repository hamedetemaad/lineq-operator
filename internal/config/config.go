package config

import (
	"fmt"
	"os"
	"time"

	"github.com/gotway/gotway/pkg/env"
)

type HA struct {
	Enabled       bool
	NodeId        string
	LeaseLockName string
	LeaseDuration time.Duration
	RenewDeadline time.Duration
	RetryPeriod   time.Duration
}

func (c HA) String() string {
	return fmt.Sprintf(
		"HA{Enabled='%v'NodeId='%s'LeaseLockName='%s'LeaseDuration='%v'RenewDeadline='%v'RetryPeriod='%v'}",
		c.Enabled,
		c.NodeId,
		c.LeaseLockName,
		c.LeaseDuration,
		c.RenewDeadline,
		c.RetryPeriod,
	)
}

type Metrics struct {
	Enabled bool
	Path    string
	Port    string
}

func (m Metrics) String() string {
	return fmt.Sprintf(
		"Metrics{Enabled='%v'Path='%s'Port='%s'}",
		m.Enabled,
		m.Path,
		m.Port,
	)
}

type Config struct {
	KubeConfig           string
	Namespace            string
	NumWorkers           int
	HA                   HA
	Metrics              Metrics
	Env                  string
	LogLevel             string
	LineqTcpAddr         string
	LineqHttpAddr        string
	LineqTcpPort         int
	LineqHttpPort        int
	RoomTableName        string
	UserTableName        string
	LineqSessionDuration int
}

func (c Config) String() string {
	return fmt.Sprintf(
		"Config{KubeConfig='%s'Namespace='%s'NumWorkers='%d'HA='%v'Metrics='%v'Env='%s'LogLevel='%s'LineqTcpAddr='%s'LineqHttpAddr='%s'LineqTcpPort='%d'LineqHttpPort='%d'RoomTableName='%s'UserTableName='%s'LineqSessionDuration='%d'}",
		c.KubeConfig,
		c.Namespace,
		c.NumWorkers,
		c.HA,
		c.Metrics,
		c.Env,
		c.LogLevel,
		c.LineqTcpAddr,
		c.LineqHttpAddr,
		c.LineqTcpPort,
		c.LineqHttpPort,
		c.RoomTableName,
		c.UserTableName,
		c.LineqSessionDuration,
	)
}

func GetConfig() (Config, error) {
	ha := env.GetBool("HA_ENABLED", false)

	var nodeId string
	if ha {
		nodeId = env.Get("HA_NODE_ID", "")
		if nodeId == "" {
			hostname, err := os.Hostname()
			if err != nil {
				return Config{}, fmt.Errorf("error getting node id %v", err)
			}
			nodeId = hostname
		}
	}

	return Config{
		KubeConfig: env.Get("KUBECONFIG", ""),
		Namespace:  env.Get("NAMESPACE", "default"),
		NumWorkers: env.GetInt("NUM_WORKERS", 4),
		HA: HA{
			Enabled:       ha,
			NodeId:        nodeId,
			LeaseLockName: env.Get("HA_LEASE_LOCK_NAME", "waitingroomoperator"),
			LeaseDuration: env.GetDuration("HA_LEASE_DURATION_SECONDS", 15) * time.Second,
			RenewDeadline: env.GetDuration("HA_RENEW_DEADLINE_SECONDS", 10) * time.Second,
			RetryPeriod:   env.GetDuration("HA_RETRY_PERIOD_SECONDS", 2) * time.Second,
		},
		Metrics: Metrics{
			Enabled: env.GetBool("METRICS_ENABLED", true),
			Path:    env.Get("METRICS_PATH", "/metrics"),
			Port:    env.Get("METRICS_PORT", "2112"),
		},
		Env:           env.Get("ENV", "local"),
		LogLevel:      env.Get("LOG_LEVEL", "debug"),
		LineqTcpAddr:  env.Get("LINEQ_TCP_ADDR", "lineq-tcp.lineq.svc"),
		LineqHttpAddr: env.Get("LINEQ_HTTP_ADDR", "lineq-http.lineq.svc"),
		LineqTcpPort:  env.GetInt("LINEQ_TCP_PORT", 11111),
		LineqHttpPort: env.GetInt("LINEQ_HTTP_PORT", 8060),
	}, nil
}
