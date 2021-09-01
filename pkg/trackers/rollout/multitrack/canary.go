package multitrack

import (
	"fmt"

	"github.com/werf/kubedog/pkg/tracker/canary"
	"k8s.io/client-go/kubernetes"
)

func (mt *multitracker) TrackCanary(kube kubernetes.Interface, spec MultitrackSpec, opts MultitrackOptions) error {
	fmt.Println("1")
	feed := canary.NewFeed()
	fmt.Println("2")

	feed.OnSucceeded(func() error {
		fmt.Println("onSucceeded")
		mt.mux.Lock()
		defer mt.mux.Unlock()

		mt.CanariesStatuses[spec.ResourceName] = feed.GetStatus()

		return mt.canarySucceeded(spec, feed)
	})
	feed.OnFailed(func(reason string) error {
		fmt.Println("OnFailed")
		mt.mux.Lock()
		defer mt.mux.Unlock()

		mt.CanariesStatuses[spec.ResourceName] = feed.GetStatus()

		return mt.canaryFailed(spec, feed, reason)
	})
	feed.OnEventMsg(func(msg string) error {
		fmt.Println("onEventMsg")
		mt.mux.Lock()
		defer mt.mux.Unlock()

		mt.CanariesStatuses[spec.ResourceName] = feed.GetStatus()

		return mt.canaryEventMsg(spec, feed, msg)
	})

	feed.OnStatus(func(status canary.CanaryStatus) error {
		fmt.Println("onStatus")
		mt.mux.Lock()
		defer mt.mux.Unlock()

		mt.CanariesStatuses[spec.ResourceName] = status

		return nil
	})

	return feed.Track(spec.ResourceName, spec.Namespace, kube, opts.Options)
}

func (mt *multitracker) canarySucceeded(spec MultitrackSpec, feed canary.Feed) error {
	mt.displayResourceTrackerMessageF("canary", spec, "succeeded")

	return mt.handleResourceReadyCondition(mt.TrackingCanaries, spec)
}

func (mt *multitracker) canaryFailed(spec MultitrackSpec, feed canary.Feed, reason string) error {
	mt.displayResourceErrorF("canary", spec, "%s", reason)

	return mt.handleResourceFailure(mt.TrackingCanaries, "deploy", spec, reason)
}

func (mt *multitracker) canaryEventMsg(spec MultitrackSpec, feed canary.Feed, msg string) error {
	mt.displayResourceEventF("canary", spec, "%s", msg)
	return nil
}
