//
// Last.Backend LLC CONFIDENTIAL
// __________________
//
// [2014] - [2018] Last.Backend LLC
// All Rights Reserved.
//
// NOTICE:  All information contained herein is, and remains
// the property of Last.Backend LLC and its suppliers,
// if any.  The intellectual and technical concepts contained
// herein are proprietary to Last.Backend LLC
// and its suppliers and may be covered by Russian Federation and Foreign Patents,
// patents in process, and are protected by trade secret or copyright law.
// Dissemination of this information or reproduction of this material
// is strictly forbidden unless prior written permission is obtained
// from Last.Backend LLC.
//

package cache

import (
	"context"
		"sync"

	"github.com/lastbackend/lastbackend/pkg/distribution/types"
	"github.com/lastbackend/lastbackend/pkg/log"
)

const logCacheNode = "api:cache:node"

type CacheNodeManifest struct {
	lock      sync.RWMutex
	manifests map[string]*types.NodeManifest
}

type NetworkManifestWatcher func(ctx context.Context, event chan *types.Event) error

type PodManifestWatcher func(ctx context.Context, event chan *types.Event) error

type VolumeManifestWatcher func(ctx context.Context, event chan *types.Event) error

type EndpointManifestWatcher func(ctx context.Context, event chan *types.Event) error

func (c *CacheNodeManifest) checkNode(node string) {
	if _, ok := c.manifests[node]; !ok {
		c.manifests[node] = new(types.NodeManifest)
	}
}

func (c *CacheNodeManifest) SetPodManifest(node, pod string, s *types.PodManifest) {
	log.Infof("%s:PodManifestSet:> %s, %s, %#v", logCacheNode, node, pod, s)
	c.lock.Lock()
	defer c.lock.Unlock()

	c.checkNode(node)

	if c.manifests[node].Pods == nil {
		sp := c.manifests[node]
		sp.Pods = make(map[string]*types.PodManifest, 0)
	}

	c.manifests[node].Pods[pod] = s
}

func (c *CacheNodeManifest) DelPodManifest(node, pod string) {
	log.Infof("%s:PodManifestDel:> %s, %s", node, pod)
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.manifests[node]; !ok {
		return
	}

	delete(c.manifests[node].Pods, pod)
}

func (c *CacheNodeManifest) SetVolumeManifest(node, volume string, s *types.VolumeManifest) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.checkNode(node)

	if c.manifests[node].Volumes == nil {
		sp := c.manifests[node]
		sp.Volumes = make(map[string]*types.VolumeManifest, 0)
	}

	c.manifests[node].Volumes[volume] = s
}

func (c *CacheNodeManifest) DelVolumeManifest(node, volume string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.manifests[node]; !ok {
		return
	}

	delete(c.manifests[node].Volumes, volume)
}

func (c *CacheNodeManifest) SetSubnetManifest(cidr string, s *types.SubnetManifest) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for n := range c.manifests {

		if _, ok := c.manifests[n].Network[cidr]; !ok {
			c.manifests[n].Network = make(map[string]*types.SubnetManifest)
		}

		c.manifests[n].Network[cidr] = s
	}
}

func (c *CacheNodeManifest) SetEndpointManifest(addr string, s *types.EndpointManifest) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, n := range c.manifests {
		if n.Endpoints == nil {
			n.Endpoints = make(map[string]*types.EndpointManifest, 0)
		}
		n.Endpoints[addr] = s
	}
}

func (c *CacheNodeManifest) Get(node string) *types.NodeManifest {
	c.lock.Lock()
	defer c.lock.Unlock()
	if s, ok := c.manifests[node]; !ok {
		return nil
	} else {
		return s
	}
}

func (c *CacheNodeManifest) Flush(node string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.manifests[node] = new(types.NodeManifest)
}

func (c *CacheNodeManifest) Clear(node string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.manifests, node)
}

func NewCacheNodeManifest() *CacheNodeManifest {
	c := new(CacheNodeManifest)
	c.manifests = make(map[string]*types.NodeManifest, 0)
	return c
}