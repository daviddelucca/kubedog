package canary

import (
	v1beta1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
	"github.com/werf/kubedog/pkg/tracker/indicators"
)

type CanaryStatus struct {
	v1beta1.CanaryStatus

	StatusGeneration uint64

	StatusIndicator *indicators.StringEqualConditionIndicator
	Duration        string
	Age             string

	WaitingForMessages []string

	IsSucceeded  bool
	IsFailed     bool
	FailedReason string
}

func NewCanaryStatus(object *v1beta1.Canary, statusGeneration uint64, isTrackerFailed bool, trackerFailedReason string, canariesStatuses map[string]v1beta1.CanaryStatus, trackedPodsNames []string) CanaryStatus {
	res := CanaryStatus{
		StatusGeneration: statusGeneration,
		Age:              "100s",
	}

	// doCheckCanaryConditions := true
	// for _, trackedPodName := range trackedPodsNames {
	// 	podStatus := podsStatuses[trackedPodName]

	// 	if !podStatus.IsFailed && !podStatus.IsSucceeded {
	// 		doCheckJobConditions = false // unterminated pods exists
	// 	}
	// }

	return res
}
