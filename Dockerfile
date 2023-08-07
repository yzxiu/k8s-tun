FROM alpine:3.15.4
RUN apk update && apk add iptables tcpdump
RUN echo "net.ipv4.ip_forward = 1" >> /etc/sysctl.conf
WORKDIR /
COPY bin/vtun-linux-amd64 .
USER 65532:65532
ENTRYPOINT ["/vtun-linux-amd64"]


# iptables -t nat -A POSTROUTING -s 10.99.99.0/24 -j MASQUERADE
# docker exec -t vtun-server iptables -t nat -A POSTROUTING -s 10.99.99.0/24 -j MASQUERADE

# windows route
# route add 10.233.0.0 MASK 255.255.0.0 10.99.99.1

# macos route
# route -n add -net 10.233.0.0 -netmask 255.255.0.0 10.99.99.1

# linux route
# ip route add 10.233.0.0/16 via 10.99.99.1

# client
# -l=:3000 -s=192.168.242.77:3001 -c=10.99.99.11/24 -k=123456 -p tcp