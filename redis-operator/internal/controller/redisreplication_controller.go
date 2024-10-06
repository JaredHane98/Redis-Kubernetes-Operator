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
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	v1 "redis.operator/api/v1"
	"redis.operator/pkg/kube/configmap"
	k8sredis "redis.operator/pkg/redis"
	"redis.operator/pkg/redisreplication"
	"redis.operator/pkg/util/result"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// RedisReplicationReconciler reconciles a RedisReplication object
type RedisReplicationReconciler struct {
	client.Client
	K8Client  kubernetes.Interface
	Dk8Client dynamic.Interface
	Scheme    *runtime.Scheme
	Log       logr.Logger
}

func (r *RedisReplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	instance := &v1.RedisReplication{}

	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return result.ReconciledWithMessage(reqLogger, "Failed to get instance. Assumming it was deleted")
		}
		return result.FailedWithError(err, reqLogger, "Error reconciling instance")
	}

	if instance.ObjectMeta.GetDeletionTimestamp() != nil {
		if err = r.HandleReplicationFinalizer(ctx, instance, v1.RedisReplicationFinalizer); err != nil {
			return result.RetryWithError(err, reqLogger, "Failed to handle finalizer")
		}
		return result.Ok()
	}

	if err = r.CreateReplicationFinalizer(ctx, instance, r.Client, v1.RedisReplicationFinalizer); err != nil {
		return result.RetryWithError(err, reqLogger, "Failed to create finalizer")
	}

	if err = r.CreateOrUpdateHeadlessService(ctx, instance, reqLogger); err != nil {
		return result.RetryWithError(err, reqLogger, "Failed to create service for redis instance")
	}

	if err = r.CreateOrUpdateService(ctx, instance, reqLogger); err != nil {
		return result.RetryWithError(err, reqLogger, "Failed to create service for redis instance")
	}

	if err = r.CreateOrUpdateConfigMap(ctx, instance, reqLogger); err != nil {
		return result.RetryWithError(err, reqLogger, "Failed to create configmap for redis instance")
	}

	if err = r.CreateOrUpdateStateful(ctx, instance, reqLogger); err != nil {
		return result.RetryWithError(err, reqLogger, "Failed to create statefulset for redis instance")
	}

	if err = r.UpdateRedisMaster(ctx, instance, reqLogger); err != nil {
		return result.RetryWithError(err, reqLogger, "Failed to update redis master")
	}

	return result.RequeueAfter(1 * time.Second)
}

func (r *RedisReplicationReconciler) CreateReplicationFinalizer(ctx context.Context, instance *v1.RedisReplication, client client.Client, finalizer string) error {
	if !controllerutil.ContainsFinalizer(instance, finalizer) {
		controllerutil.AddFinalizer(instance, finalizer)
		return client.Update(ctx, instance)
	}
	return nil
}

func (r *RedisReplicationReconciler) HandleReplicationFinalizer(ctx context.Context, instance *v1.RedisReplication, finalizer string) error {
	if controllerutil.ContainsFinalizer(instance, finalizer) {
		controllerutil.RemoveFinalizer(instance, finalizer)
		return r.Client.Update(ctx, instance)
	}
	return nil
}

func (r *RedisReplicationReconciler) UpdateReplicationLabels(ctx context.Context, instance *v1.RedisReplication, reqLogger logr.Logger) error {

	labels := redisreplication.GetReplicationServiceLabels(instance)
	selector := make([]string, 0, len(labels))
	for key, value := range labels {
		selector = append(selector, fmt.Sprintf("%s=%s", key, value))
	}

	podList, err := r.K8Client.CoreV1().Pods(instance.Namespace).List(ctx, metav1.ListOptions{LabelSelector: strings.Join(selector, ",")})
	if err != nil {
		return err
	}

	replicaInfo, err := k8sredis.GetReplicaInfo(ctx, r.K8Client, instance, reqLogger)
	if err != nil {
		return err
	}

	for _, pod := range podList.Items {
		indexStr, ok := pod.Labels["apps.kubernetes.io/pod-index"]
		if !ok {
			continue
		}

		index, err := strconv.Atoi(indexStr)
		if err != nil {
			return err
		}

		for _, info := range replicaInfo {
			if info.PodIndex == index {
				if role, ok := info.Info["role"]; ok {
					if currentLabel, ok := pod.Labels["redis.operator/redis-role"]; ok {
						if currentLabel == role {
							break
						}
					}
					pod.Labels["redis.operator/redis-role"] = role
					_, err := r.K8Client.CoreV1().Pods(instance.Namespace).Update(ctx, &pod, metav1.UpdateOptions{})
					if err != nil {
						return err
					}
					break
				}
			}
		}
	}

	return nil
}

func (r *RedisReplicationReconciler) CreateOrUpdateConfigMap(ctx context.Context, instance *v1.RedisReplication, reqLogger logr.Logger) error {

	configMap := configmap.NewBuilder().
		SetName(instance.GetConfigName()). // name was: redis-config
		SetNamespace(instance.Namespace).
		SetData(instance.Spec.RedisConfig.Data). // key was: redis.conf
		BuildWithOwner(instance.GetOwnerReference())

	_, err := r.K8Client.CoreV1().ConfigMaps(instance.Namespace).Get(ctx, instance.GetConfigName(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			reqLogger.Info("Creating configmap")
			_, err := r.K8Client.CoreV1().ConfigMaps(instance.Namespace).Create(ctx, configMap, metav1.CreateOptions{})
			return err
		}
		return err
	}

	_, err = r.K8Client.CoreV1().ConfigMaps(instance.Namespace).Update(ctx, configMap, metav1.UpdateOptions{})
	return err
}

func (r *RedisReplicationReconciler) GetRedisSentinelInstance(ctx context.Context, instance *v1.RedisReplication) (*v1.RedisSentinel, error) {

	if instance.Spec.RedisSentinelConfig == nil {
		return nil, fmt.Errorf("redisSentinelName is not set")
	}

	customObject, err := r.Dk8Client.Resource(schema.GroupVersionResource{
		Group:    "redis.redis.operator",
		Version:  "v1",
		Resource: "redissentinels",
	}).Namespace(instance.Namespace).Get(ctx, instance.Spec.RedisSentinelConfig.RedisSentinelName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	redisSentinel := &v1.RedisSentinel{}
	sentinelJSON, err := customObject.MarshalJSON()
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(sentinelJSON, redisSentinel); err != nil {
		return nil, err
	}

	return redisSentinel, nil
}

// go through the sentinels masters and find the 'agreed master' using the supplied quorum. We don't use any
// sentinels considered down
func GetSentinelMasterCandidate(sentinelMasters []k8sredis.RedisCommandInfo, sentinelInstance *v1.RedisSentinel) (string, error) {

	candidates := make([]struct {
		DNS    string
		Agreed int
	}, 0)

	//candidates := []RedisCandidates{}
	for _, sentinelMaster := range sentinelMasters {

		if _, ok := sentinelMaster.Info["s-down-time"]; ok { // don't use down sentinels
			continue
		}

		if ip, ok := sentinelMaster.Info["ip"]; ok {

			found := false
			for i, candidate := range candidates {
				if candidate.DNS == ip {
					candidates[i].Agreed++
					found = true
					break
				}
			}

			if !found {
				candidates = append(candidates, struct {
					DNS    string
					Agreed int
				}{
					DNS:    ip,
					Agreed: 1,
				})
			}
		}
	}

	for _, candidate := range candidates {
		if candidate.Agreed >= sentinelInstance.Spec.RedisSentinelQuorum {
			return candidate.DNS, nil
		}
	}

	return "", nil
}

// returns the master with the highest number of slaves.
func GetHighestConnectedSlaves(replicationInfo []k8sredis.RedisCommandInfo) (string, error) {
	realMaster := ""
	slaveHigh := 0
	for _, info := range replicationInfo {
		if slaves, ok := info.Info["connected_slaves"]; ok {
			slaveCount, err := strconv.Atoi(slaves)
			if err != nil {
				return "", err
			}
			if slaveHigh < slaveCount {
				slaveHigh = slaveCount
				realMaster = info.DNS
			} else if slaveHigh == slaveCount {
				realMaster = "" // don't demote unless certain.
			}
		}
	}
	return realMaster, nil
}

func (r *RedisReplicationReconciler) UpdateRedisMaster(ctx context.Context, instance *v1.RedisReplication, reqLogger logr.Logger) error {

	replicationInfo, err := k8sredis.GetReplicaInfo(ctx, r.K8Client, instance, reqLogger)
	if err != nil {
		return err
	}

	if len(replicationInfo) == 0 {
		return nil
	}

	masters := 0
	slaves := 0
	for _, info := range replicationInfo {
		if role, ok := info.Info["role"]; ok {
			if role == "master" {
				masters++
			} else if role == "slave" {
				slaves++
			}
		}
	}

	if masters == 1 {
		return nil
	}

	if instance.Spec.RedisSentinelConfig == nil {
		if slaves == 0 {
			realMaster := replicationInfo[0].DNS
			reqLogger.Info("running without a sentinel. promoting first instance", "master", realMaster)
			return k8sredis.SetReplicationMaster(ctx, r.K8Client, instance, realMaster, reqLogger)
		}
		realMaster, err := GetHighestConnectedSlaves(replicationInfo)
		if err != nil {
			return err
		}
		if realMaster == "" {
			reqLogger.Info("uncertainity about current master. cannot promote an instance")
			return nil // cannot promote unless certain
		}
		reqLogger.Info("updating redis master instances", "master", realMaster)
		return k8sredis.SetReplicationMaster(ctx, r.K8Client, instance, realMaster, reqLogger)
	}

	sentinelInstance, err := r.GetRedisSentinelInstance(ctx, instance)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		if slaves == 0 {
			realMaster := replicationInfo[0].DNS
			reqLogger.Info("no sentinel instance found. promoting first instance", "master", realMaster)
			return k8sredis.SetReplicationMaster(ctx, r.K8Client, instance, realMaster, reqLogger)
		}
		realMaster, err := GetHighestConnectedSlaves(replicationInfo)
		if err != nil {
			return err
		}
		if realMaster == "" {
			reqLogger.Info("uncertainity about current master", "master", realMaster)
			return nil // wait for the sentinel. Do not promote unless certain
		}
		reqLogger.Info("updating redis master instances", "master", realMaster)
		return k8sredis.SetReplicationMaster(ctx, r.K8Client, instance, realMaster, reqLogger)
	}

	sentinelMasters, err := k8sredis.GetSentinelMasters(ctx, r.K8Client, sentinelInstance, instance)
	if err != nil {
		return err
	}

	if len(sentinelMasters) == 0 {
		return fmt.Errorf("failed to query sentinel masters")
	}

	candidate, err := GetSentinelMasterCandidate(sentinelMasters, sentinelInstance)
	if err != nil {
		return err
	}

	if candidate != "" {
		reqLogger.Info("sentinels agreed on a new master. updating instances...", "master", candidate)
		return k8sredis.SetReplicationMaster(ctx, r.K8Client, instance, candidate, reqLogger)
	}

	reqLogger.Info("Sentinels have not agreed on a new master")
	return nil
}

func (r *RedisReplicationReconciler) CreateOrUpdateStateful(ctx context.Context, instance *v1.RedisReplication, reqLogger logr.Logger) error {

	initContainer, err := redisreplication.CreateContainer(instance)
	if err != nil {
		return err
	}

	redisContainers, err := redisreplication.CreateContainers(instance)
	if err != nil {
		return err
	}

	statefulSet := redisreplication.CreateStatefulSet(instance, redisContainers, initContainer)

	typeMeta := metav1.TypeMeta{Kind: "StatefulSet", APIVersion: "apps/v1"}
	_, err = r.K8Client.AppsV1().StatefulSets(instance.GetNamespace()).Get(ctx, instance.GetName(), metav1.GetOptions{TypeMeta: typeMeta})
	if err != nil {
		if apierrors.IsNotFound(err) {
			reqLogger.Info("Creating statefulset")
			_, err = r.K8Client.AppsV1().StatefulSets(instance.GetNamespace()).Create(ctx, statefulSet, metav1.CreateOptions{})
			return err
		}
	}

	_, err = r.K8Client.AppsV1().StatefulSets(instance.GetNamespace()).Update(ctx, statefulSet, metav1.UpdateOptions{})
	return err
}

func (r *RedisReplicationReconciler) CreateOrUpdateHeadlessService(ctx context.Context, instance *v1.RedisReplication, reqLogger logr.Logger) error {

	newService := redisreplication.CreateHeadlessReplicationService(instance)

	_, err := r.K8Client.CoreV1().Services(instance.Namespace).Get(ctx, instance.GetHeadlessServiceName(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			reqLogger.Info("Creating headless service")
			_, err := r.K8Client.CoreV1().Services(instance.GetNamespace()).Create(ctx, &newService, metav1.CreateOptions{})
			return err
		}
		return err
	}

	_, err = r.K8Client.CoreV1().Services(instance.Namespace).Update(ctx, &newService, metav1.UpdateOptions{})
	return err
}

func (r *RedisReplicationReconciler) CreateOrUpdateService(ctx context.Context, instance *v1.RedisReplication, reqLogger logr.Logger) error {

	newService := redisreplication.CreateReplicationService(instance)

	_, err := r.K8Client.CoreV1().Services(instance.Namespace).Get(ctx, instance.GetServiceName(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			reqLogger.Info("Creating service")
			_, err := r.K8Client.CoreV1().Services(instance.GetNamespace()).Create(ctx, &newService, metav1.CreateOptions{})
			return err
		}
		return err
	}
	_, err = r.K8Client.CoreV1().Services(instance.Namespace).Update(ctx, &newService, metav1.UpdateOptions{})
	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisReplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.RedisReplication{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}
