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

package runtime

import (
	"github.com/lastbackend/lastbackend/pkg/distribution"
	"context"
	"github.com/lastbackend/lastbackend/pkg/api/envs"
	"github.com/lastbackend/lastbackend/pkg/distribution/types"
)

type Runtime struct {
}

func New() *Runtime {
	return new(Runtime)
}

func (r *Runtime) Run() {

	var ctx = context.Background()

	go r.podManifestWatch(ctx, nil)
	go r.volumeManifestWatch(ctx, nil)
	go r.endpointManifestWatch(ctx, nil)
	go r.subnetManifestWatch(ctx, nil)

	//go c.Node().CachePods(stg.Node().EventPodSpec)
	//go c.Node().CacheVolumes(stg.Node().EventVolumeSpec)
	//go c.Node().CacheEndpoints(stg.Endpoint().EventSpec)
	//go c.Node().Del(stg.Node().EventStatus)
	//
	//go c.Ingress().CacheRoutes(stg.Route().WatchSpecEvents)
	//go c.Ingress().Status(stg.Ingress().WatchStatus)
}

func (r *Runtime) podManifestWatch(ctx context.Context, rev *int64) {

	// Watch pods change
	var (
		p = make(chan types.PodManifestEvent)
		c = envs.Get().GetCache()
	)

	mm := distribution.NewPodModel(ctx, envs.Get().GetStorage())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case w := <-p:

				if w.Data == nil {
					continue
				}

				if w.IsActionRemove() {
					c.Node().DelPodManifest(w.Node, w.SelfLink)
					continue
				}

				c.Node().SetPodManifest(w.Node, w.SelfLink, w.Data)
			}
		}
	}()

	mm.ManifestWatch(types.EmptyString, p, rev)
}

func (r *Runtime) volumeManifestWatch(ctx context.Context, rev *int64) {

	// Watch volumes change
	var (
		v = make(chan types.VolumeManifestEvent)
		c = envs.Get().GetCache()
	)

	mm := distribution.NewVolumeModel(ctx, envs.Get().GetStorage())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case w := <-v:

				if w.Data == nil {
					continue
				}

				if w.IsActionRemove() {
					c.Node().DelVolumeManifest(w.Node, w.SelfLink)
					continue
				}

				c.Node().SetVolumeManifest(w.Node, w.SelfLink, w.Data)
			}
		}
	}()

	mm.ManifestWatch(types.EmptyString, v, rev)
}

func (r *Runtime) endpointManifestWatch(ctx context.Context, rev *int64) {

	// Watch volumes change
	var (
		v = make(chan types.EndpointManifestEvent)
		c = envs.Get().GetCache()
	)

	mm := distribution.NewEndpointModel(ctx, envs.Get().GetStorage())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case w := <-v:

				if w.Data == nil {
					continue
				}

				if w.IsActionRemove() {
					w.Data.State = types.StateDestroy
				}

				c.Node().SetEndpointManifest(w.Name, w.Data)
			}
		}
	}()

	mm.ManifestWatch(v, rev)
}

func (r *Runtime) subnetManifestWatch(ctx context.Context, rev *int64) {

	// Watch volumes change
	var (
		v = make(chan types.SubnetManifestEvent)
		c = envs.Get().GetCache()
	)

	mm := distribution.NewNetworkModel(ctx, envs.Get().GetStorage())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case w := <-v:

				if w.Data == nil {
					continue
				}

				if w.IsActionRemove() {
					w.Data.State = types.StateDestroy
				}

				c.Node().SetSubnetManifest(w.Name, w.Data)
			}
		}
	}()

	mm.SubnetManifestWatch(v, rev)
}

func (r *Runtime) nodeWatch(ctx context.Context, rev *int64) {

	// Watch node changes
	var (
		n = make(chan types.NodeEvent)
		c = envs.Get().GetCache()
	)

	mm := distribution.NewNodeModel(ctx, envs.Get().GetStorage())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case w := <-n:

				if w.Data == nil {
					continue
				}

				if w.IsActionRemove() {
					c.Node().Clear(w.Name)
					continue
				}

				if !w.Data.Status.Online {
					c.Node().Clear(w.Name)
				}

			}
		}
	}()

	mm.Watch(n, rev)
}
