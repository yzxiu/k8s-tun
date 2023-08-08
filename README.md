# k8s-tun

### Introduction

Access k8s pod ip, service ip

![img.png](img.png)

### Usage

**server**

```bash
kubectl apply -f https://raw.githubusercontent.com/yzxiu/k8s-tun/master/deploy.yaml
```

view the deployment
```log
~ kubectl get pod -n k8s-tun -o wide
NAME                       READY   STATUS    RESTARTS       AGE     IP               NODE          
k8s-tun-54b5865fdc-sz48s   2/2     Running   15 (72m ago)   4d20h   10.233.114.158   ubuntu2004-34 
```

**client**

Linux & Mac

```shell
# download client
wget https://github.com/yzxiu/k8s-tun/releases/download/0.86-3/client-darwin-amd64-086-3
chmod +x client-linux-amd64-086-3

# start client
sudo ./client-linux-amd64-086-3 -s <k8s-node-ip>:30011
```
or, use docker to start the client
```bash
docker run --rm -it --name k8s-tun-client \
  --network=host --cap-add NET_ADMIN \
  --device=/dev/net/tun \
  q946666800/k8s-tun-client:0.86 \
  -s <k8s-node-ip>:30011
```

when the client starts successfully, you can directly access the pod ip & svc ip
```log
~ ping 10.233.114.158
PING 10.233.114.158 (10.233.114.158) 56(84) bytes of data.
64 bytes from 10.233.114.158: icmp_seq=2 ttl=64 time=75.4 ms
64 bytes from 10.233.114.158: icmp_seq=3 ttl=64 time=50.0 ms
64 bytes from 10.233.114.158: icmp_seq=5 ttl=64 time=49.6 ms
```

Windows (alpha)
```shell
# 1
download [client-windows-amd64-086-3.exe](https://github.com/yzxiu/k8s-tun/releases/download/0.86-3/client-windows-amd64-086-3.exe)
# 2
install the attached tap-windows-9.24.2-I601-Win10 driver
# 3
right click `client-windows-amd64-086-3.exe` and run as administrator
```

### Uninstall
```bash
kubectl delete -f https://raw.githubusercontent.com/yzxiu/k8s-tun/master/deploy.yaml
```

### Notice


Advantages: In theory, all cni plug-ins are supported, and vm-01 can be in a different network from the k8s cluster, so it is more flexible to use. The client does not need to configure kubeconfig

Cons: Traffic is tunneled (similar to openvpn), less efficient.

Tunnel implementation refers to https://github.com/net-byte/vtun