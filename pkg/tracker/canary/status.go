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

	Duration string
	Age      string

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
		StatusIndicator:  &indicators.StringEqualConditionIndicator{},
		Age:              utils.TranslateTimestampSince(object.CreationTimestamp),
	}

	// TODO validar status

	if object.Status.Phase == v1beta1.CanaryPhaseSucceeded {
		res.IsSucceeded = true
	} else {
		// res.StatusIndicator =
		fmt.Println("ASD", object.Status)
		fmt.Println("ASD", res.StatusIndicator)
		res.StatusIndicator.Value = string(object.Status.Phase)
		// res.IsSucceeded = true
	}

	fmt.Println("BLA", res.IsSucceeded)

	// if !res.IsSucceeded && !res.IsFailed {
	// 	res.IsFailed = isTrackerFailed
	// 	res.FailedReason = trackerFailedReason
	// }

	return res
}
