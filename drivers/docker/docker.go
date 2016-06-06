package docker

import (
	"github.com/deployithq/deployit/drivers/interfaces"
	"github.com/fsouza/go-dockerclient"
	"net"
	"strings"
)

type Containers struct {
}

var (
	DOCKER_URI  = ""
	DOCKER_CERT = ""
	DOCKER_CA   = ""
	DOCKER_KEY  = ""
)

func (d *Containers) client() (*docker.Client, error) {

	var err error
	var client *docker.Client
	switch {
	case DOCKER_URI != "" && DOCKER_CERT == "" && DOCKER_CA == "" && DOCKER_KEY == "":
		client, err = docker.NewClient(DOCKER_URI)
		break

	case DOCKER_URI != "" && DOCKER_CERT != "" && DOCKER_CA != "" && DOCKER_KEY != "":
		client, err = docker.NewTLSClient(DOCKER_URI, DOCKER_CERT, DOCKER_KEY, DOCKER_CA)
		break
	default:
		client, err = docker.NewClientFromEnv()
	}

	if err != nil {
		return client, err
	}

	return client, nil
}

func (d *Containers) System() (*interfaces.Node, error) {
	system := &interfaces.Node{}
	client, err := d.client()
	if err != nil {
		return system, err
	}

	info, err := client.Info()
	if err != nil {
		return system, err
	}

	ver, err := client.Version()
	if err != nil {
		return system, err
	}

	system.Driver.Name = "docker"
	system.Driver.Version = ver.Get("Version")
	system.Hostname = info.Name
	system.Architecture = info.Architecture

	system.OS.Type = info.OSType
	system.OS.Name = info.OperatingSystem
	system.Memory.Total = info.MemTotal / 1024 / 1024

	for dr := range info.DriverStatus {
		switch info.DriverStatus[dr][0] {
		case "Data Space Total":
			system.Storage.Total = info.DriverStatus[dr][1]
		case "Data Space Used":
			system.Storage.Used = info.DriverStatus[dr][1]
		case "Data Space Available":
			system.Storage.Available = info.DriverStatus[dr][1]
		}
	}

	ifaces, err := net.Interfaces()
	// handle err
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return system, err
		}
		// handle err
		for _, addr := range addrs {

			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					system.IPs = append(system.IPs, struct {
						Interface string `json:"interfaces,omitempty"`
						IP        string `json:"ip,omitempty"`
						Main      bool   `json:"main,omitempty"`
					}{
						Interface: i.Name,
						IP:        ipnet.IP.To4().String(),
						Main:      i.Name == `eth0`,
					})
				}
			}
		}
	}

	return system, nil
}

func (d *Containers) PullImage(i interfaces.Image) error {

	registry := "index.docker.io"
	repo := i.Name
	tag := "latest"

	client, err := d.client()
	if err != nil {
		return err
	}

	s := strings.Split(i.Name, "/")
	if len(s) == 2 {
		registry = s[0]
		repo = s[1]
	}

	if len(s) == 3 {
		registry = s[0]
		repo = s[2]
	}

	t := strings.Split(repo, ":")
	if len(t) == 2 {
		tag = t[1]
	}

	return client.PullImage(docker.PullImageOptions{
		Repository: i.Name,
		Registry:   registry,
		Tag:        tag,
	}, docker.AuthConfiguration{
		Username:      i.Auth.Username,
		Password:      i.Auth.Password,
		Email:         i.Auth.Email,
		ServerAddress: i.Auth.Host,
	})
}

func (d *Containers) BuildImage(opts interfaces.BuildImageOptions) error {

	client, err := d.client()
	if err != nil {
		return err
	}

	o := docker.BuildImageOptions{
		Name:           opts.Name,
		RmTmpContainer: opts.RmTmpContainer,
		InputStream:    opts.InputStream,
		OutputStream:   opts.OutputStream,
		ContextDir:     opts.ContextDir,
		RawJSONStream:  opts.RawJSONStream,
	}

	if err := client.BuildImage(o); err != nil {
		return err
	}

	return nil
}

func (d *Containers) StartContainer(c *interfaces.Container) error {

	client, err := d.client()
	if err != nil {
		return err
	}

	config := CreateConfig(c.Config)
	hostconf := CreateHostconfig(c.HostConfig)

	if c.CID == "" {
		options := docker.CreateContainerOptions{
			Config:     &config,
			HostConfig: &hostconf,
		}

		container, err := client.CreateContainer(options)
		if err != nil {
			return err
		}

		c.CID = container.ID
	}

	if err := client.StartContainer(c.CID, &hostconf); err != nil {
		return err
	}

	return nil
}

func (d *Containers) StopContainer(c *interfaces.Container) error {

	client, err := d.client()
	if err != nil {
		return err
	}

	return client.StopContainer(c.CID, 10)
}

func (d *Containers) RestartContainer(c *interfaces.Container) error {
	client, err := d.client()
	if err != nil {
		return err
	}

	return client.RestartContainer(c.CID, 10)
}

func (d *Containers) RemoveContainer(c *interfaces.Container) error {
	client, err := d.client()
	if err != nil {
		return err
	}

	return client.RemoveContainer(docker.RemoveContainerOptions{
		ID:            c.CID,
		RemoveVolumes: true,
		Force:         true,
	})
}

func (d *Containers) ListImages() (map[string]interfaces.Image, error) {

	var (
		images = make(map[string]interfaces.Image)
		err    error
	)

	if err != nil {
		return images, err
	}

	client, err := d.client()
	if err != nil {
		return images, err
	}

	ims, err := client.ListImages(docker.ListImagesOptions{All: true})
	if err != nil {
		return images, err
	}

	for index := range ims {
		i := ims[index]

		for index := range i.RepoTags {
			name := i.RepoTags[index]

			if name == "<none>:<none>" {
				continue
			}

			if _, ok := images[name]; !ok {
				images[name] = interfaces.Image{}
				//images[name].Name = strings.Split(name, ":")[0]
			}

			im, err := client.InspectImage(name)
			if err != nil {
				continue
			}

			if images[name], err = convertImage(im); err != nil {
				return images, err
			}
		}
	}

	return images, nil
}

func (d *Containers) ListContainers() (map[string]interfaces.Container, error) {

	var (
		containers = make(map[string]interfaces.Container)
		err        error
	)

	client, err := d.client()
	if err != nil {
		return containers, err
	}

	cs, err := client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return containers, err
	}

	for i := range cs {
		c := cs[i]

		if _, ok := containers[c.ID]; !ok {
			containers[c.ID] = interfaces.Container{}
		}

		info, err := client.InspectContainer(c.ID)
		if err != nil {
			return containers, err
		}

		if containers[c.ID], err = ConvertContainer(info); err != nil {
			return containers, err
		}
	}

	return containers, err
}