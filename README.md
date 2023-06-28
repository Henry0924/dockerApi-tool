# dockerApi-tool

1.在config.json文件里设置镜像地址
    配置文件参数说明：

        image - 镜像地址
        bridgedNetworkMode - 是否开启桥接网络模式，默认为false，开启则填true，下面的参数均为开启桥接网络模式后用到的参数
        androidHost - 桥接网络模式下启动的安卓的ip
        dns1 - dns1
        dns2 - dns2
        gateway - 网关
        subnet - 子网，例如192.168.100.0/24， 192.168.0.0/16

2.启动参数说明，-help可查看参数

        -command string
        操作指令, create--创建容器, start--启动容器, stop--停止容器, pause--暂停容器, unpause--取消暂停, 
        remove--删除容器, list--容器列表 (default "list")

        -host string
        主机host (default "192.168.30.3")

        -i string
        容器序号1-12, 创建容器时必传, 例1、2、3 (default "1")

        -name string
        容器别名, 例001、001,002、001,002,003

3.创建容器

        .\tool.exe -command=create -i=1 -host=192.168.100.10 -name=t001

4.启动容器

        .\tool.exe -command=start -host=192.168.100.10 -name=t001

5.停止容器

        .\tool.exe -command=stop -host=192.168.100.10 -name=t001

6.删除容器

        .\tool.exe -command=remove -host=192.168.100.10 -name=t001

7.容器列表

        .\tool.exe -command=list -host=192.168.100.10
