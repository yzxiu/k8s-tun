package info

import (
	"context"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"net"
	"strings"
	"time"
)

func GetNodeList(kubeClient *kubernetes.Clientset) {

}

// GetSvcCIDR 获取 service-cluster-ip-range
func GetSvcCIDR(kubeClient *kubernetes.Clientset) string {
	return getConfigureFromPod(kubeClient, "kube-system", map[string]string{
		"component": "kube-apiserver",
	}, "service-cluster-ip-range")
}

func GetPodCIDR(dyClient dynamic.Interface, kubeClient *kubernetes.Clientset) string {
	// try get from ippools.crd.projectcalico.org
	var ippoolGVR = schema.GroupVersionResource{
		Group:    "crd.projectcalico.org",
		Version:  "v1",
		Resource: "ippools",
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ippools, err := dyClient.Resource(ippoolGVR).List(ctx, v1.ListOptions{})
	if err != nil || len(ippools.Items) == 0 {
	} else {
		var podcidrs []string
		if ls, err := ippools.MarshalJSON(); err == nil {
			var lss IPPoolList
			if err := json.Unmarshal(ls, &lss); err == nil {
				for _, ippool := range lss.Items {
					_, _, err = net.ParseCIDR(ippool.Spec.CIDR)
					if err == nil {
						podcidrs = append(podcidrs, ippool.Spec.CIDR)
					}
				}
			}
		}
		if len(podcidrs) > 0 {
			return strings.Join(podcidrs, ",")
		}
	}

	// try get form configmap kubeadm-config

	// try get from apiserver
	return getConfigureFromPod(kubeClient, "kube-system", map[string]string{
		"component": "kube-controller-manager",
	}, "cluster-cidr")

}

func getConfigureFromPod(kubeClient *kubernetes.Clientset, ns string, label map[string]string, configure string) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	pods, err := kubeClient.CoreV1().Pods(ns).List(ctx, v1.ListOptions{
		LabelSelector: labels.SelectorFromSet(label).String(),
	})
	if err != nil || len(pods.Items) == 0 {
		return ""
	}
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			for _, s := range container.Command {
				if strings.Contains(s, configure) {
					if len(strings.Split(s, "=")) == 2 {
						v := strings.Split(s, "=")[1]
						_, _, err = net.ParseCIDR(v)
						if err == nil {
							return v
						}
					}
				}
			}
		}
	}
	return ""
}

type TunServer struct {
	Host string
	Port string
}

func GetTunServer(kubeClient *kubernetes.Clientset) (*TunServer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	pods, err := kubeClient.CoreV1().Pods("k8s-tun").List(ctx, v1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app": "k8s-tun",
		}).String(),
	})
	if err != nil {
		return nil, err
	}
	count := 0
	host := ""
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning {
			host = pod.Status.HostIP
			count++
		}
	}
	if count != 1 {
		return nil, fmt.Errorf("tun server is restarting, please try again later")
	}
	svc, err := kubeClient.CoreV1().Services("k8s-tun").Get(ctx, "k8s-tun", v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	port := GetWsPort(svc)
	if len(port) < 1 {
		return nil, fmt.Errorf("get tun server port error")
	}
	return &TunServer{
		Host: host,
		Port: port,
	}, nil
}

func GetWsPort(svc *corev1.Service) string {
	for _, port := range svc.Spec.Ports {
		if port.Name == "ws" {
			return fmt.Sprint(port.NodePort)
		}
	}
	return ""
}
