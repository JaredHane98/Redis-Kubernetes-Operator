package result

import (
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func Ok() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func RequeueAfter(duration time.Duration) (reconcile.Result, error) {
	return reconcile.Result{Requeue: true, RequeueAfter: duration}, nil
}

func RequeueAfterWithMessage(duration time.Duration, logger logr.Logger, message string, keysAndValues ...interface{}) (reconcile.Result, error) {
	if len(message) == 0 {
		message = "requeing after"
	}
	logger.Info(message, keysAndValues...)
	return RequeueAfter(duration)
}

func RetryWithError(err error, logger logr.Logger, message string, keysAndValues ...interface{}) (reconcile.Result, error) {
	if len(message) == 0 {
		message = "retrying with error"
	}
	logger.Error(err, message, keysAndValues...)
	return reconcile.Result{Requeue: true}, err
}

func FailedWithError(err error, logger logr.Logger, message string, keysAndValues ...interface{}) (reconcile.Result, error) {
	if len(message) == 0 {
		message = "failed with error"
	}
	logger.Error(err, message, keysAndValues...)
	return RequeueAfter(0)
}

func ReconciledWithMessage(logger logr.Logger, message string, keysAndValues ...interface{}) (reconcile.Result, error) {
	if len(message) == 0 {
		message = "retrying with message"
	}
	logger.Info(message, keysAndValues...)
	return Ok()
}
