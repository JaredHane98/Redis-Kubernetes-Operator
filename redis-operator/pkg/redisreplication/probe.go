package redisreplication

import (
	corev1 "k8s.io/api/core/v1"
	v1 "redis.operator/api/v1"
	"redis.operator/pkg/kube/probe"
	"redis.operator/pkg/util/scripts"
)

func GetLivenessScript(instance *v1.RedisReplication) ([]string, error) {
	if instance.Spec.TLSConfig != nil {

		tlsParams, err := instance.Spec.RedisConfig.GetConfigMapTLS()
		if err != nil {
			return nil, err
		}
		return scripts.GetPingScriptAuth(instance.GetRedisPort(), tlsParams.Cert, tlsParams.Key, tlsParams.CACert, tlsParams.Password), nil
	} else {
		password, err := instance.Spec.RedisConfig.GetValue("redis.conf", "requirepass")
		if err != nil {
			return nil, err
		}
		return scripts.GetPingScript(instance.GetRedisPort(), password), nil
	}
}

func GetReadinessScript(instance *v1.RedisReplication) ([]string, error) {
	if instance.Spec.TLSConfig != nil {
		tlsParams, err := instance.Spec.RedisConfig.GetConfigMapTLS()
		if err != nil {
			return nil, err
		}
		return scripts.GetReplicaReadinessScriptAuth(instance.GetRedisPort(), tlsParams.Cert, tlsParams.Key, tlsParams.CACert, tlsParams.Password), nil
	} else {
		password, err := instance.Spec.RedisConfig.GetValue("redis.conf", "requirepass")
		if err != nil {
			return nil, err
		}
		return scripts.GetReplicaReadinessScript(instance.GetRedisPort(), password), nil
	}
}

func GetLivenessProbe(instance *v1.RedisReplication) (*corev1.Probe, error) {
	script, err := GetLivenessScript(instance)
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

func GetReadinessProbe(instance *v1.RedisReplication) (*corev1.Probe, error) {

	script, err := GetReadinessScript(instance)

	if err != nil {
		return nil, err
	}

	return probe.NewBuilder().
		SetInitialDelaySeconds(30). // need a lengthy delay
		SetTimeoutSeconds(1).
		SetPeriodSeconds(5).
		SetSuccessThreshold(1).
		SetFailureThreshold(3).
		SetExecAction(script).
		Build(), nil
}
