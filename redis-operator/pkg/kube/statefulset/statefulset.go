package statefulset

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func GetProbe(enableTLS, enableAuth bool) *corev1.Probe {
	healthChecker := []string{
		"redis-cli",
		"-h", "$(hostname)",
		"-p", "${REDIS_PORT}",
	}

	if enableAuth {
		healthChecker = append(healthChecker, "-a", "${REDIS_PASSWORD}")
	}
	if enableTLS {
		healthChecker = append(healthChecker, "--tls", "--cert", "${REDIS_TLS_CERT}", "--key", "${REDIS_TLS_CERT_KEY}", "--cacert", "${REDIS_TLS_CA_KEY}")
	}

	healthChecker = append(healthChecker, "ping")

	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"sh", "-c", strings.Join(healthChecker, " ")},
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}
