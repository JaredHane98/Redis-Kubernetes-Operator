package configmap

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Builder struct {
	data            map[string]string
	name            string
	namespace       string
	ownerReferences []metav1.OwnerReference
	labels          map[string]string
}

func (b *Builder) SetName(name string) *Builder {
	b.name = name
	return b
}

func (b *Builder) SetNamespace(namespace string) *Builder {
	b.namespace = namespace
	return b
}

func (b *Builder) SetDataField(key, value string) *Builder {
	b.data[key] = value
	return b
}

func (b *Builder) SetOwnerReferences(ownerReferences []metav1.OwnerReference) *Builder {
	b.ownerReferences = ownerReferences
	return b
}

func (b *Builder) SetLabels(labels map[string]string) *Builder {
	newLabels := make(map[string]string)
	for k, v := range labels {
		newLabels[k] = v
	}
	b.labels = newLabels
	return b
}

func (b *Builder) SetData(data map[string]string) *Builder {
	for k, v := range data {
		b.SetDataField(k, v)
	}
	return b
}

func (b *Builder) Build() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            b.name,
			Namespace:       b.namespace,
			OwnerReferences: b.ownerReferences,
			Labels:          b.labels,
		},
		Data: b.data,
	}
}

func (b *Builder) BuildWithOwner(owner metav1.OwnerReference) *corev1.ConfigMap {
	configMap := b.Build()
	configMap.SetOwnerReferences(append(configMap.GetOwnerReferences(), owner))
	return configMap
}

func NewBuilder() *Builder {
	return &Builder{
		data:            map[string]string{},
		ownerReferences: []metav1.OwnerReference{},
		labels:          map[string]string{},
	}
}
