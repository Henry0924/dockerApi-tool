package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	command = flag.String("command", "list", "操作指令, create--创建容器, start--启动容器, stop--停止容器, pause--暂停容器, unpause--取消暂停, remove--删除容器, list--容器列表")
	index   = flag.String("i", "1", "容器序号1-12, 创建容器时必传, 例1、2、3")
	host    = flag.String("host", "192.168.30.3", "主机host")
	name    = flag.String("name", "", "容器别名, 例001、001,002、001,002,003")
)

func main() {
	flag.Parse()

	if *host == "" {
		log.Fatal("host不能为空")
	}

	switch *command {
	case "create":
		if *name == "" {
			log.Fatal("容器别名不能为空")
		}

		var config map[string]interface{}
		content, err := os.ReadFile("config.json")
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(content, &config)
		if err != nil {
			log.Fatal(err)
		}

		atoi, _ := strconv.Atoi(*index)
		if atoi < 1 || atoi > 12 {
			log.Fatal("容器序号错误")
		}
		nameSplit := strings.Split(*name, ",")
		for _, s := range nameSplit {
			id := DockerApi.CreateContainer(atoi, *host, strings.TrimSpace(s), config["image"].(string))
			log.Printf("容器授权序号: %d, 容器别名: %s, 容器ID: %s", atoi, strings.TrimSpace(s), id)
		}

	case "start":
		if *name == "" {
			log.Fatal("容器别名不能为空")
		}
		nameSplit := strings.Split(*name, ",")
		for _, s := range nameSplit {
			id := DockerApi.Start(*host, strings.TrimSpace(s))
			log.Printf("start 容器别名: %s, 容器ID: %s", strings.TrimSpace(s), id)
		}

	case "stop":
		if *name == "" {
			log.Fatal("容器别名不能为空")
		}
		nameSplit := strings.Split(*name, ",")
		for _, s := range nameSplit {
			id := DockerApi.Stop(*host, strings.TrimSpace(s))
			log.Printf("stop 容器别名: %s, 容器ID: %s", strings.TrimSpace(s), id)
		}

	case "pause":
		if *name == "" {
			log.Fatal("容器别名不能为空")
		}
		nameSplit := strings.Split(*name, ",")
		for _, s := range nameSplit {
			id := DockerApi.Pause(*host, strings.TrimSpace(s))
			log.Printf("pause 容器别名: %s, 容器ID: %s", strings.TrimSpace(s), id)
		}

	case "unpause":
		if *name == "" {
			log.Fatal("容器别名不能为空")
		}
		nameSplit := strings.Split(*name, ",")
		for _, s := range nameSplit {
			id := DockerApi.Unpause(*host, strings.TrimSpace(s))
			log.Printf("unpause 容器别名: %s, 容器ID: %s", strings.TrimSpace(s), id)
		}

	case "remove":
		if *name == "" {
			log.Fatal("容器别名不能为空")
		}
		nameSplit := strings.Split(*name, ",")
		for _, s := range nameSplit {
			id := DockerApi.Remove(*host, strings.TrimSpace(s))
			log.Printf("remove 容器别名: %s, 容器ID: %s", strings.TrimSpace(s), id)
		}

	case "list":
		containers := DockerApi.List(*host)
		for i, item := range containers {
			n := ""
			if len(item.Names) > 0 {
				n = strings.ReplaceAll(item.Names[0], "/", "")
			}
			log.Printf("序号%d\n 容器ID: %s\n 容器状态: %s\n 容器镜像: %s\n 容器别名: %s\n \n",
				i+1, item.ID, item.Status, item.Image, n)
		}

	default:
		log.Fatal("command错误")
	}

}
