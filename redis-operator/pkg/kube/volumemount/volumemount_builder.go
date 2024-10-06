package volumemount

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

type Builder struct {
	volumeMounts corev1.VolumeMount
}

func (r *Builder) SetName(name string) *Builder {
	r.volumeMounts.Name = name
	return r
}

func (r *Builder) SetReadOnly(readOnly bool) *Builder {
	r.volumeMounts.ReadOnly = readOnly
	return r
}

func (r *Builder) SetRecursiveReadOnly(setting corev1.RecursiveReadOnlyMode) *Builder {
	r.volumeMounts.RecursiveReadOnly = ptr.To(setting)
	return r
}
func (r *Builder) SetSubPath(subPath string) *Builder {
	r.volumeMounts.SubPath = subPath
	return r
}

func (r *Builder) SetMountPath(path string) *Builder {
	r.volumeMounts.MountPath = path
	return r
}

func (r *Builder) SetsubPathExpr(expr string) *Builder {
	r.volumeMounts.SubPathExpr = expr
	return r
}

func (r *Builder) Build() corev1.VolumeMount {
	return r.volumeMounts
}

func NewBuilder() *Builder {
	return &Builder{
		volumeMounts: corev1.VolumeMount{},
	}
}
