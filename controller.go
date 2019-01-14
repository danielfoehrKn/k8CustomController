package main

import (
	"fmt"
	"time"

	"github.com/golang/glog"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	//has stuff e.g structs needed for the definiton of the FooType (all obligatory fields in go)
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clientset "github.com/danielfoehrkn/custom-database-controller/pkg/client/clientset/versioned"
	samplescheme "github.com/danielfoehrkn/custom-database-controller/pkg/client/clientset/versioned/scheme"
	informers "github.com/danielfoehrkn/custom-database-controller/pkg/client/informers/externalversions"
	listers "github.com/danielfoehrkn/custom-database-controller/pkg/client/listers/danielfoehrkn.com/v1"

	danielfoehrknApiV1 "github.com/danielfoehrkn/custom-database-controller/pkg/apis/danielfoehrkn.com/v1"
)

const controllerAgentName = "tagger-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Database is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Database fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by TAGGER"
	// MessageResourceSynced is the message used for an Event fired when a Database
	// is synced successfully
	MessageResourceSynced = "Tagger synced successfully"
)

// Controller is the controller implementation for Database resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// exampleclientset is a clientset for our own API group
	danielfoehrknclientset clientset.Interface

	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
	taggerLister      listers.TaggerLister
	taggerSynced      cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	danielfoehrknclientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	exampleInformerFactory informers.SharedInformerFactory) *Controller {

	// obtain references to shared index informers for the Deployment and Database
	// types.
	deploymentInformer := kubeInformerFactory.Apps().V1().Deployments()
	taggerInformer := exampleInformerFactory.Danielfoehrkn().V1().Taggers()

	// Create event broadcaster
	// Add example database types to the default Kubernetes Scheme so Events can be
	// logged for database types.
	samplescheme.AddToScheme(scheme.Scheme)
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:          kubeclientset,
		danielfoehrknclientset: danielfoehrknclientset,
		deploymentsLister:      deploymentInformer.Lister(),
		deploymentsSynced:      deploymentInformer.Informer().HasSynced,
		taggerLister:           taggerInformer.Lister(),
		taggerSynced:           taggerInformer.Informer().HasSynced,
		workqueue:              workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), ""),
		recorder:               recorder,
	}

	glog.Info("Setting up event handlers")
	// Set up an event handler for when Tagger resources change
	taggerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueTagger,
		UpdateFunc: func(old, new interface{}) {
			newTagger := new.(*danielfoehrknApiV1.Tagger)
			oldTagger := old.(*danielfoehrknApiV1.Tagger)
			if newTagger.ResourceVersion == oldTagger.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.

				//TODO: why is this event being called ALL THE TIME so that I need to make a diff on the resource version from etcd? -> etcd watch should only give updates once the key changed
				glog.Info("EVENT FOR UNCHANGED TAGGER RECEIVED")
				return
			}
			controller.enqueueTagger(new)
		},
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting Tagger controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced, c.taggerSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	// Launch two workers to process Tagger resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string (format: nampesace/name) in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Tagger resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Database resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Tagger resource with this namespace/name

	tagger, err := c.taggerLister.Taggers(namespace).Get(name)
	if err != nil {
		// The tagger resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("tagger '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	glog.Infof("Tagger resource retrieved %s:%s", tagger.Namespace, tagger.Name)

	//check if needed fields in the tagger APi Object have been specified by the user

	apiObjectLabel := tagger.Spec.Label
	apiObjectKey := tagger.Spec.Key
	if apiObjectLabel == "" || apiObjectKey == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		runtime.HandleError(fmt.Errorf("%s: label and key must be specified", apiObjectLabel))
		return nil
	}

	//tag the pods with that label
	options := v1.ListOptions{}

	podList, err := c.kubeclientset.Core().Pods(tagger.Namespace).List(options)

	if err != nil {
		glog.Infof("Pod listing failed %s", err.Error)
		return err
	}

	glog.Infof("Pods retrieved: %d", len(podList.Items))

	for _, pod := range podList.Items {
		//tag it
		podCopy := pod.DeepCopy()

		if _, ok := podCopy.Labels[apiObjectKey]; !ok {
			glog.Infof("The Pod %s does not have the tag %s yet", pod.Name, apiObjectKey)

			//Add label
			labelMap := make(map[string]string)

			//use label from user-created tagger Object
			labelMap[apiObjectKey] = apiObjectLabel

			for index, tag := range labelMap {
				podCopy.Labels[index] = tag
			}

			//tag it
			_, err := c.kubeclientset.Core().Pods(tagger.Namespace).Update(podCopy)
			// If an error occurs during Get/Create, we'll requeue the item so we can
			// attempt processing again later. This could have been caused by a
			// temporary network failure, or any other transient reason.
			if err != nil {
				glog.Infof("Tagging of Pod %s failed", podCopy.Name)
				return err
			}
			glog.Infof("Tagging of Pod %s successful", podCopy.Name)
		}

	}

	c.recorder.Event(tagger, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

// enqueueTagger takes a Tagger resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Database.
func (c *Controller) enqueueTagger(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	glog.Infof("Enqueueing tagger  %s", key)
	c.workqueue.AddRateLimited(key)
}
