package controller

import (
	"context"
	"errors"
	"time"

	"github.com/gotway/gotway/pkg/log"

	wrv1alpha1 "github.com/hamedetemaad/lineq-operator/pkg/waitingroom/v1alpha1"
	wrv1alpha1clientset "github.com/hamedetemaad/lineq-operator/pkg/waitingroom/v1alpha1/apis/clientset/versioned"
	wrinformers "github.com/hamedetemaad/lineq-operator/pkg/waitingroom/v1alpha1/apis/informers/externalversions"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	kubeClientSet kubernetes.Interface

	wrInformer  cache.SharedIndexInformer
	ingInformer cache.SharedIndexInformer

	queue workqueue.RateLimitingInterface

	namespace string

	logger log.Logger
}

func (c *Controller) Run(ctx context.Context, numWorkers int) error {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("starting controller")

	c.logger.Info("starting informers")
	for _, i := range []cache.SharedIndexInformer{
		c.wrInformer,
		c.ingInformer,
	} {
		go i.Run(ctx.Done())
	}

	c.logger.Info("waiting for informer caches to sync")
	if !cache.WaitForCacheSync(ctx.Done(), []cache.InformerSynced{
		c.wrInformer.HasSynced,
		c.ingInformer.HasSynced,
	}...) {
		err := errors.New("failed to wait for informers caches to sync")
		utilruntime.HandleError(err)
		return err
	}

	c.logger.Infof("starting %d workers", numWorkers)
	for i := 0; i < numWorkers; i++ {
		go wait.Until(func() {
			c.runWorker(ctx)
		}, time.Second, ctx.Done())
	}
	c.logger.Info("controller ready")

	<-ctx.Done()
	c.logger.Info("stopping controller")

	return nil
}

func (c *Controller) addWaitingRoom(obj interface{}) {
	c.logger.Debug("adding waiting room")
	wr, ok := obj.(*wrv1alpha1.WaitingRoom)
	if !ok {
		c.logger.Errorf("unexpected object %v", obj)
		return
	}
	c.queue.Add(event{
		eventType: addWaitingRoom,
		newObj:    wr.DeepCopy(),
	})
}

func New(
	kubeClientSet kubernetes.Interface,
	wrClientSet wrv1alpha1clientset.Interface,
	namespace string,
	logger log.Logger,
) *Controller {

	wrInformerFactory := wrinformers.NewSharedInformerFactory(
		wrClientSet,
		10*time.Second,
	)
	wrInformer := wrInformerFactory.Lineq().V1alpha1().WaitingRooms().Informer()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClientSet, 10*time.Second)
	ingInformer := kubeInformerFactory.Networking().V1().Ingresses().Informer()

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	ctrl := &Controller{
		kubeClientSet: kubeClientSet,

		wrInformer:  wrInformer,
		ingInformer: ingInformer,

		queue: queue,

		namespace: namespace,

		logger: logger,
	}

	wrInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctrl.addWaitingRoom,
	})

	return ctrl
}
