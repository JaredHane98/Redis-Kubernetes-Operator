package redissentinel

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	v1 "redis.operator/api/v1"
	"redis.operator/pkg/kube/container"
)

func GetSecurityContext() *corev1.SecurityContext {
	return &corev1.SecurityContext{
		RunAsUser:    ptr.To(int64(2000)),
		RunAsGroup:   ptr.To(int64(2000)),
		RunAsNonRoot: ptr.To(true),
	}
}

func CreateInitContainer(instance *v1.RedisSentinel) corev1.Container {

	initContainer := container.NewBuilder().
		SetName(instance.Name + "-init").
		SetImage("busybox").
		SetCommand([]string{"/bin/sh", "-c", "mkdir -p /tmp/redis && cp /tmp/sentinel.conf /tmp/redis/"}).
		SetImagePullPolicy(corev1.PullIfNotPresent).
		SetSecurityContext(GetSecurityContext()).
		SetVolumeMounts([]corev1.VolumeMount{
			{
				Name:      instance.GetConfigName(),
				MountPath: "/tmp",
			},
			{
				Name:      "redis-data",
				MountPath: "/tmp/redis",
			},
		})

	return initContainer.Build()
}

func CreateContainer(instance *v1.RedisSentinel, replicaInstance *v1.RedisReplication) ([]corev1.Container, error) {

	readinessProbe, err := GetReadinessProbe(instance, replicaInstance)
	if err != nil {
		return nil, err
	}
	livenessProbe, err := GetLivenessProbe(instance, replicaInstance)
	if err != nil {
		return nil, err
	}

	sentinelContainer := container.NewBuilder().
		SetName(instance.Name).
		SetImage(container.GetRedisSentinelImage()).
		SetImagePullPolicy(corev1.PullIfNotPresent).
		SetReadinessProbe(readinessProbe).
		SetLivenessProbe(livenessProbe).
		SetResourceRequirements(instance.Spec.Resources).
		SetSecurityContext(GetSecurityContext()).
		SetVolumeMount(corev1.VolumeMount{
			Name:      "redis-data",
			MountPath: "/tmp/redis",
		}).
		SetArgs([]string{"/tmp/redis/sentinel.conf", "--sentinel"})

	if replicaInstance.Spec.TLSConfig != nil {
		for _, volume := range replicaInstance.Spec.VolumeMounts {
			if volume.Name == replicaInstance.Spec.TLSConfig.Name {
				sentinelContainer.SetVolumeMount(volume)
			}
		}
	}

	return []corev1.Container{sentinelContainer.Build()}, nil
}
