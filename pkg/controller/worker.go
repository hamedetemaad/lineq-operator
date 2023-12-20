package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	wrv1alpha1 "github.com/hamedetemaad/lineq-operator/pkg/waitingroom/v1alpha1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const maxRetries = 3

type RequestBody struct {
	Name            string `json:"name"`
	Path            string `json:"path"`
	ActiveUsers     int    `json:"activeUsers"`
	SessionDuration int    `json:"sessionDuration"`
}

type ResponseBody struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

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

func (c *Controller) sendBackendRequest(wr *wrv1alpha1.WaitingRoom, name string) {

	url := "http://lineq-http.lineq.svc:8060/create"

	requestBody := RequestBody{
		Name:            name,
		Path:            wr.Spec.Path,
		ActiveUsers:     wr.Spec.ActiveUsers,
		SessionDuration: wr.Spec.SessionDuration,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	response, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer response.Body.Close()

	var responseBody ResponseBody
	err = json.NewDecoder(response.Body).Decode(&responseBody)
	if err != nil {
		fmt.Println("Error decoding JSON response:", err)
		return
	}

	fmt.Printf("Response Status: %s\n", responseBody.Status)
	fmt.Printf("Response Message: %s\n", responseBody.Message)
}

func (c *Controller) createName(wr *wrv1alpha1.WaitingRoom) string {
	domain := strings.Replace(wr.Spec.Host, ".", "_", -1)
	path := strings.Replace(wr.Spec.Path, "/", "_", -1)
	name := fmt.Sprintf("%s%s", domain, path)

	return name
}

func (c *Controller) editHAProxyAuxConfigMap(wr *wrv1alpha1.WaitingRoom, name string) {
	auxCm, err := c.kubeClientSet.CoreV1().ConfigMaps("haproxy-controller").Get(context.TODO(), "haproxy-auxiliary-configmap", metav1.GetOptions{})
	fmt.Println(err)
	backend := `

backend %s
  stick-table type string len 36 size 100k expire %dm store gpc1 peers lineq

`
	backend = fmt.Sprintf(backend, name, wr.Spec.SessionDuration)

	auxCm.Data["haproxy-auxiliary.cfg"] += backend
	c.kubeClientSet.CoreV1().ConfigMaps("haproxy-controller").Update(context.TODO(), auxCm, metav1.UpdateOptions{})
}

func (c *Controller) processAddWaitingRoom(ctx context.Context, wr *wrv1alpha1.WaitingRoom) error {
	name := c.createName(wr)
	c.sendBackendRequest(wr, name)

	ing := createIngress(wr, wr.Namespace)
	exists, err := resourceExists(ing, c.ingInformer.GetIndexer())
	if err != nil {
		return fmt.Errorf("error checking ingress existence %v", err)
	}
	if exists {
		c.logger.Debug("ingress already exists, skipping")
		return nil
	}

	_, err = c.kubeClientSet.NetworkingV1().
		Ingresses(wr.Namespace).
		Create(ctx, ing, metav1.CreateOptions{})

	c.editHAProxyAuxConfigMap(wr, name)
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
