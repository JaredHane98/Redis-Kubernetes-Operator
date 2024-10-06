package v1

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/stretchr/objx"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// wrapper around statefulset
type StatefulSetConfiguration struct {
	// +kubebuilder:pruning:PreserveUnknownFields
	Wrapper StatefulSpecWrapper `json:"spec"`
	//+optional
	MetaData StatefulSetMetadataWrapper `json:"metadata"`
}

func (r *StatefulSetConfiguration) GetReplicas() int {
	if r.Wrapper.Spec.Replicas != nil {
		return int(*r.Wrapper.Spec.Replicas)
	} else {
		return 1
	}
}

type RedisConfigurationData struct {
	Data map[string]string `json:"data"`
}

type TLSConfig struct {
	Cert       string
	Key        string
	CACert     string
	Password   string
	PassPhrase string
}

func (r *RedisConfigurationData) GetConfigMapTLS() (*TLSConfig, error) {
	tls := TLSConfig{}
	var err error

	if tls.Cert, err = r.GetValue("redis.conf", "tls-client-cert-file"); err != nil {
		return nil, err
	}
	if tls.Cert == "" {
		if tls.Cert, err = r.GetValue("redis.conf", "tls-cert-file"); err != nil {
			return nil, err
		}
	}

	if tls.Key, err = r.GetValue("redis.conf", "tls-client-key-file"); err != nil {
		return nil, err
	}
	if tls.Key == "" {
		if tls.Key, err = r.GetValue("redis.conf", "tls-key-file"); err != nil {
			return nil, err
		}
	}

	if tls.PassPhrase, err = r.GetValue("redis.conf", "tls-client-passphrase"); err != nil {
		return nil, err
	}
	if tls.PassPhrase == "" {
		if tls.PassPhrase, err = r.GetValue("redis.conf", "tls-passphrase"); err != nil {
			return nil, err
		}
	}

	if tls.CACert, err = r.GetValue("redis.conf", "tls-ca-cert-file"); err != nil {
		return nil, err
	}

	if tls.Password, err = r.GetValue("redis.conf", "requirepass"); err != nil {
		return nil, err
	}

	if tls.Cert == "" || tls.Key == "" || tls.CACert == "" || tls.Password == "" {
		return nil, fmt.Errorf("error, missing tls key. %v", tls)
	}

	return &tls, nil
}

func (r *RedisConfigurationData) GetValues(key string, substr string) ([]string, error) {
	if foundValue, ok := r.Data[key]; ok {
		substrFields := strings.Fields(substr)
		scanner := bufio.NewScanner(strings.NewReader(foundValue))
		for scanner.Scan() {

			line := scanner.Text()
			fields := strings.Fields(line)

			if len(fields) < len(substrFields) {
				continue
			}

			found := true
			for index, sub := range substrFields {
				if sub != fields[index] {
					found = false
					break
				}
			}

			if found {
				return fields[len(substrFields):], nil
			}

			if err := scanner.Err(); err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

func (r *RedisConfigurationData) GetValue(key string, substr string) (string, error) {
	values, err := r.GetValues(key, substr)
	if err != nil {
		return "", err
	}
	if len(values) == 1 {
		return values[0], nil
	}
	return "", nil
}

func (r *RedisConfigurationData) UpdateValue(key string, substr string, value string) bool {

	substrFields := strings.Fields(substr)
	if foundValue, ok := r.Data[key]; ok {

		lines := strings.Split(foundValue, "\n")

		for i, line := range lines {
			fields := strings.Fields(line)

			if len(fields) < len(substrFields) {
				continue
			}

			found := true
			for index, sub := range substrFields {
				if sub != fields[index] {
					found = false
					break
				}
			}

			if found {
				lines[i] = substr + " " + value
				r.Data[key] = strings.Join(lines, "\n")
				return true
			}
		}
	}
	r.Data[key] = r.Data[key] + "\n" + substr + " " + value
	return false
}

type RedisConfigMapWrapper struct {
	MapWrapper `json:"-"`
}

type MapWrapper struct {
	Object map[string]interface{} `json:"-"`
}

// MarshalJSON defers JSON encoding to the wrapped map
func (m *MapWrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Object)
}

// UnmarshalJSON will decode the data into the wrapped map
func (m *MapWrapper) UnmarshalJSON(data []byte) error {
	if m.Object == nil {
		m.Object = map[string]interface{}{}
	}

	// Handle keys like net.port to be set as nested maps.
	// Without this after unmarshalling there is just key "net.port" which is not
	// a nested map and methods like GetPort() cannot access the value.
	tmpMap := map[string]interface{}{}
	err := json.Unmarshal(data, &tmpMap)
	if err != nil {
		return err
	}

	for k, v := range tmpMap {
		m.SetOption(k, v)
	}

	return nil
}

func (m *MapWrapper) DeepCopy() *MapWrapper {
	if m != nil && m.Object != nil {
		return &MapWrapper{
			Object: runtime.DeepCopyJSON(m.Object),
		}
	}
	c := NewMapWrapper()
	return &c
}

// NewMapWrapper returns an empty MapWrapper
func NewMapWrapper() MapWrapper {
	return MapWrapper{Object: map[string]interface{}{}}
}

func (m MapWrapper) SetOption(key string, value interface{}) MapWrapper {
	m.Object = objx.New(m.Object).Set(key, value)
	return m
}

type StatefulSpecWrapper struct {
	Spec appsv1.StatefulSetSpec `json:"-"`
}

// MarshalJSON defers JSON encoding to the wrapped map
func (m *StatefulSpecWrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Spec)
}

// UnmarshalJSON will decode the data into the wrapped map
func (m *StatefulSpecWrapper) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &m.Spec)
}

func (m *StatefulSpecWrapper) DeepCopy() *StatefulSpecWrapper {
	return &StatefulSpecWrapper{
		Spec: m.Spec,
	}
}

type StatefulSetMetadataWrapper struct {
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

type RedisTLSConfiguration struct {
	Name       string `json:"name"`
	SecretName string `json:"secretName"`
}
