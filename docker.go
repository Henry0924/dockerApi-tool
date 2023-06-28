package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var DockerApi = dockerApi{}

type dockerApi struct{}

func (a *dockerApi) newClient(host string, timeout client.Opt) (cli *client.Client, err error) {
	host = fmt.Sprintf("http://%s:2375", host)
	if timeout == nil {
		cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation(), client.WithHost(host))
	} else {
		cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation(), client.WithHost(host), timeout)
	}
	if err != nil {
		return
	}
	return
}

func (a *dockerApi) CreateContainer(index int, host string, name string, bridged bool, params map[string]interface{}) (containerId string) {
	cli, err := a.newClient(host, nil)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	image := params["image"].(string)
	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	imageExists := false
	for _, ig := range images {
		if len(ig.RepoTags) == 0 {
			continue
		}
		if image == ig.RepoTags[0] {
			imageExists = true
			break
		}
	}

	if !imageExists {
		out, err := cli.ImagePull(context.Background(), image, types.ImagePullOptions{})
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()
		io.Copy(os.Stdout, out)
	}

	hostConfig := genContainerHostConfig(index, bridged)

	cmd := []string{
		"androidboot.hardware=rk30board",
		"androidboot.redroid_fps=24",
		"androidboot.selinux=permissive",
		"qemu=1",
	}

	exposedPorts := map[nat.Port]struct{}{"9082/tcp": {}, "10000/tcp": {}, "10001/udp": {}, "5555/tcp": {}}

	netConfig := new(network.NetworkingConfig)
	if bridged {
		netId := a.CreateMacvlan(host, params["gateway"].(string), params["subnet"].(string))
		netConfig.EndpointsConfig = map[string]*network.EndpointSettings{
			"myt": &network.EndpointSettings{IPAMConfig: &network.EndpointIPAMConfig{IPv4Address: params["androidHost"].(string)}, NetworkID: netId},
		}

		cmd = append(cmd, []string{fmt.Sprintf("net.dns1=%s", params["dns1"]), fmt.Sprintf("net.dns2=%s", params["dns2"])}...)

		exposedPorts = map[nat.Port]struct{}{}
	}

	resp, err := cli.ContainerCreate(context.Background(), &container.Config{
		Image:        image,
		Tty:          true,
		OpenStdin:    true,
		Cmd:          cmd,
		ExposedPorts: exposedPorts,
	}, hostConfig, netConfig, nil, name)
	if err != nil {
		log.Fatal(err)
		return
	}

	containerId = resp.ID
	return
}

func genContainerHostConfig(index int, bridged bool) (hostConfig *container.HostConfig) {
	now := time.Now().Unix()
	hostConfig = new(container.HostConfig)

	binderIndex := (index-1)*3 + 1
	resources := container.Resources{
		Devices: []container.DeviceMapping{
			{fmt.Sprintf("/dev/binder%d", binderIndex), "/dev/binder", "rwm"},
			{fmt.Sprintf("/dev/binder%d", binderIndex+1), "/dev/hwbinder", "rwm"},
			{fmt.Sprintf("/dev/binder%d", binderIndex+2), "/dev/vndbinder", "rwm"},
			{"/dev/tee0", "/dev/tee0", "rwm"},
			{"/dev/teepriv0", "/dev/teepriv0", "rwm"},
			{"/dev/crypto", "/dev/crypto", "rwm"},
			{"/dev/mali0", "/dev/mali0", "rwm"},
			{"/dev/rga", "/dev/rga", "rwm"},
			{"/dev/dri", "/dev/dri", "rwm"},
			{"/dev/mpp_service", "/dev/mpp_service", "rwm"},
			{"/dev/fuse", "/dev/fuse", "rwm"},
			{"/dev/input/event0", "/dev/input/event0", "rwm"},
			{"/dev/dma_heap/cma", "/dev/dma_heap/cma", "rwm"},
			{"/dev/dma_heap/cma-uncached", "/dev/dma_heap/cma-uncached", "rwm"},
			{"/dev/dma_heap/system", "/dev/dma_heap/system", "rwm"},
			{"/dev/dma_heap/system-dma32", "/dev/dma_heap/system-dma32", "rwm"},
			{"/dev/dma_heap/system-uncached", "/dev/dma_heap/system-uncached", "rwm"},
			{"/dev/dma_heap/system-uncached-dma32", "/dev/dma_heap/system-uncached-dma32", "rwm"},
			{"/dev/ashmem", "/dev/ashmem", "rwm"},
		},
	}
	hostConfig.Resources = resources

	hostConfig.Binds = []string{
		fmt.Sprintf("/mmc/custom/data%d_%d/data:/data", index, now),
		"/dev/net/tun:/dev/tun",
		"/dev/mali0:/dev/mali0",
	}
	hostConfig.RestartPolicy = container.RestartPolicy{Name: "unless-stopped"}

	if !bridged {
		var (
			tcpPort = 10000 + index*3
			udpPort = tcpPort + 1
			webPort = tcpPort + 2
			adbPort = 5000 + index
		)
		hostConfig.PortBindings = map[nat.Port][]nat.PortBinding{
			"9082/tcp":  {nat.PortBinding{HostIP: "", HostPort: strconv.Itoa(webPort)}},
			"10000/tcp": {nat.PortBinding{HostIP: "", HostPort: strconv.Itoa(tcpPort)}},
			"10001/udp": {nat.PortBinding{HostIP: "", HostPort: strconv.Itoa(udpPort)}},
			"5555/tcp":  {nat.PortBinding{HostIP: "", HostPort: strconv.Itoa(adbPort)}},
		}
		hostConfig.NetworkMode = "default"
	} else {
		hostConfig.NetworkMode = "myt"
	}

	hostConfig.CapAdd = []string{
		"SYSLOG",
		"AUDIT_CONTROL",
		"SETGID",
		"DAC_READ_SEARCH",
		"SYS_ADMIN",
		"NET_ADMIN",
		"SYS_MODULE",
		"SYS_NICE",
		"SYS_TIME",
		"SYS_TTY_CONFIG",
		"NET_BROADCAST",
		"IPC_LOCK",
		"SYS_RESOURCE",
		"SYS_PTRACE",
		"WAKE_ALARM",
		"BLOCK_SUSPEND",
	}

	hostConfig.SecurityOpt = []string{"seccomp=unconfined"}
	hostConfig.Sysctls = map[string]string{"net.ipv4.conf.eth0.rp_filter": "2"}
	return
}

func (a *dockerApi) getNetwork(cli *client.Client) []types.NetworkResource {
	list, err := cli.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	return list
}

func (a *dockerApi) CreateMacvlan(host string, gateway string, subnet string) string {
	cli, err := a.newClient(host, nil)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	list := a.getNetwork(cli)
	for _, resource := range list {
		if resource.Name == "myt" {
			return resource.ID
		}
	}
	config := types.NetworkCreate{
		Driver:     "macvlan",
		EnableIPv6: false,
		IPAM: &network.IPAM{
			Driver: "default",
			Config: []network.IPAMConfig{{Subnet: subnet, Gateway: gateway}},
		},
		Options: map[string]string{"parent": "eth0"},
	}

	resp, err := cli.NetworkCreate(ctx, "myt", config)
	if err != nil {
		log.Fatal(err)
	}
	return resp.ID
}

func (a *dockerApi) List(host string) (containers []types.Container) {
	cli, err := a.newClient(host, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	ctx := context.Background()

	containers, err = cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		log.Fatal(err)
		return
	}
	return
}

func (a *dockerApi) containerIdByName(cli *client.Client, name string) (id string) {
	ctx := context.Background()

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, t := range containers {
		if len(t.Names) > 0 {
			if strings.ReplaceAll(t.Names[0], "/", "") == name {
				id = t.ID
				break
			}
		}
	}

	return
}

func (a *dockerApi) Start(host string, name string) (id string) {
	cli, err := a.newClient(host, nil)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	id = a.containerIdByName(cli, name)
	if id == "" {
		log.Fatal("容器别名不存在")
	}

	if err = cli.ContainerStart(ctx, id, types.ContainerStartOptions{}); err != nil {
		log.Fatal(err)
	}

	return
}

func (a *dockerApi) Stop(host string, name string) (id string) {
	cli, err := a.newClient(host, nil)
	if err != nil {
		return
	}
	ctx := context.Background()

	id = a.containerIdByName(cli, name)
	if id == "" {
		log.Fatal("容器别名不存在")
	}

	if err = cli.ContainerStop(ctx, id, container.StopOptions{}); err != nil {
		log.Fatal(err)
	}

	return
}

func (a *dockerApi) Pause(host string, name string) (id string) {
	cli, err := a.newClient(host, nil)
	if err != nil {
		return
	}
	ctx := context.Background()

	id = a.containerIdByName(cli, name)
	if id == "" {
		log.Fatal("容器别名不存在")
	}

	if err = cli.ContainerPause(ctx, id); err != nil {
		log.Fatal(err)
	}

	return
}

func (a *dockerApi) Unpause(host string, name string) (id string) {
	cli, err := a.newClient(host, nil)
	if err != nil {
		return
	}
	ctx := context.Background()

	id = a.containerIdByName(cli, name)
	if id == "" {
		log.Fatal("容器别名不存在")
	}

	if err = cli.ContainerUnpause(ctx, id); err != nil {
		log.Fatal(err)
	}
	return
}

func (a *dockerApi) Remove(host string, name string) (id string) {
	cli, err := a.newClient(host, nil)
	if err != nil {
		return
	}
	ctx := context.Background()

	id = a.containerIdByName(cli, name)
	if id == "" {
		log.Fatal("容器别名不存在")
	}

	if err = cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: true}); err != nil {
		log.Fatal(err)
	}
	return
}
