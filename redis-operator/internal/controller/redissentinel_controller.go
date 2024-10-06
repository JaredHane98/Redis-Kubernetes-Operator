/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	v1 "redis.operator/api/v1"
	"redis.operator/pkg/kube/configmap"
	k8sredis "redis.operator/pkg/redis"
	"redis.operator/pkg/redissentinel"
	"redis.operator/pkg/util/result"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// RedisSentinelReconciler reconciles a RedisSentinel object
type RedisSentinelReconciler struct {
	client.Client
	K8Client  kubernetes.Interface
	Dk8Client dynamic.Interface
	Log       logr.Logger
	Scheme    *runtime.Scheme
}

func (r *RedisSentinelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	instance := &v1.RedisSentinel{}

	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return result.ReconciledWithMessage(reqLogger, "Failed to get instance. Assumming it was deleted")
		}
		return result.FailedWithError(err, reqLogger, "Error reconciling instance")
	}

	if instance.ObjectMeta.GetDeletionTimestamp() != nil {
		if err = r.HandleReplicationFinalizer(ctx, instance, v1.RedisSentinelFinalizer); err != nil {
			return result.RetryWithError(err, reqLogger, "Failed to handle finalizer")
		}
		return result.Ok()
	}

	if err = r.CreateReplicationFinalizer(ctx, instance, v1.RedisSentinelFinalizer); err != nil {
		return result.RetryWithError(err, reqLogger, "Failed to create finalizer")
	}

	if err, critical := r.CreateOrUpdateConfigMap(ctx, instance, reqLogger); err != nil {
		if critical {
			return result.RetryWithError(err, reqLogger, "Failed to create or update configmap")
		} else {
			return result.RequeueAfterWithMessage(1*time.Second, reqLogger, "non-critical error occurred", "error", err)
		}
	}

	if err := r.CreateOrUpdateHeadlessService(ctx, instance, reqLogger); err != nil {
		return result.RetryWithError(err, reqLogger, "Failed creating or updating headless service")
	}

	if err := r.CreateOrUpdateService(ctx, instance, reqLogger); err != nil {
		return result.RetryWithError(err, reqLogger, "Failed creating or updating service")
	}

	if err := r.CreateOrUpdateSentinel(ctx, instance, reqLogger); err != nil {
		return result.RetryWithError(err, reqLogger, "Failed creating or updating sentinel")
	}

	if err := r.CheckSentinelStatus(ctx, instance, reqLogger); err != nil {
		return result.RetryWithError(err, reqLogger, "Failed to check sentinel status")
	}

	if err := r.UpdateSentinelLabels(ctx, instance, reqLogger); err != nil {
		return result.RetryWithError(err, reqLogger, "Failed to update sentinel labels")
	}

	return result.RequeueAfter(1 * time.Second)
}

func (r *RedisSentinelReconciler) CreateReplicationFinalizer(ctx context.Context, instance *v1.RedisSentinel, finalizer string) error {
	if !controllerutil.ContainsFinalizer(instance, finalizer) {
		controllerutil.AddFinalizer(instance, finalizer)
		return r.Client.Update(ctx, instance)
	}
	return nil
}

func (r *RedisSentinelReconciler) HandleReplicationFinalizer(ctx context.Context, instance *v1.RedisSentinel, finalizer string) error {
	if controllerutil.ContainsFinalizer(instance, finalizer) {
		controllerutil.RemoveFinalizer(instance, finalizer)
		return r.Client.Update(ctx, instance)
	}
	return nil
}

func (r *RedisSentinelReconciler) UpdateSentinelLabels(ctx context.Context, instance *v1.RedisSentinel, reqLogger logr.Logger) error {
	labels := redissentinel.GetSentinelServiceLabels(instance)
	selector := make([]string, 0, len(labels))
	for key, value := range labels {
		selector = append(selector, fmt.Sprintf("%s=%s", key, value))
	}

	podList, err := r.K8Client.CoreV1().Pods(instance.Namespace).List(ctx, metav1.ListOptions{LabelSelector: strings.Join(selector, ",")})
	if err != nil {
		return err
	}

	replicaInstance, err := r.GetRedisReplicationInstance(ctx, instance)
	if err != nil {
		return err
	}

	replicaInfo, err := k8sredis.GetReplicaInfo(ctx, r.K8Client, replicaInstance, reqLogger)
	if err != nil {
		return err
	}

	masterDNS := "none"
	for _, info := range replicaInfo {
		if role, ok := info.Info["role"]; ok {
			if role == "master" {
				masterDNS = info.DNS
				break
			}
		}
	}

	if len(masterDNS) > 63 { // labels must be less than 64 characters
		masterDNS = masterDNS[:63]
	}

	for _, pod := range podList.Items {

		if master, ok := pod.Labels["redis.operator/redis-master"]; ok {
			if master == masterDNS {
				continue
			}
		}
		pod.Labels["redis.operator/redis-master"] = masterDNS
		_, err := r.K8Client.CoreV1().Pods(instance.Namespace).Update(ctx, &pod, metav1.UpdateOptions{})
		if err != nil {
			continue // ignore the error. We don't care about the state of the pod.
		}
	}

	return nil
}

func (r *RedisSentinelReconciler) CheckSentinelStatus(ctx context.Context, instance *v1.RedisSentinel, logger logr.Logger) error {

	replicaInstance, err := r.GetRedisReplicationInstance(ctx, instance)
	if err != nil {
		return err
	}

	redisInfo, err := k8sredis.GetSentinelMasters(ctx, r.K8Client, instance, replicaInstance)
	if err != nil {
		return err
	}

	for _, info := range redisInfo {

		if foundDownTime, ok := info.Info["s-down-time"]; ok {
			downTime, err := strconv.Atoi(foundDownTime)
			if err != nil {
				return err
			}
			if downTime > 20000 { // 20s
				podName := fmt.Sprintf("%s-%d", instance.Name, info.PodIndex)
				logger.Info("Detected a sentinel down longer than 20s. Restarting", "PodIP", info.DNS)
				err = r.K8Client.CoreV1().Pods(instance.Namespace).Delete(ctx, podName, metav1.DeleteOptions{}) // delete the pod to force it to restart with the updated configmap.
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (r *RedisSentinelReconciler) CreateOrUpdateService(ctx context.Context, instance *v1.RedisSentinel, logger logr.Logger) error {

	newService := redissentinel.CreateSentinelService(instance)

	_, err := r.K8Client.CoreV1().Services(instance.Namespace).Get(ctx, instance.GetServiceName(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Creating service")
			_, err := r.K8Client.CoreV1().Services(instance.Namespace).Create(ctx, &newService, metav1.CreateOptions{})
			return err
		}
		return err
	}
	_, err = r.K8Client.CoreV1().Services(instance.Namespace).Update(ctx, &newService, metav1.UpdateOptions{})
	return err
}

func (r *RedisSentinelReconciler) CreateOrUpdateHeadlessService(ctx context.Context, instance *v1.RedisSentinel, logger logr.Logger) error {

	newService := redissentinel.CreateSentinelHeadlessService(instance)

	_, err := r.K8Client.CoreV1().Services(instance.Namespace).Get(ctx, instance.GetHeadlessServiceName(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Creating headless service")
			_, err := r.K8Client.CoreV1().Services(instance.Namespace).Create(ctx, &newService, metav1.CreateOptions{})
			return err
		}
		return err
	}
	_, err = r.K8Client.CoreV1().Services(instance.Namespace).Update(ctx, &newService, metav1.UpdateOptions{})
	return err
}

func (r *RedisSentinelReconciler) CreateOrUpdateSentinel(ctx context.Context, instance *v1.RedisSentinel, logger logr.Logger) error {

	replicationInstance, err := r.GetRedisReplicationInstance(ctx, instance)
	if err != nil {
		return fmt.Errorf("failed to get replication instance: %v", err)
	}

	initContainer := redissentinel.CreateInitContainer(instance)

	redisContainer, err := redissentinel.CreateContainer(instance, replicationInstance)
	if err != nil {
		return err
	}

	statefulSet := redissentinel.CreateStatefulSet(instance, replicationInstance, redisContainer, initContainer)

	typeMeta := metav1.TypeMeta{Kind: "StatefulSet", APIVersion: "apps/v1"}
	_, err = r.K8Client.AppsV1().StatefulSets(instance.Namespace).Get(ctx, instance.Name, metav1.GetOptions{TypeMeta: typeMeta})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Creating statefulset")
			_, err = r.K8Client.AppsV1().StatefulSets(instance.GetNamespace()).Create(ctx, statefulSet, metav1.CreateOptions{})
			return err
		}
		return err
	}
	_, err = r.K8Client.AppsV1().StatefulSets(instance.Namespace).Update(ctx, statefulSet, metav1.UpdateOptions{})
	return err
}

func (r *RedisSentinelReconciler) GetRedisReplicationInstance(ctx context.Context, instance *v1.RedisSentinel) (*v1.RedisReplication, error) {
	customObject, err := r.Dk8Client.Resource(schema.GroupVersionResource{
		Group:    "redis.redis.operator",
		Version:  "v1",
		Resource: "redisreplications",
	}).Namespace(instance.Namespace).Get(ctx, instance.Spec.RedisReplicationName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	redisReplication := &v1.RedisReplication{}
	replicationJSON, err := customObject.MarshalJSON()
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(replicationJSON, redisReplication); err != nil {
		return nil, err
	}

	return redisReplication, nil
}

func (r *RedisSentinelReconciler) CreateOrUpdateConfigMap(ctx context.Context, instance *v1.RedisSentinel, logger logr.Logger) (error, bool) {

	if _, ok := instance.Spec.RedisConfig.Data["sentinel.conf"]; !ok {
		return fmt.Errorf("sentinel.conf not found in redisConfig"), true
	}

	replicaInstance, err := r.GetRedisReplicationInstance(ctx, instance)
	if err != nil {
		return err, true
	}

	newConfigMap := configmap.NewBuilder().
		SetName(instance.GetConfigName()).
		SetNamespace(instance.Namespace).
		SetData(instance.Spec.RedisConfig.Data).
		BuildWithOwner(instance.GetOwnerReference())

	if err, critical := redissentinel.UpdateConfigMap(ctx, instance, replicaInstance, r.K8Client, newConfigMap, logger); err != nil {
		return err, critical
	}

	_, err = r.K8Client.CoreV1().ConfigMaps(instance.Namespace).Get(ctx, instance.GetConfigName(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Creating configmap")
			_, err = r.K8Client.CoreV1().ConfigMaps(instance.Namespace).Create(ctx, newConfigMap, metav1.CreateOptions{})
			return err, true
		}
		return err, true
	}

	_, err = r.K8Client.CoreV1().ConfigMaps(instance.Namespace).Update(ctx, newConfigMap, metav1.UpdateOptions{})
	return err, true
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisSentinelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.RedisSentinel{}).
		Complete(r)
}
