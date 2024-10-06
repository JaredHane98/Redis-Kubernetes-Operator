package k8sredis

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"crypto/tls"
	"crypto/x509"

	"github.com/go-logr/logr"
	"github.com/redis/go-redis/v9"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "redis.operator/api/v1"
)

type RedisCommandInfo struct {
	Info     map[string]string
	DNS      string
	PodIndex int
}

func GetClient(ip string, port string, tlsConfig *tls.Config, password string, timeout time.Duration) *redis.Client {
	opts := &redis.Options{
		Addr:         ip + ":" + port,
		Password:     password, // "supersecretpasswordnobodywillguess",
		DB:           0,
		TLSConfig:    tlsConfig,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		DialTimeout:  timeout,
	}
	return redis.NewClient(opts)
}

func GetSentinelClient(ip string, port string, tlsConfig *tls.Config, password string, timeout time.Duration) *redis.SentinelClient {
	opts := &redis.Options{
		Addr:         ip + ":" + port,
		Password:     password,
		DB:           0,
		TLSConfig:    tlsConfig,
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
		DialTimeout:  timeout,
	}
	return redis.NewSentinelClient(opts)
}

func ConvertRedisInterfaceToMap(value string) (map[string]string, error) {

	redisMap := make(map[string]string)

	// get the start and end of the sequence
	sequenceStart := strings.Index(value, "map")
	if sequenceStart == -1 {
		return nil, fmt.Errorf("failed to find start of map sequence")
	}

	mapSequence := value[sequenceStart+4:]

	regex := regexp.MustCompile(`[\[\]]`)
	mapSequence = regex.ReplaceAllString(mapSequence, "")

	lines := strings.Split(mapSequence, " ")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			values := strings.Split(line, ":")
			if len(values) == 2 {
				redisMap[values[0]] = values[1]
			}
		}
	}
	return redisMap, nil
}

func GetReplicationInfo(client *redis.Client, ctx context.Context) (map[string]string, error) {
	info, err := client.Info(ctx, "Replication").Result()
	if err != nil {
		return nil, err
	}

	redisInfo := map[string]string{}

	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			values := strings.Split(line, ":")
			if len(values) == 2 {
				redisInfo[values[0]] = values[1]
			}
		}
	}
	return redisInfo, nil
}

func GetTLSConfig(ctx context.Context, k8Client kubernetes.Interface, secretName string, configmap map[string]string, namespace string) (*tls.Config, error) {
	secret, err := k8Client.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	tlsCert, ok := secret.Data["tls.crt"]
	if !ok {
		return nil, fmt.Errorf("ca.crt not found in secret")
	}

	tlsKey, ok := secret.Data["tls.key"]
	if !ok {
		return nil, fmt.Errorf("tls.key not found in secret")
	}

	caCert, ok := secret.Data["ca.crt"]
	if !ok {
		return nil, fmt.Errorf("ca.crt not found in secret")
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA cert")
	}

	cert, err := tls.X509KeyPair(tlsCert, tlsKey)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
		RootCAs:      caCertPool,
	}, nil
}

func GetSentinelMasters(ctx context.Context, k8Client kubernetes.Interface, instance *v1.RedisSentinel, replicaInstance *v1.RedisReplication) ([]RedisCommandInfo, error) {

	tlsConfig := &tls.Config{}
	password, err := instance.Spec.RedisConfig.GetValue("sentinel.conf", "requirepass")
	if err != nil {
		return nil, err
	}

	if replicaInstance.Spec.TLSConfig != nil {
		if tlsConfig, err = GetTLSConfig(ctx, k8Client, replicaInstance.Spec.TLSConfig.SecretName, replicaInstance.Spec.RedisConfig.Data, replicaInstance.Namespace); err != nil {
			return nil, err
		}
	}

	replicas := instance.Spec.StatefulsetConfig.GetReplicas()
	replicaInfo := []RedisCommandInfo{}

	for i := 0; i < replicas; i++ {
		podDNS := fmt.Sprintf("%s-%d.%s.%s.svc.cluster.local", instance.Name, i, instance.GetHeadlessServiceName(), instance.Namespace)

		redisClient := GetSentinelClient(podDNS, instance.GetRedisPort(), tlsConfig, password, time.Second*1)
		defer redisClient.Close()

		if redisClient.Ping(ctx).Val() != "PONG" {
			continue
		}

		slicecmd := redisClient.Masters(ctx)
		err = slicecmd.Err()
		if err != nil {
			return nil, err
		}

		cmdMap, err := ConvertRedisInterfaceToMap(slicecmd.String())
		if err != nil {
			return nil, err
		}

		replicaInfo = append(replicaInfo, RedisCommandInfo{Info: cmdMap, DNS: podDNS, PodIndex: i})
	}
	return replicaInfo, nil
}

func GetReplicaInfo(ctx context.Context, k8Client kubernetes.Interface, instance *v1.RedisReplication, reqLogger logr.Logger) ([]RedisCommandInfo, error) {

	tlsConfig := &tls.Config{}
	var err error

	if instance.Spec.TLSConfig != nil {
		if tlsConfig, err = GetTLSConfig(ctx, k8Client, instance.Spec.TLSConfig.SecretName, instance.Spec.RedisConfig.Data, instance.Namespace); err != nil {
			return nil, err
		}
	}

	password, err := instance.Spec.RedisConfig.GetValue("redis.conf", "requirepass")
	if err != nil {
		return nil, err
	}

	replicas := instance.Spec.StatefulsetConfig.GetReplicas()
	replicaInfo := []RedisCommandInfo{}

	for i := 0; i < replicas; i++ {

		podDNS := fmt.Sprintf("%s-%d.%s.%s.svc.cluster.local", instance.Name, i, instance.GetHeadlessServiceName(), instance.Namespace)

		redisClient := GetClient(podDNS, instance.GetRedisPort(), tlsConfig, password, time.Second*1)
		defer redisClient.Close()

		if result := redisClient.Ping(ctx); result.Val() != "PONG" {
			continue // down, ignore
		}

		info, err := GetReplicationInfo(redisClient, ctx)
		if err != nil {
			return nil, err
		}

		replicaInfo = append(replicaInfo, RedisCommandInfo{Info: info, DNS: podDNS, PodIndex: i})
	}
	return replicaInfo, err
}

func SetReplicationMaster(ctx context.Context, k8Client kubernetes.Interface, instance *v1.RedisReplication, masterDNS string, reqLogger logr.Logger) error {
	var tlsConfig *tls.Config = nil
	var err error

	if instance.Spec.TLSConfig != nil {
		if tlsConfig, err = GetTLSConfig(ctx, k8Client, instance.Spec.TLSConfig.SecretName, instance.Spec.RedisConfig.Data, instance.Namespace); err != nil {
			return err
		}
	}

	password, err := instance.Spec.RedisConfig.GetValue("redis.conf", "requirepass")
	if err != nil {
		return err
	}

	replicas := instance.Spec.StatefulsetConfig.GetReplicas()
	for i := 0; i < replicas; i++ {

		podDNS := fmt.Sprintf("%s-%d.%s.%s.svc.cluster.local", instance.Name, i, instance.GetHeadlessServiceName(), instance.Namespace)

		redisClient := GetClient(podDNS, instance.GetRedisPort(), tlsConfig, password, time.Second*1)
		defer redisClient.Close()

		if masterDNS == podDNS {
			if err := redisClient.SlaveOf(ctx, "NO", "ONE").Err(); err != nil {
				return fmt.Errorf("error setting replication master: %v", err)
			}

		} else {
			if err := redisClient.SlaveOf(ctx, masterDNS, instance.GetRedisPort()).Err(); err != nil {
				reqLogger.Info("failed to set replication master. slave is probably down ", "error", err)
			}
		}
	}

	return nil
}
