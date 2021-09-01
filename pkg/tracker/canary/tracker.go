package canary

import (
	"context"
	"fmt"
	"time"

	v1beta1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
	flaggerv1beta1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/flagger/v1beta1"
	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/kubedog/pkg/tracker"
	"github.com/werf/kubedog/pkg/tracker/debug"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"
)

type FailedReport struct {
	FailedReason string
	CanaryStatus CanaryStatus
}

type Tracker struct {
	tracker.Tracker
	LogsFromTime time.Time

	Added     chan CanaryStatus
	Succeeded chan CanaryStatus
	Failed    chan CanaryStatus
	Status    chan CanaryStatus

	EventMsg chan string

	State tracker.TrackerState

	lastObject     *v1beta1.Canary
	failedReason   string
	canaryStatuses map[string]v1beta1.CanaryStatus

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
	fmt.Println("BLABLA")
	for {
		// TODO Rever e tratar as 3 situacoes
		fmt.Println("QQQQQQQ")
		select {
		case object := <-canary.objectAdded:
			if err := canary.handleCanaryState(ctx, object); err != nil {
				return err
			}
		case reason := <-canary.objectFailed:
			canary.State = tracker.ResourceFailed
			canary.failedReason = reason

			var status CanaryStatus

			status = CanaryStatus{IsFailed: true, FailedReason: reason}
			canary.Failed <- status
		case <-ctx.Done():
			if ctx.Err() == context.Canceled {
				return nil
			}
			return ctx.Err()
		case err := <-canary.errors:
			return err
		}

	}
}

func (canary *Tracker) runInformer(ctx context.Context) error {
	// TODO rever
	fmt.Println("AQUI")
	config, err := kube.GetKubeConfig(kube.KubeConfigOptions{})
	if err != nil {
		fmt.Print(err)
	}
	flagger, err := flaggerv1beta1.NewForConfig(config.Config)
	if err != nil {
		fmt.Print(err)
	}
	//rubles, _ := flagger.Canaries("iti-auth-partner").Get(ctx, "iti-auth-partner", v1.GetOptions{})
	// rubles, _ := flagger.Canaries(canary.Namespace).Get(ctx, canary.ResourceName, v1.GetOptions{})
	// fmt.Print(rubles)

	tweakListOptions := func(options metav1.ListOptions) metav1.ListOptions {
		options.FieldSelector = fields.OneTermEqualSelector("metadata.name", canary.ResourceName).String()
		return options
	}
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return flagger.Canaries(canary.Namespace).List(ctx, tweakListOptions(options))
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return flagger.Canaries(canary.Namespace).Watch(ctx, tweakListOptions(options))
		},
	}

	go func() {
		_, err := watchtools.UntilWithSync(ctx, lw, &v1beta1.Canary{}, nil, func(e watch.Event) (bool, error) {
			if debug.Debug() {
				fmt.Printf("Canary `%s` informer event: %#v\n", canary.ResourceName, e.Type)
			}

			var object *v1beta1.Canary

			if e.Type != watch.Error {
				var ok bool
				object, ok = e.Object.(*v1beta1.Canary)
				if !ok {
					return true, fmt.Errorf("expected %s to be a *v1beta1.Canary, got %T", canary.ResourceName, e.Object)
				}
			}

			fmt.Println("e", e.Type)

			if e.Type == watch.Added {
				canary.objectAdded <- object
			} else if e.Type == watch.Modified {
				canary.objectModified <- object
			} else if e.Type == watch.Deleted {
				canary.objectDeleted <- object
			}

			return false, nil
		})

		if err != tracker.AdaptInformerError(err) {
			canary.errors <- fmt.Errorf("canary informer error: %s", err)
		}

		if debug.Debug() {
			fmt.Printf("Canary `%s` informer done\n", canary.ResourceName)
		}
	}()

	return nil
}

func (canary *Tracker) handleCanaryState(ctx context.Context, object *v1beta1.Canary) error {
	canary.lastObject = object
	canary.StatusGeneration++

	//
	status := NewCanaryStatus(object, canary.StatusGeneration, canary.State == tracker.ResourceFailed, canary.failedReason, canary.canaryStatuses)

	fmt.Println("Bla2", canary.State)
	fmt.Println("Bla3", status.IsSucceeded)

	switch canary.State {
	case tracker.Initial:
		// canary.runPodsInformer(ctx, object)
		if status.IsFailed {
			canary.State = tracker.ResourceFailed
			canary.Failed <- status
		} else if status.IsSucceeded {
			fmt.Println("SUCCEEDED2")
			canary.State = tracker.ResourceSucceeded
			canary.Succeeded <- status
		} else {
			fmt.Println("ELSE")
			canary.State = tracker.ResourceAdded
			canary.Succeeded <- status
		}
	case tracker.ResourceAdded, tracker.ResourceFailed:
		fmt.Println("CASENOVO")
		if status.IsFailed {
			canary.State = tracker.ResourceFailed
			canary.Failed <- status
		} else if status.IsSucceeded {
			canary.State = tracker.ResourceSucceeded
			canary.Succeeded <- status
		} else {
			canary.Status <- status
		}
	case tracker.ResourceSucceeded:
		fmt.Println("SUCCEEDED")
		canary.State = tracker.ResourceSucceeded
		canary.Succeeded <- status
	}

	// switch job.State {
	// case tracker.Initial:
	// 	job.runPodsInformer(ctx, object)

	// 	if os.Getenv("KUBEDOG_DISABLE_EVENTS") != "1" {
	// 		job.runEventsInformer(ctx, object)
	// 	}

	// 	if status.IsFailed {
	// 		job.State = tracker.ResourceFailed
	// 		job.Failed <- status
	// 	} else if status.IsSucceeded {
	// 		job.State = tracker.ResourceSucceeded
	// 		job.Succeeded <- status
	// 	} else {
	// 		job.State = tracker.ResourceAdded
	// 		job.Added <- status
	// 	}
	// case tracker.ResourceAdded, tracker.ResourceFailed:
	// 	if status.IsFailed {
	// 		job.State = tracker.ResourceFailed
	// 		job.Failed <- status
	// 	} else if status.IsSucceeded {
	// 		job.State = tracker.ResourceSucceeded
	// 		job.Succeeded <- status
	// 	} else {
	// 		job.Status <- status
	// 	}
	// case tracker.ResourceSucceeded:
	// 	job.Status <- status
	// case tracker.ResourceDeleted:
	// 	if status.IsFailed {
	// 		job.State = tracker.ResourceFailed
	// 		job.Failed <- status
	// 	} else if status.IsSucceeded {
	// 		job.State = tracker.ResourceSucceeded
	// 		job.Succeeded <- status
	// 	} else {
	// 		job.State = tracker.ResourceAdded
	// 		job.Added <- status
	// 	}
	// }

	return nil
}
