package redisreplication

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	v1 "redis.operator/api/v1"
	"redis.operator/pkg/kube/service"
)

func GetReplicationServiceLabels(instance *v1.RedisReplication) map[string]string {

	return map[string]string{
		"app.kubernetes.io/name":       instance.Name + "-service",
		"app.kubernetes.io/instance":   "redis",
		"app.kubernetes.io/version":    "1.0",
		"app.kubernetes.io/component":  "redis-database",
		"app.kubernetes.io/part-of":    "redisreplication",
		"app.kubernetes.io/managed-by": "redis-operator",
	}
}

func CreateHeadlessReplicationService(instance *v1.RedisReplication) corev1.Service {

	port := instance.GetRedisPortInt32()
	labels := GetReplicationServiceLabels(instance)

	serviceBuilder := service.NewBuilder().
		SetName(instance.GetHeadlessServiceName()).
		SetNamespace(instance.Namespace).
		SetSelector(labels).
		SetLabels(labels).
		SetServiceType(corev1.ServiceTypeClusterIP).
		SetClusterIP("None").
		SetPublishNotReadyAddresses(true).
		SetOwnerReference(instance.GetOwnerReference()).
		SetPort(corev1.ServicePort{
			Name:       "redis-client",
			Port:       port,
			TargetPort: intstr.FromInt32(port),
			Protocol:   corev1.ProtocolTCP,
		})

	if instance.Spec.EnableExporter {
		serviceBuilder.SetPort(corev1.ServicePort{
			Name:       "redis-exporter",
			Port:       9121,
			TargetPort: intstr.FromInt(9121),
			Protocol:   corev1.ProtocolTCP,
		})
	}
	return serviceBuilder.Build()
}

func CreateReplicationService(instance *v1.RedisReplication) corev1.Service {
	port := instance.GetRedisPortInt32()
	labels := GetReplicationServiceLabels(instance)

	serviceBuilder := service.NewBuilder().
		SetName(instance.GetServiceName()).
		SetNamespace(instance.Namespace).
		SetSelector(labels).
		SetLabels(labels).
		SetServiceType(corev1.ServiceTypeNodePort).
		SetOwnerReference(instance.GetOwnerReference()).
		SetPort(corev1.ServicePort{
			Name:       "redis-client",
			Port:       port,
			TargetPort: intstr.FromInt32(port),
			Protocol:   corev1.ProtocolTCP,
		})
	return serviceBuilder.Build()
}
