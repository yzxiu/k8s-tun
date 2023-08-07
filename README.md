# k8s-tun

### Introduction

Access k8s pod ip, service ip

![img.png](img.png)

### Usage

**server**

```bash
kubectl apply -f https://raw.githubusercontent.com/yzxiu/k8s-tun/master/deploy.yaml
```

**client**

Linux & Mac

```shell
# download client
wget https://github.com/yzxiu/k8s-tun/releases/download/0.86-3/client-darwin-amd64-086-3
chmod +x client-linux-amd64-086-3

sudo ./client-linux-amd64-086-3 -s <k8s-node-ip>:30011
```

Windows

download [client-windows-amd64-086-3.exe](https://github.com/yzxiu/k8s-tun/releases/download/0.86-3/client-windows-amd64-086-3.exe)

install the attached tap-windows-9.24.2-I601-Win10 driver

right click `client-windows-amd64-086-3.exe` and run as administrator




Tunnel implementation refers to https://github.com/net-byte/vtun