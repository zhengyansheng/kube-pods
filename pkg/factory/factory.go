package factory

import (
	"time"

	"github.com/zhengyansheng/kube-pods/pkg/models"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	workers = 1
)

type informerFactory struct {
	clientSet     *kubernetes.Clientset
	defaultResync time.Duration
	queue         workqueue.RateLimitingInterface
	indexer       cache.Indexer
	workers       int
}

func NewInformerFactory(clientSet *kubernetes.Clientset, defaultResync time.Duration) *informerFactory {
	return &informerFactory{
		clientSet:     clientSet,
		defaultResync: defaultResync,
		queue:         workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		workers:       workers}
}

func (f *informerFactory) Run(gvr schema.GroupVersionResource, stopCh chan struct{}) {
	// 1. new informer factory
	factory := informers.NewSharedInformerFactory(f.clientSet, f.defaultResync)

	// 2. register gvr
	genericInformer, err := factory.ForResource(gvr)
	if err != nil {
		klog.Error(err)
	}

	// 3. start factory (lw)
	factory.Start(stopCh)

	// 4. wait all cache sync
	factory.WaitForCacheSync(stopCh)

	// 5. 初始化indexer
	f.indexer = genericInformer.Informer().GetIndexer()

	// 6. add event handler
	f.addEventHandler(genericInformer)

	// 7. pop key from queue
	for i := 0; i < f.workers; i++ {
		go wait.Until(f.runWorker, time.Second, stopCh)
	}
}

func (f *informerFactory) addEventHandler(genericInformer informers.GenericInformer) {
	genericInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// obj 是一个pod对象，这里的key是namespace/name
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				// 将key放入队列
				f.queue.Add(key)

			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err == nil {
				// 将key放入队列
				f.queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				// 将key放入队列
				f.queue.Add(key)
			}
		},
	})
}

func (f *informerFactory) runWorker() {
	for f.processNextItem() {
	}
}

func (f *informerFactory) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := f.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer f.queue.Done(key)

	// Invoke the method containing the business logic
	err := f.syncToStdout(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	f.handleErr(err, key)
	return true
}

func (f *informerFactory) syncToStdout(key string) error {
	obj, exists, err := f.indexer.GetByKey(key)
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}
	if !exists {
		namespace, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			klog.Errorf("Invalid resource key: %s", key)
			return nil
		}
		klog.Infof("Delete %v/%v", namespace, name)
		return nil
	}

	switch obj.(type) {
	case *v1.Pod:
		p := obj.(*v1.Pod)
		pod := models.NewPod(p)
		pod.Notify()
		//klog.Infof("Add/Update %v/%v", p.GetNamespace(), p.GetName())
	case *v1.Node:
		p := obj.(*v1.Node)
		klog.Infof("Add/Update %v/%v", p.GetNamespace(), p.GetName())
	}

	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (f *informerFactory) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		f.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if f.queue.NumRequeues(key) < 5 {
		klog.Infof("Error syncing pod %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		f.queue.AddRateLimited(key)
		return
	}

	f.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	klog.Infof("Dropping pod %q out of the queue: %v", key, err)
}
