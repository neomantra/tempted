package app

import (
	"sync"

	temporalClient "go.temporal.io/sdk/client"
)

var (
	updateID    int
	updateIDMtx sync.Mutex
)

func nextUpdateID() int {
	updateIDMtx.Lock()
	defer updateIDMtx.Unlock()
	updateID++
	return updateID
}

func (c Config) client() (*temporalClient.Client, error) {
	// opts :=
	client, err := temporalClient.NewLazyClient(temporalClient.Options{
		HostPort:  c.HostPort,
		Namespace: c.Namespace,
	})
	if err != nil {
		return nil, err
	}
	return &client, nil
}
