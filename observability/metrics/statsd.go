package metrics

import (
	"fmt"

	"github.com/DataDog/datadog-go/v5/statsd"
)

const (
	dogstatsdAddr = "unix:///var/run/datadog/dsd.socket"
)

type StatsDClient struct {
	*statsd.Client
}

func NewMetricAgent(service, version, env, tenant, team string) (*StatsDClient, error) {
	tags := []string{
		fmt.Sprintf("env:%s", env),
		fmt.Sprintf("service:%s", service),
		fmt.Sprintf("version:%s", version),
		fmt.Sprintf("flyr_tenant:%s", tenant),
		fmt.Sprintf("flyr_team:%s", team),
	}

	client, err := statsd.New(dogstatsdAddr, statsd.WithTags(tags))
	if err != nil {
		return nil, fmt.Errorf("statsd.New: %v", err)
	}
	return &StatsDClient{Client: client}, nil
}
