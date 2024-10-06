package redisreplication

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "redis.operator/api/v1"
)

func CreateStatefulSet(instance *v1.RedisReplication, redisContainers []corev1.Container, initContainer corev1.Container) *appsv1.StatefulSet {

	volumes := []corev1.Volume{
		{
			Name: instance.GetConfigName(),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: instance.GetConfigName(),
					},
				},
			},
		},
		{
			Name: "redis-data",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	if instance.Spec.TLSConfig != nil {
		volumes = append(volumes, corev1.Volume{
			Name: instance.Spec.TLSConfig.Name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: instance.Spec.TLSConfig.SecretName,
				},
			},
		})
	}

	statefulSet := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    GetReplicationServiceLabels(instance),
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: instance.Spec.StatefulsetConfig.Wrapper.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: GetReplicationServiceLabels(instance),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: GetReplicationServiceLabels(instance),
				},
				Spec: corev1.PodSpec{
					Volumes:                       append(volumes, instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.Volumes...),
					InitContainers:                append([]corev1.Container{initContainer}, instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.InitContainers...), // init containers are okay, infact they're great to use with PVCs
					Containers:                    redisContainers,
					EphemeralContainers:           instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.EphemeralContainers,
					RestartPolicy:                 instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.RestartPolicy,
					TerminationGracePeriodSeconds: instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.TerminationGracePeriodSeconds,
					ActiveDeadlineSeconds:         instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.ActiveDeadlineSeconds,
					DNSPolicy:                     instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.DNSPolicy, // might break my code
					NodeSelector:                  instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.NodeSelector,
					ServiceAccountName:            instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.ServiceAccountName,
					DeprecatedServiceAccount:      instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.DeprecatedServiceAccount,
					AutomountServiceAccountToken:  instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.AutomountServiceAccountToken,
					NodeName:                      instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.NodeName,
					HostNetwork:                   instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.HostNetwork,
					HostPID:                       instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.HostPID,
					HostIPC:                       instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.HostIPC,
					ShareProcessNamespace:         instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.ShareProcessNamespace,
					SecurityContext:               instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.SecurityContext,
					ImagePullSecrets:              instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.ImagePullSecrets,
					Hostname:                      instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.Hostname,
					Subdomain:                     instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.Subdomain,
					Affinity:                      instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.Affinity,
					SchedulerName:                 instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.SchedulerName,
					Tolerations:                   instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.Tolerations,
					HostAliases:                   instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.HostAliases,
					PriorityClassName:             instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.PriorityClassName,
					Priority:                      instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.Priority,
					ReadinessGates:                instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.ReadinessGates,
					RuntimeClassName:              instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.RuntimeClassName,
					EnableServiceLinks:            instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.EnableServiceLinks,
					PreemptionPolicy:              instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.PreemptionPolicy,
					Overhead:                      instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.Overhead,
					TopologySpreadConstraints:     instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.TopologySpreadConstraints,
					SetHostnameAsFQDN:             instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.SetHostnameAsFQDN,
					OS:                            instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.OS,
					HostUsers:                     instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.HostUsers,
					SchedulingGates:               instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.SchedulingGates,
					ResourceClaims:                instance.Spec.StatefulsetConfig.Wrapper.Spec.Template.Spec.ResourceClaims,
				},
			},
			VolumeClaimTemplates:                 instance.Spec.StatefulsetConfig.Wrapper.Spec.VolumeClaimTemplates,
			ServiceName:                          instance.GetHeadlessServiceName(),
			PodManagementPolicy:                  appsv1.ParallelPodManagement,
			UpdateStrategy:                       instance.Spec.StatefulsetConfig.Wrapper.Spec.UpdateStrategy,
			RevisionHistoryLimit:                 instance.Spec.StatefulsetConfig.Wrapper.Spec.RevisionHistoryLimit,
			MinReadySeconds:                      instance.Spec.StatefulsetConfig.Wrapper.Spec.MinReadySeconds,
			PersistentVolumeClaimRetentionPolicy: instance.Spec.StatefulsetConfig.Wrapper.Spec.PersistentVolumeClaimRetentionPolicy,
			Ordinals:                             nil, // not going to allow oridinals
		},
	}

	statefulSet.SetOwnerReferences(append(statefulSet.GetOwnerReferences(), instance.GetOwnerReference()))

	return statefulSet
}
