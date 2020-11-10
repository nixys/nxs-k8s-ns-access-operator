package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultClusterRoleName = "edit"
)

func main() {

	log.SetOutput(os.Stdout)

	sigs := make(chan os.Signal, 1)
	stop := make(chan struct{})

	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	wg := &sync.WaitGroup{}

	k8sClient, err := newK8sClient()
	if err != nil {
		panic(err.Error())
	}

	// Set cluster role name
	clusterRoleName := os.Getenv("CLUSTER_ROLE_NAME")
	if clusterRoleName == "" {
		clusterRoleName = defaultClusterRoleName
	}

	go nsControllerExec(k8sClient, clusterRoleName, stop, wg)

	<-sigs
	log.Println("Shutting down...")

	close(stop)
	wg.Wait()
}

func newK8sClient() (*kubernetes.Clientset, error) {

	k8sConfig := os.Getenv("KUBECONFIG")

	config, err := clientcmd.BuildConfigFromFlags("", k8sConfig)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}
