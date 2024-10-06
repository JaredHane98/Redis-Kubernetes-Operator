package k8sredis

import (
	v1 "redis.operator/api/v1"
)

const (
	defaultConfig = `daemonize no
        			 pidfile /var/run/redis-sentinel.pid
        			 logfile ""
        			 dir /tmp
        			 acllog-max-len 128
        			 sentinel deny-scripts-reconfig yes
        			 sentinel resolve-hostnames no
        			 sentinel announce-hostnames no
        			 protected-mode "no"
        			 port 26379`
)

func GetSentinelConfigMap(config *v1.RedisSentinelConfiguration) *v1.RedisSentinelConfiguration {

	if config == nil {
		return &v1.RedisSentinelConfiguration{
			RedisConfigurationData: v1.RedisConfigurationData{
				Data: map[string]string{
					"sentinel.conf": defaultConfig,
				},
			},
		}
	}

	if _, ok := config.Data["sentinel.conf"]; !ok {
		config.Data["sentinel.conf"] = defaultConfig
	}
	return config
}
