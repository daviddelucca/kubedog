package canary

import (
	"fmt"

	v1beta1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
	"github.com/werf/kubedog/pkg/tracker/indicators"
	"github.com/werf/kubedog/pkg/utils"
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

func NewCanaryStatus(object *v1beta1.Canary, statusGeneration uint64, isTrackerFailed bool, trackerFailedReason string, canariesStatuses map[string]v1beta1.CanaryStatus) CanaryStatus {
	fmt.Println("statusGeneration", statusGeneration)
	res := CanaryStatus{
		CanaryStatus:     object.Status,
		StatusGeneration: statusGeneration,
		Age:              utils.TranslateTimestampSince(object.CreationTimestamp),
	}

	// doCheckCanaryConditions := true
	// for _, trackedPodName := range trackedPodsNames {
	// 	podStatus := podsStatuses[trackedPodName]

	// 	if !podStatus.IsFailed && !podStatus.IsSucceeded {
	// 		doCheckJobConditions = false // unterminated pods exists
	// 	}
	// }

	// TODO validar status

	if object.Status.Phase == v1beta1.CanaryPhaseSucceeded {
		res.IsSucceeded = true
	}

	//if object.Status.Phase != "Succeeded" && object.Status.Phase != "Failed" {

	// res.WaitingForMessages = append(res.WaitingForMessages, fmt.Sprintf("rollout %d", res.CanaryWeight))

	// status
	// weight

	fmt.Println("BLA", res.IsSucceeded)

	return res
}
