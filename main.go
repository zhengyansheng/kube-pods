package main

import (
	"flag"
	"time"

	"github.com/zhengyansheng/kube-pods/pkg/client"
	"github.com/zhengyansheng/kube-pods/pkg/factory"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

var (
	kubeconfig    *string = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	defaultResync         = time.Minute * 5
)

func initGvr() (gvrs []schema.GroupVersionResource) {
	gvrs = append(gvrs, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"})
	gvrs = append(gvrs, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"})
	return gvrs
}

func main() {
	flag.Parse()

	stopCh := make(chan struct{}, 0)

	clientSet, err := client.GetKubernetesClient(kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}

	f := factory.NewInformerFactory(clientSet, defaultResync)
	for _, gvr := range initGvr() {
		go func(gvr schema.GroupVersionResource) {
			klog.Infof("start gvr: %s", gvr.String())
			f.Run(gvr, stopCh)
		}(gvr)
	}

	<-stopCh
}
