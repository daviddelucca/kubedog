package canary

import (
	"context"
	"fmt"
	"time"

	v1beta1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
	scheme "github.com/fluxcd/flagger/pkg/client/clientset/versioned/scheme"
	"github.com/werf/kubedog/pkg/tracker"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type FailedReport struct {
	FailedReason string
	CanaryStatus CanaryStatus
}

type Tracker struct {
	tracker.Tracker
	LogsFromTime time.Time

	Succeeded chan CanaryStatus
	Failed    chan CanaryStatus
	Status    chan CanaryStatus

	EventMsg chan string

	State tracker.TrackerState

	failedReason string

	errors chan error

	objectAdded    chan *v1beta1.Canary
	objectModified chan *v1beta1.Canary
	objectDeleted  chan *v1beta1.Canary
	objectFailed   chan string
}

func NewTracker(name, namespace string, kube kubernetes.Interface, opts tracker.Options) *Tracker {
	return &Tracker{
		Tracker: tracker.Tracker{
			Kube:             kube,
			Namespace:        namespace,
			FullResourceName: fmt.Sprintf("job/%s", name),
			ResourceName:     name,
			LogsFromTime:     opts.LogsFromTime,
		},

		Succeeded: make(chan CanaryStatus, 0),
		Failed:    make(chan CanaryStatus, 0),
		Status:    make(chan CanaryStatus, 100),

		EventMsg: make(chan string, 1),

		State: tracker.Initial,

		objectAdded:    make(chan *v1beta1.Canary, 0),
		objectModified: make(chan *v1beta1.Canary, 0),
		objectDeleted:  make(chan *v1beta1.Canary, 0),
		objectFailed:   make(chan string, 1),
		errors:         make(chan error, 0),
	}
}

func (canary *Tracker) Track(ctx context.Context) error {
	var err error

	err = canary.runInformer(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case reason := <-canary.objectFailed:
			canary.State = tracker.ResourceFailed
			canary.failedReason = reason

			var status CanaryStatus

			status = CanaryStatus{IsFailed: true, FailedReason: reason}
			canary.Failed <- status
		}
	}
}

func (canary *Tracker) runInformer(ctx context.Context) error {
	// TODO rever
	fmt.Println("AQUI")
	result := &v1beta1.Canary{}
	err := canary.Kube.CoreV1().RESTClient().
		Get().
		Namespace("iti-auth-partner").
		Resource("canaries").
		Name("iti-auth-partner").
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	fmt.Println(err)
	// cfg, err := clientcmd.BuildConfigFromFlags("", "/Users/luccada/.kube/config")

	// if err != nil {
	// 	fmt.Println("Error building kubeconfig: %v", err)
	// }

	// flagger, err := clientset.NewForConfig(cfg)
	// if err != nil {
	// 	fmt.Println("Error building flagger clientset: %s", err.Error())
	// }

	// tweakListOptions := func(options metav1.ListOptions) metav1.ListOptions {
	// 	options.FieldSelector = fields.OneTermEqualSelector("metadata.name", canary.ResourceName).String()
	// 	return options
	// }
	// lw := &cache.ListWatch{
	// 	ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
	// 		return flagger.FlaggerV1beta1().Canaries(canary.Namespace).List(ctx, tweakListOptions(options))
	// 	},
	// 	WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
	// 		return flagger.FlaggerV1beta1().Canaries(canary.Namespace).Watch(ctx, tweakListOptions(options))
	// 	},
	// }

	// // opts := &v1.GetOptions{}
	// // flagger.FlaggerV1beta1().Canaries(canary.Namespace).Get(ctx, "iti-auth-partner", opts)

	// go func() {
	// 	_, err := watchtools.UntilWithSync(ctx, lw, &v1beta1.Canary{}, nil, func(e watch.Event) (bool, error) {
	// 		if debug.Debug() {
	// 			fmt.Printf("Canary `%s` informer event: %#v\n", canary.ResourceName, e.Type)
	// 		}

	// 		var object *v1beta1.Canary

	// 		if e.Type != watch.Error {
	// 			var ok bool
	// 			object, ok = e.Object.(*v1beta1.Canary)
	// 			if !ok {
	// 				return true, fmt.Errorf("expected %s to be a *v1beta1.Canary, got %T", canary.ResourceName, e.Object)
	// 			}
	// 		}

	// 		if e.Type == watch.Added {
	// 			canary.objectAdded <- object
	// 		} else if e.Type == watch.Modified {
	// 			canary.objectModified <- object
	// 		} else if e.Type == watch.Deleted {
	// 			canary.objectDeleted <- object
	// 		}

	// 		return false, nil
	// 	})

	// 	if err != tracker.AdaptInformerError(err) {
	// 		canary.errors <- fmt.Errorf("canary informer error: %s", err)
	// 	}

	// 	if debug.Debug() {
	// 		fmt.Printf("Canary `%s` informer done\n", canary.ResourceName)
	// 	}
	// }()

	return nil
}
