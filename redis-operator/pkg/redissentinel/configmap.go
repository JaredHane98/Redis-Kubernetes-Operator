package redissentinel

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	v1 "redis.operator/api/v1"
	"redis.operator/pkg/kube/configmap"
	k8sredis "redis.operator/pkg/redis"
)

func UpdateConfigMap(ctx context.Context, sentinelInstance *v1.RedisSentinel, replicaInstance *v1.RedisReplication, k8Client kubernetes.Interface, configMap *corev1.ConfigMap, reqLogger logr.Logger) (error, bool) {

	replicaPort := replicaInstance.GetRedisPort()

	redisInfo, err := k8sredis.GetReplicaInfo(ctx, k8Client, replicaInstance, reqLogger)
	if err != nil {
		return err, true
	}

	masterDNS := []string{}
	for _, info := range redisInfo {
		if role, ok := info.Info["role"]; ok {
			if role == "master" {
				masterDNS = append(masterDNS, info.DNS)
			}
		}
	}

	if len(masterDNS) != 1 {
		return fmt.Errorf("failed to update configmap. uncertain master IP. retry later"), false
	}

	if !configmap.UpdateConfigMapKey(configMap, "sentinel.conf", "sentinel monitor", fmt.Sprintf("sentinel monitor %s %s %s %d", sentinelInstance.Spec.MasterName, masterDNS[0], replicaPort, sentinelInstance.Spec.RedisSentinelQuorum)) {
		return fmt.Errorf("failed to update configmap"), true
	}
	if !configmap.UpdateConfigMapKey(configMap, "sentinel.conf", "SENTINEL resolve-hostnames", "SENTINEL resolve-hostnames yes") {
		return fmt.Errorf("failed to update configmap"), true
	}
	if !configmap.UpdateConfigMapKey(configMap, "sentinel.conf", "SENTINEL announce-hostnames", "SENTINEL announce-hostnames yes") {
		return fmt.Errorf("failed to update configmap"), true
	}

	return nil, true
}
