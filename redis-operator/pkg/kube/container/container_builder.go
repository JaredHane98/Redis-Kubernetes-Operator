package container

import (
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func GetRedisReplicationImage() string {

	var envVariable = "REPLICA_IMAGE"
	if image, found := os.LookupEnv(envVariable); found {
		return image
	}
	panic("didn't find replication image")
}

func GetRedisExporterImage() string {
	var envVariable = "EXPORTER_IMAGE"
	if image, found := os.LookupEnv(envVariable); found {
		return image
	}
	panic("didn't find exporter image")
}

func GetRedisSentinelImage() string {

	var envVariable = "SENTINEL_IMAGE"
	if image, found := os.LookupEnv(envVariable); found {
		return image
	}
	panic("didn't find sentinel image")
}

type Builder struct {
	Container corev1.Container
}

func (r *Builder) SetName(name string) *Builder {
	r.Container.Name = name
	return r
}

func (r *Builder) SetImage(image string) *Builder {
	r.Container.Image = image
	return r
}

func (r *Builder) SetCommand(command []string) *Builder {
	r.Container.Command = append(r.Container.Command, command...)
	return r
}

func (r *Builder) SetArgs(args []string) *Builder {
	r.Container.Args = append(r.Container.Args, args...)
	return r
}

func (r *Builder) SetWorkingDir(dir string) *Builder {
	r.Container.WorkingDir = dir
	return r
}

func (r *Builder) SetPorts(ports []corev1.ContainerPort) *Builder {
	r.Container.Ports = append(r.Container.Ports, ports...)
	return r
}

func (r *Builder) SetEnvs(env []corev1.EnvVar) *Builder {
	r.Container.Env = append(r.Container.Env, env...)
	return r
}

func (r *Builder) SetResourceRequirements(resourceReq *corev1.ResourceRequirements) *Builder {
	if resourceReq == nil {
		resourceReq = &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("100Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("100Mi"),
			},
		}
	}

	r.Container.Resources = *resourceReq
	return r
}

func (r *Builder) SetVolumeMounts(volumeMounts []corev1.VolumeMount) *Builder {
	r.Container.VolumeMounts = append(r.Container.VolumeMounts, volumeMounts...)
	return r
}
func (r *Builder) SetVolumeMount(volumeMount corev1.VolumeMount) *Builder {
	r.Container.VolumeMounts = append(r.Container.VolumeMounts, volumeMount)
	return r
}

func (r *Builder) SetResizePolicy(resizePolicy []corev1.ContainerResizePolicy) *Builder {
	r.Container.ResizePolicy = resizePolicy
	return r
}

func (r *Builder) SetRestartPolicy(policy *corev1.ContainerRestartPolicy) *Builder {
	r.Container.RestartPolicy = policy
	return r
}

func (r *Builder) SetVolumeDevices(volumeDevices []corev1.VolumeDevice) *Builder {
	r.Container.VolumeDevices = volumeDevices
	return r
}

func (r *Builder) SetLivenessProbe(probe *corev1.Probe) *Builder {
	r.Container.LivenessProbe = probe
	return r
}

func (r *Builder) SetReadinessProbe(probe *corev1.Probe) *Builder {
	r.Container.ReadinessProbe = probe
	return r
}

func (r *Builder) SetStartupProbe(probe *corev1.Probe) *Builder {
	r.Container.StartupProbe = probe
	return r
}

func (r *Builder) SetLifecycle(lifecycle *corev1.Lifecycle) *Builder {
	r.Container.Lifecycle = lifecycle
	return r
}

func (r *Builder) SetImagePullPolicy(policy corev1.PullPolicy) *Builder {
	r.Container.ImagePullPolicy = policy
	return r
}

func (r *Builder) SetSecurityContext(securityContext *corev1.SecurityContext) *Builder {
	r.Container.SecurityContext = securityContext
	return r
}

func (r *Builder) Build() corev1.Container {
	return r.Container
}

func NewBuilder() *Builder {
	return &Builder{
		Container: corev1.Container{},
	}
}
