package redissentinel

import (
	corev1 "k8s.io/api/core/v1"
	v1 "redis.operator/api/v1"
	"redis.operator/pkg/kube/configmap"
	"redis.operator/pkg/kube/probe"
	"redis.operator/pkg/util/scripts"
)

func GetLivenessScript(instance *v1.RedisSentinel, replicaInstance *v1.RedisReplication) ([]string, error) {
	if replicaInstance.Spec.TLSConfig != nil {

		tlsParams, err := replicaInstance.Spec.RedisConfig.GetConfigMapTLS()
		if err != nil {
			return nil, err
		}

		password := configmap.GetConfigMapValue(instance.Spec.RedisConfig.Data, "sentinel.conf", "requirepass")

		return scripts.GetPingScriptAuth(instance.GetRedisPort(), tlsParams.Cert, tlsParams.Key, tlsParams.CACert, password), nil

	} else {
		redisPassword := configmap.GetConfigMapValue(instance.Spec.RedisConfig.Data, "sentinel.conf", "requirepass")

		return scripts.GetPingScript(instance.GetRedisPort(), redisPassword), nil
	}
}

func GetReadinessScript(instance *v1.RedisSentinel, replicaInstance *v1.RedisReplication) ([]string, error) {
	if replicaInstance.Spec.TLSConfig != nil {

		tlsParams, err := replicaInstance.Spec.RedisConfig.GetConfigMapTLS()
		if err != nil {
			return nil, err
		}
		redisPassword := configmap.GetConfigMapValue(instance.Spec.RedisConfig.Data, "sentinel.conf", "requirepass")

		return scripts.GetDownTimeScriptAuth(replicaInstance.GetRedisPort(), tlsParams.Cert, tlsParams.Key, tlsParams.CACert, redisPassword), nil

	} else {
		redisPassword := configmap.GetConfigMapValue(instance.Spec.RedisConfig.Data, "sentinel.conf", "requirepass")

		return scripts.GetDownTimeScript(replicaInstance.GetRedisPort(), redisPassword), nil
	}
}

func GetLivenessProbe(instance *v1.RedisSentinel, replicaInstance *v1.RedisReplication) (*corev1.Probe, error) {
	script, err := GetLivenessScript(instance, replicaInstance)
	if err != nil {
		return nil, err
	}

	return probe.NewBuilder().
		SetInitialDelaySeconds(10).
		SetTimeoutSeconds(1).
		SetPeriodSeconds(10).
		SetSuccessThreshold(1).
		SetFailureThreshold(3).
		SetExecAction(script).
		Build(), nil
}

func GetReadinessProbe(instance *v1.RedisSentinel, replicaInstance *v1.RedisReplication) (*corev1.Probe, error) {
	script, err := GetReadinessScript(instance, replicaInstance)
	if err != nil {
		return nil, err
	}

	return probe.NewBuilder().
		SetInitialDelaySeconds(10).
		SetTimeoutSeconds(1).
		SetPeriodSeconds(10).
		SetSuccessThreshold(1).
		SetFailureThreshold(3).
		SetExecAction(script).
		Build(), nil
}
