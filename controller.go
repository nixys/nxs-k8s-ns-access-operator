package main

import (
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type nsController struct {
	nsInformer      cache.SharedIndexInformer
	k8sClient       *kubernetes.Clientset
	clusterRoleName string
}

func nsControllerExec(k8sClient *kubernetes.Clientset, clusterRoleName string, stopCh <-chan struct{}, wg *sync.WaitGroup) {

	nsWatcher := &nsController{}

	nsInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return k8sClient.CoreV1().Namespaces().List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return k8sClient.CoreV1().Namespaces().Watch(options)
			},
		},
		&v1.Namespace{},
		1*time.Minute,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	nsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: nsWatcher.createRoleBinding,
	})

	nsWatcher.k8sClient = k8sClient
	nsWatcher.nsInformer = nsInformer
	nsWatcher.clusterRoleName = clusterRoleName

	nsWatcher.run(stopCh, wg)
}

func (c *nsController) run(stopCh <-chan struct{}, wg *sync.WaitGroup) {

	defer wg.Done()

	wg.Add(1)

	go c.nsInformer.Run(stopCh)

	<-stopCh
}

func (c *nsController) createRoleBinding(obj interface{}) {

	nsObj := obj.(*v1.Namespace)
	nsName := nsObj.Name

	r := regexp.MustCompile("^(.*)-msvc-.*$")
	result := r.FindStringSubmatch(nsName)

	if result != nil {

		if len(result) < 2 {
			log.Printf("Wrong ns regular expression result: can't extract username from ns name")
			return
		}

		roleBinding := &rbacv1.RoleBinding{
			TypeMeta: metav1.TypeMeta{
				Kind:       "RoleBinding",
				APIVersion: "rbac.authorization.k8s.io/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("auto-ns-edit"),
				Namespace: nsName,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind: "User",
					Name: result[1],
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     c.clusterRoleName,
			},
		}

		_, err := c.k8sClient.RbacV1().RoleBindings(nsName).Create(roleBinding)
		if err != nil {
			log.Printf("Can't create RoleBinding '%s' in namespace '%s': %s", roleBinding.Name, nsName, err.Error())
			return
		}

		log.Printf("Created RoleBinding '%s' in Namespace: %s", roleBinding.Name, nsName)
	}
}
