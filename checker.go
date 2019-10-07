package main

type HealthChecker interface {
	Healthy() bool
}

type ClusterHealthChecker struct {
}

func (c *ClusterHealthChecker) Healthy() bool {
	return true
}
