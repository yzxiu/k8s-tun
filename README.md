# 访问k8s pod ip,service ip

## 服务端

```bash
k apply -f https://raw.githubusercontent.com/yzxiu/k8s-tun/master/deploy.yaml
```

## 客户端
#### Linux & Mac

1. 设置权限

   ```shell
   chmod +x client-linux-amd64
   ```

2. 运行

   ```shell
   sudo ./client-linux-amd64 -s <node ip>:<node port>
   ```

3. 退出

   ctrl + c

#### Windows

1. 安装驱动

   安装附带的tap-windows-9.24.2-I601-Win10驱动

2. 右键，以管理员身份运行

3. 退出

   关闭窗口 或者 ctrl + c