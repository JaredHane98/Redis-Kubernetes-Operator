package probe

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

type Builder struct {
	probe corev1.Probe
}

func (b *Builder) SetInitialDelaySeconds(initialDelaySeconds int32) *Builder {
	b.probe.InitialDelaySeconds = initialDelaySeconds
	return b
}

func (b *Builder) SetTimeoutSeconds(timeoutSeconds int32) *Builder {
	b.probe.TimeoutSeconds = timeoutSeconds
	return b
}

func (b *Builder) SetPeriodSeconds(periodSeconds int32) *Builder {
	b.probe.PeriodSeconds = periodSeconds
	return b
}

func (b *Builder) SetSuccessThreshold(successThreshold int32) *Builder {
	b.probe.SuccessThreshold = successThreshold
	return b
}

func (b *Builder) SetFailureThreshold(failureThreshold int32) *Builder {
	b.probe.FailureThreshold = failureThreshold
	return b
}

func (b *Builder) SetTerminationGracePeriodSeconds(terminationGracePeriodSeconds int64) *Builder {
	b.probe.TerminationGracePeriodSeconds = ptr.To(terminationGracePeriodSeconds)
	return b
}

func (b *Builder) SetExecAction(command []string) *Builder {
	b.probe.Exec = &corev1.ExecAction{
		Command: command,
	}
	return b
}

func (b *Builder) SetHttpGetAction(path string, port int32, host string, scheme corev1.URIScheme, HttpHeaders []corev1.HTTPHeader) *Builder {
	b.probe.HTTPGet = &corev1.HTTPGetAction{
		Path:        path,
		Port:        intstr.FromInt32(port),
		Host:        host,
		Scheme:      scheme,
		HTTPHeaders: HttpHeaders,
	}
	return b
}

func (b *Builder) SetTcpSocketAction(host string, port int32) *Builder {
	b.probe.TCPSocket = &corev1.TCPSocketAction{
		Host: host,
		Port: intstr.FromInt32(port),
	}
	return b
}

func (b *Builder) SetGRPCAction(port int32, service string) *Builder {
	b.probe.GRPC = &corev1.GRPCAction{
		Port:    port,
		Service: ptr.To(service),
	}
	return b
}

func (b *Builder) Build() *corev1.Probe {
	return &b.probe
}

func NewBuilder() *Builder {
	return &Builder{
		probe: corev1.Probe{},
	}
}
