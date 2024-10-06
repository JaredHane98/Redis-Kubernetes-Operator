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
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"redis.operator/pkg/kube/configmap"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RedisSentinelSpec defines the desired state of RedisSentinel
type RedisSentinelSpec struct {
	//+optional
	Resources            *corev1.ResourceRequirements `json:"resources,omitempty"`
	StatefulsetConfig    StatefulSetConfiguration     `json:"statefulSet,omitempty"`
	MasterName           string                       `json:"masterName,omitempty"`
	RedisReplicationName string                       `json:"redisReplicationName,omitempty"`
	RedisSentinelQuorum  int                          `json:"redisSentinelQuorum,omitempty"`
	RedisConfig          RedisSentinelConfiguration   `json:"config,omitempty"`
}

type RedisSentinelConfiguration struct {
	RedisConfigurationData `json:",inline"`
}

// RedisSentinelStatus defines the observed state of RedisSentinel
type RedisSentinelStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// RedisSentinel is the Schema for the redissentinels API
type RedisSentinel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisSentinelSpec   `json:"spec,omitempty"`
	Status RedisSentinelStatus `json:"status,omitempty"`
}

const (
	RedisSentinelFinalizer = "redis-operator.redissentinel.k8s.example.com/finalizer"
)

func (r *RedisSentinel) GetHeadlessServiceName() string {
	return r.Name + "-headless"
}

func (r *RedisSentinel) GetServiceName() string {
	return r.Name + "-service"
}

func (r *RedisSentinel) GetConfigName() string {
	return r.Name + "-conf"
}

func (r *RedisSentinel) GetRedisPort() string {

	if port := configmap.GetConfigMapValue(r.Spec.RedisConfig.Data, "sentinel.conf", "tls-port"); port != "" {
		return port
	}
	if port := configmap.GetConfigMapValue(r.Spec.RedisConfig.Data, "sentinel.conf", "port"); port != "" {
		return port
	}
	return "26379"
}

func (r *RedisSentinel) GetRedisPortInt32() int32 {
	portStr := r.GetRedisPort()
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 26379
	}
	return int32(port)
}

func (r *RedisSentinel) GetOwnerReference() metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: r.APIVersion,
		Kind:       r.Kind,
		Name:       r.Name,
		UID:        r.UID,
		Controller: ptr.To(true),
	}
}

func (r *RedisSentinel) GetSentinelName() string {
	return r.Name + "-sentinel"
}

// +kubebuilder:object:root=true

// RedisSentinelList contains a list of RedisSentinel
type RedisSentinelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisSentinel `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedisSentinel{}, &RedisSentinelList{})
}
