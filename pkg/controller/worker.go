package controller

import (
	"context"
	"fmt"

	wrv1alpha1 "github.com/hamedetemaad/lineq-operator/pkg/waitingroom/v1alpha1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const maxRetries = 3

func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextItem(ctx) {
	}
}

func (c *Controller) processNextItem(ctx context.Context) bool {
	obj, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(obj)

	err := c.processEvent(ctx, obj)
	if err == nil {
		c.logger.Debug("processed item")
		c.queue.Forget(obj)
	} else if c.queue.NumRequeues(obj) < maxRetries {
		c.logger.Errorf("error processing event: %v, retrying", err)
		c.queue.AddRateLimited(obj)
	} else {
		c.logger.Errorf("error processing event: %v, max retries reached", err)
		c.queue.Forget(obj)
		utilruntime.HandleError(err)
	}

	return true
}

func (c *Controller) processEvent(ctx context.Context, obj interface{}) error {
	event, ok := obj.(event)
	if !ok {
		c.logger.Error("unexpected event ", obj)
		return nil
	}
	switch event.eventType {
	case addWaitingRoom:
		return c.processAddWaitingRoom(ctx, event.newObj.(*wrv1alpha1.WaitingRoom))
	}
	return nil
}

func (c *Controller) processAddWaitingRoom(ctx context.Context, wr *wrv1alpha1.WaitingRoom) error {
	ing := createIngress(wr, c.namespace)
	cm, err := c.kubeClientSet.CoreV1().ConfigMaps("haproxy-controller").Get(context.TODO(), "haproxy-kubernetes-ingress", metav1.GetOptions{})
	auxCm, err := c.kubeClientSet.CoreV1().ConfigMaps("haproxy-controller").Get(context.TODO(), "haproxy-auxiliary-configmap", metav1.GetOptions{})
	editHAProxyConfigMap(cm)
	editHAProxyAuxConfigMap(auxCm)
	exists, err := resourceExists(ing, c.ingInformer.GetIndexer())
	if err != nil {
		return fmt.Errorf("error checking ingress existence %v", err)
	}
	if exists {
		c.logger.Debug("ingress already exists, skipping")
		return nil
	}

	_, err = c.kubeClientSet.NetworkingV1().
		Ingresses(c.namespace).
		Create(ctx, ing, metav1.CreateOptions{})
	return err
}

func resourceExists(obj interface{}, indexer cache.Indexer) (bool, error) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		return false, fmt.Errorf("error getting key %v", err)
	}
	_, exists, err := indexer.GetByKey(key)
	return exists, err
}
