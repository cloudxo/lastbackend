//
// Last.Backend LLC CONFIDENTIAL
// __________________
//
// [2014] - [2017] Last.Backend LLC
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

package types

import (
	"sync"
	"time"
)


type Pod struct {
	lock sync.RWMutex

	// Pod Meta
	Meta PodMeta `json:"meta"`
	// Container spec
	Spec PodSpec `json:"spec"`
	// Containers status info
	Containers map[string]*Container `json:"containers"`
	// Secrets
	Secrets map[string]*PodSecret `json:"secrets"`
}

type PodNodeSpec struct {
	// Pod Meta
	Meta PodMeta `json:"meta"`
	// Pod spec
	Spec PodSpec `json:"spec"`
}

type PodNodeState struct {
	// Pod Meta
	Meta PodMeta `json:"meta"`
	// Containers status info
	Containers map[string]*Container `json:"containers"`
}


type PodMeta struct {
	Meta
	// Pod state
	State PodState `json:"state"`
	// Pod hostname
	Hostname string `json:"hostname"`
}

type PodSpec struct {
	// Provision ID
	ID string `json:"id"`
	// Provision state
	State string `json:"state"`
	// Provision status
	Status string `json:"status"`

	// Containers spec for pod
	Containers []*ContainerSpec `json:"containers"`

	// Provision create time
	Created time.Time `json:"created"`
	// Provision update time
	Updated time.Time `json:"updated"`
}

type PodState struct {
	// Pod current state
	State string `json:"state"`
	// Pod current status
	Status string `json:"status"`
	// Container total
	Containers PodContainersState `json:"containers"`
}

type PodContainersState struct {
	// Total containers
	Total int `json:"total"`
	// Total running containers
	Running int `json:"running"`
	// Total created containers
	Created int `json:"created"`
	// Total stopped containers
	Stopped int `json:"stopped"`
	// Total errored containers
	Errored int `json:"errored"`
}

type PodSecret struct {}

type PodCRIMeta struct {
	Type    string `json:"type"`
	Version string `json:"version"`
}

type PodNetwork struct {
	Interface string   `json:"interface,omitempty"`
	IP        []string `json:"ip,omitempty"`
}


func (p *Pod) AddContainer(c *Container) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.Containers[c.ID] = c
}

func (p *Pod) SetContainer(c *Container) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.Containers[c.ID] = c
}

func (p *Pod) DelContainer(ID string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	delete(p.Containers, ID)
}

func (p *Pod) UpdateState() {

	p.Meta.State.State = ""
	p.Meta.State.Status = ""
	p.Meta.State.Containers = PodContainersState{}

	for _, c := range p.Containers {
		p.Meta.State.Containers.Total++

		if c.State == "created" {
			p.Meta.State.Containers.Created++
		}

		if c.State == "running" {
			p.Meta.State.Containers.Running++
		}

		if c.State == "stopped" {
			p.Meta.State.Containers.Stopped++
		}

		if c.State == "exited" {
			p.Meta.State.Containers.Stopped++
		}

		if c.State == "error" {
			p.Meta.State.Containers.Errored++
		}

		if c.State == p.Meta.State.State {
			continue
		}

		if c.State == "exited" && p.Meta.State.Status == "" {
			p.Meta.State.State = "stopped"
			continue
		}

		if p.Meta.State.State == "" {
			p.Meta.State.State = c.State
			continue
		}

		if c.State == "exited" && p.Meta.State.Status != "stopped" {
			p.Meta.State.State = "warning"
			continue
		}

		if c.State == "running" && p.Meta.State.Status != "running" {
			p.Meta.State.State = "warning"
			continue
		}

		if p.Meta.State.State == "error" {
			continue
		}


		if c.State == "error" {
			p.Meta.State.Containers.Errored++
			p.Meta.State.State = c.State
			p.Meta.State.Status = c.Status
			continue
		}
	}

}

const PodStateRunning = "running"
const PodStateStarted = "started"
const PodStateRestarted = "restarted"
const PodStateStopped = "stopped"

func NewPod() *Pod {
	return &Pod{
		Containers: make(map[string]*Container),
		Secrets:    make(map[string]*PodSecret),
	}
}
