package redisreplication

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

func CreateContainer(instance *v1.RedisReplication) (corev1.Container, error) {

	// The sleep timer is equal to down-after-milliseconds + 1 second to prevent the Sentinel from re-recognizing a down master
	seconds := 0
	if instance.Spec.RedisSentinelConfig != nil {
		if instance.Spec.RedisSentinelConfig.RedisSentinelDowntime == nil {
			seconds = 30 // default
		} else {
			seconds = (*instance.Spec.RedisSentinelConfig.RedisSentinelDowntime / 1000) + 1
		}
	}

	args := fmt.Sprintf(
		`
	mkdir -p /tmp/redis
	cp tmp/redis.conf /tmp/redis/
	replica_announce_ip="${POD_NAME}.%s.%s.svc.cluster.local"
	echo "replica-announce-ip ${replica_announce_ip}" >> /tmp/redis/redis.conf
	sleep %d
	`, instance.GetHeadlessServiceName(), instance.Namespace, seconds)

	initContainer := container.NewBuilder().
		SetName(instance.Name + "-init").
		SetImage("busybox").
		SetCommand([]string{"/bin/sh", "-c"}).
		SetEnvs([]corev1.EnvVar{
			{
				Name: "POD_NAME",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.name",
					},
				},
			},
		}).
		SetArgs([]string{args}).
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

	return initContainer.Build(), nil
}

func CreateContainers(instance *v1.RedisReplication) ([]corev1.Container, error) {

	livenessProbe, err := GetLivenessProbe(instance)
	if err != nil {
		return nil, err
	}

	readinessProbe, err := GetReadinessProbe(instance)
	if err != nil {
		return nil, err
	}

	containers := []corev1.Container{}

	redisContainer := container.NewBuilder().
		SetName(instance.Name).
		SetImage(container.GetRedisReplicationImage()).
		SetImagePullPolicy(corev1.PullIfNotPresent).
		SetResourceRequirements(instance.Spec.Resources).
		SetLivenessProbe(livenessProbe).
		SetReadinessProbe(readinessProbe).
		SetSecurityContext(GetSecurityContext()).
		SetVolumeMounts(instance.Spec.VolumeMounts). // set user volume mounts
		SetVolumeMount(corev1.VolumeMount{           // set config volume mount

			Name:      "redis-data",
			MountPath: "/tmp/redis",
		}).
		SetArgs([]string{"/tmp/redis/redis.conf"})

	containers = append(containers, redisContainer.Build())

	if instance.Spec.EnableExporter {
		exportContainer := container.NewBuilder().
			SetName(instance.Name + "-exporter").
			SetImage(container.GetRedisExporterImage()).
			SetImagePullPolicy(corev1.PullIfNotPresent).
			SetVolumeMounts(instance.Spec.VolumeMounts).
			SetSecurityContext(GetSecurityContext()).
			SetVolumeMount(corev1.VolumeMount{
				Name:      "redis-data",
				MountPath: "/tmp/redis",
			}).
			SetResourceRequirements(&corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("100Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("100Mi"),
				},
			}).
			SetPorts([]corev1.ContainerPort{
				{
					Name:          "redis-exporter",
					ContainerPort: 9121,
					Protocol:      corev1.ProtocolTCP,
				},
			})

		password, err := instance.Spec.RedisConfig.GetValue("redis.conf", "requirepass")
		if err != nil {
			return nil, err
		}

		exportContainer.SetEnvs([]corev1.EnvVar{
			{
				Name:  "REDIS_EXPORTER_INCL_SYSTEM_METRICS",
				Value: "true",
			},
			{
				Name:  "REDIS_PASSWORD",
				Value: password,
			},
		})

		if instance.Spec.TLSConfig != nil {

			tlsConfig, err := instance.Spec.RedisConfig.GetConfigMapTLS()
			if err != nil {
				return nil, err
			}
			exportContainer.SetEnvs([]corev1.EnvVar{
				{
					Name:  "REDIS_EXPORTER_TLS_CLIENT_KEY_FILE",
					Value: tlsConfig.Key,
				},
				{
					Name:  "REDIS_EXPORTER_TLS_CA_CERT_FILE",
					Value: tlsConfig.CACert,
				},
				{
					Name:  "REDIS_EXPORTER_TLS_CLIENT_CERT_FILE",
					Value: tlsConfig.Cert,
				},
				{
					Name:  "REDIS_ADDR",
					Value: "rediss://localhost:" + instance.GetRedisPort(),
				},
				{
					Name:  "REDIS_EXPORTER_SKIP_TLS_VERIFICATION",
					Value: "true",
				},
				{
					Name:  "REDIS_EXPORTER_DEBUG",
					Value: "true",
				},
			})
		} else {
			exportContainer.SetEnvs([]corev1.EnvVar{
				{
					Name:  "REDIS_ADDR",
					Value: "redis://localhost:" + instance.GetRedisPort(),
				},
			})
		}
		containers = append(containers, exportContainer.Build())
	}

	return containers, nil
}
