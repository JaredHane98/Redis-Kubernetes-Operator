/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"context"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"
	"redis.operator/pkg/kube/configmap"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RedisReplicationSpec defines the desired state of RedisReplication
type RedisReplicationSpec struct {
	//+optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	//+optional
	VolumeMounts      []corev1.VolumeMount          `json:"volumeMounts,omitempty"`
	StatefulsetConfig StatefulSetConfiguration      `json:"statefulSet,omitempty"`
	RedisConfig       RedisReplicationConfiguration `json:"config,omitempty"`
	//+optional
	TLSConfig      *RedisTLSConfiguration `json:"tls,omitempty"`
	EnableExporter bool                   `json:"enableExporter,omitempty"`
	//+optional
	RedisSentinelConfig *RedisReplicationSentinelConfig `json:"sentinelConfig,omitempty"`
}

type RedisReplicationSentinelConfig struct {
	RedisSentinelName string `json:"redisSentinelName,omitempty"`
	//+optional
	RedisSentinelDowntime *int `json:"redisSentinelDowntime,omitempty"`
}

type RedisReplicationConfiguration struct {
	RedisConfigurationData `json:",inline"`
}

// RedisReplicationStatus defines the observed state of RedisReplication
type RedisReplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	MasterDns string `json:"masterNode,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// RedisReplication is the Schema for the redisreplications API
type RedisReplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisReplicationSpec   `json:"spec,omitempty"`
	Status RedisReplicationStatus `json:"status,omitempty"`
}

const (
	RedisReplicationFinalizer = "redis-operator.redisreplication.k8s.example.com/finalizer"
)

func (r *RedisReplication) GetConfigName() string {
	return r.Name + "-config"
}

func (r *RedisReplication) IsStatefulSetReady(ctx context.Context, k8Client kubernetes.Interface) bool {

	typeMeta := metav1.TypeMeta{Kind: "StatefulSet", APIVersion: "apps/v1"}
	statefulset, err := k8Client.AppsV1().StatefulSets(r.Namespace).Get(ctx, r.Name, metav1.GetOptions{TypeMeta: typeMeta})
	if err != nil {
		return false
	}
	if statefulset.Status.CurrentRevision != statefulset.Status.UpdateRevision {
		return false
	}
	if statefulset.Status.ObservedGeneration != statefulset.ObjectMeta.Generation {
		return false
	}

	return true
}

func (r *RedisReplication) GetRedisPort() string {
	var portQuery string
	if r.Spec.TLSConfig != nil {
		portQuery = "tls-port"
	} else {
		portQuery = "port"
	}

	if port := configmap.GetConfigMapValue(r.Spec.RedisConfig.Data, "redis.conf", portQuery); port != "" {
		return port
	}

	return "6379"
}

func (r *RedisReplication) GetRedisPortInt32() int32 {
	portStr := r.GetRedisPort()
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0
	}
	return int32(port)
}

func (r *RedisReplication) GetHeadlessServiceName() string {
	return r.Name + "-headless"
}

func (r *RedisReplication) GetServiceName() string {
	return r.Name + "-service"
}

func (r *RedisReplication) GetOwnerReference() metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: r.APIVersion,
		Kind:       r.Kind,
		Name:       r.Name,
		UID:        r.UID,
		Controller: ptr.To(true),
	}
}

// +kubebuilder:object:root=true

// RedisReplicationList contains a list of RedisReplication
type RedisReplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisReplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedisReplication{}, &RedisReplicationList{})
}
