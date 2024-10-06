package statefulset

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//"redis.operator/pkg/util/basic"
	"k8s.io/utils/ptr"
)

type Builder struct {
	Name            string
	Namespace       string
	ServiceName     string
	Labels          map[string]string
	MatchLabels     map[string]string
	OwnerReference  []metav1.OwnerReference
	Containers      []corev1.Container
	InitContainers  []corev1.Container
	Volumes         []corev1.Volume
	StatefulSpec    *appsv1.StatefulSetSpec
	SecurityContext *corev1.PodSecurityContext
}

func (s *Builder) SetInitContainer(container corev1.Container) *Builder {
	s.InitContainers = append(s.InitContainers, container)
	return s
}

func (s *Builder) SetInitContainers(containers []corev1.Container) *Builder {
	s.InitContainers = append(s.InitContainers, containers...)
	return s
}

func (s *Builder) SetVolumes(volumes []corev1.Volume) *Builder {
	s.Volumes = append(s.Volumes, volumes...)
	return s
}

func (s *Builder) SetVolume(volume corev1.Volume) *Builder {
	s.Volumes = append(s.Volumes, volume)
	return s
}

func (s *Builder) SetName(name string) *Builder {
	s.Name = name
	return s
}

func (s *Builder) SetNamespace(namespace string) *Builder {
	s.Namespace = namespace
	return s
}

func (s *Builder) SetServiceName(serviceName string) *Builder {
	s.ServiceName = serviceName
	return s
}

func (s *Builder) SetLabels(labels map[string]string) *Builder {
	s.Labels = labels
	return s
}

func (s *Builder) SetMatchLabels(matchLabels map[string]string) *Builder {
	s.MatchLabels = matchLabels
	return s
}

func (s *Builder) SetOwnerReference(ownerReference []metav1.OwnerReference) *Builder {
	s.OwnerReference = ownerReference
	return s
}

// The spec should be validated before being passed here. Additionally the container and labels will be probably be overwritten
func (s *Builder) SetStatefulSpec(statefulSpec *appsv1.StatefulSetSpec) *Builder {
	s.StatefulSpec = statefulSpec
	return s
}

func (s *Builder) SetContainers(containers []corev1.Container) *Builder {
	s.Containers = append(s.Containers, containers...)
	return s
}

func (s *Builder) SetContainer(container corev1.Container) *Builder {
	s.Containers = append(s.Containers, container)
	return s
}

func (s *Builder) Build() (*appsv1.StatefulSet, error) {

	statefulSpec, err := s.GetStatefulSpec()
	if err != nil {
		return nil, err
	}
	if len(s.Containers) == 0 {
		return nil, fmt.Errorf("containers are required to build a statefulset")
	}

	statefulSpec.Template.Spec.Containers = append(statefulSpec.Template.Spec.Containers, s.Containers...)
	statefulSpec.Template.Spec.Volumes = append(statefulSpec.Template.Spec.Volumes, s.Volumes...)
	statefulSpec.Template.Spec.InitContainers = append(statefulSpec.Template.Spec.InitContainers, s.InitContainers...)

	statefulSet := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: s.Namespace,
			Labels:    s.Labels,
		},
		Spec: *statefulSpec,
	}

	return statefulSet, nil
}

func (s *Builder) BuildWithOwner(owner metav1.OwnerReference) (*appsv1.StatefulSet, error) {
	statefulset, err := s.Build()
	if err != nil {
		return statefulset, err
	}
	statefulset.SetOwnerReferences(append(statefulset.GetOwnerReferences(), owner))
	return statefulset, nil
}

func NewBuilder() *Builder {
	return &Builder{}
}

// need to fix
func (s *Builder) CreateDefaultStatefulSpec() (*appsv1.StatefulSetSpec, error) {

	if len(s.Labels) == 0 {
		s.Labels = map[string]string{
			"app": s.ServiceName,
		}
	}

	spec := appsv1.StatefulSetSpec{
		Replicas: ptr.To(int32(1)),
		Selector: &metav1.LabelSelector{
			MatchLabels: s.MatchLabels,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: s.MatchLabels,
			},
			Spec: corev1.PodSpec{
				Containers: s.Containers,
			},
		},
	}

	return &spec, nil
}

func (s *Builder) GetStatefulSpec() (*appsv1.StatefulSetSpec, error) {
	if s.StatefulSpec == nil {
		return nil, fmt.Errorf("statefulspec is required to build a statefulset")
	}
	return s.StatefulSpec.DeepCopy(), nil
}

func (s *Builder) GetContainers() ([]corev1.Container, error) {
	if len(s.Containers) == 0 {
		if len(s.StatefulSpec.Template.Spec.Containers) == 0 {
			return nil, fmt.Errorf("statefulset was not associated with a container")
		}
		return s.StatefulSpec.Template.Spec.Containers, nil
	}
	return s.Containers, nil
}

func GetResourceRequirements(resources *corev1.ResourceRequirements) *corev1.ResourceRequirements {
	if resources == nil {
		return &corev1.ResourceRequirements{
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
	return resources
}
