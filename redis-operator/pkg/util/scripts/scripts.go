package scripts

import "fmt"

const PingScriptAuth = `
result=$(redis-cli -p "%s" --tls --cert "%s" --key "%s" --cacert "%s" -h localhost <<EOF
AUTH "%s"
PING
EOF
)

if echo "$result" | grep -q "PONG"; then
    exit 0
else
    exit 1
fi
`

const PingScriptNoAuthPassword = `
result=$(redis-cli -p "%s" <<EOF
AUTH "%s"
PING
EOF
)

if echo "$result" | grep -q "PONG"; then
    exit 0
else
    exit 1
fi
`

const PingScriptNoAuth = `
result=$(redis-cli -p "%s" <<EOF
PING
EOF
)

if echo "$result" | grep -q "PONG"; then
    exit 0
else
    exit 1
fi
`

const SyncScriptAuth = `
result=$(redis-cli -p "%s" --tls --cert "%s" --key "%s" --cacert "%s" -h localhost <<EOF
AUTH "%s"
INFO replication
EOF
)

role=$(echo "$result" | grep -oP 'role:\K\w+')
sync_in_progress=$(echo "$result" | grep -oP 'master_sync_in_progress:\K\d+')
if [ "$role" = "master" ] || [ "$sync_in_progress" == "0" ]; then
	exit 0
else
	exit 1
fi
`

const SyncScriptNoAuthPassword = `
result=$(redis-cli -p "%s" <<EOF
AUTH "%s"
INFO replication
EOF
)

role=$(echo "$result" | grep -oP 'role:\K\w+')
sync_in_progress=$(echo "$result" | grep -oP 'master_sync_in_progress:\K\d+')
if [ "$role" = "master" ] || [ "$sync_in_progress" == "0" ]; then
	exit 0
else
	exit 1
fi
`

const SyncScriptNoAuth = `
result=$(redis-cli -p "%s" <<EOF
INFO replication
EOF
)

role=$(echo "$result" | grep -oP 'role:\K\w+')
sync_in_progress=$(echo "$result" | grep -oP 'master_sync_in_progress:\K\d+')
if [ "$role" = "master" ] || [ "$sync_in_progress" == "0" ]; then
	exit 0
else
	exit 1
fi
`

const ReadinessScriptAuth = `
result=$(redis-cli -p "%s" --tls --cert "%s" --key "%s" --cacert "%s" -h localhost <<EOF
AUTH "%s"
INFO replication
EOF
)

role=$(echo "$result" | grep -oP 'role:\K\w+')
if [ "$role" = "master" ]; then
	exit 0
else
	exit 1
fi
`

const ReadinessScriptNoAuthPassword = `
result=$(redis-cli -p "%s" <<EOF
AUTH "%s"
INFO replication
EOF
)

role=$(echo "$result" | grep -oP 'role:\K\w+')
if [ "$role" = "master" ]; then
	exit 0
else
	exit 1
fi
`

const ReadinessScriptNoAuth = `
result=$(redis-cli -p "%s" <<EOF
INFO replication
EOF
)

role=$(echo "$result" | grep -oP 'role:\K\w+')
if [ "$role" = "master" ]; then
	exit 0
else
	exit 1
fi
`

const DownTimeScriptAuth = `
s_down_time=$(redis-cli -p "%s" --tls --cert "%s" --key "%s" --cacert "%s" -h localhost <<EOF
AUTH "%s"
SENTINEL MASTERS
EOF
)

s_down_time=$(echo "$s_down_time" | grep -A1 's-down-time' | tail -n1 | tr -d ' ')

if [ -z "$s_down_time" ]; then
    exit 0
elif [ "$s_down_time" -gt 5000 ]; then
    exit 1
else
    exit 0
fi
`

const DownTimeScriptNoAuthPassword = `
s_down_time=$(redis-cli -p "%s" -h localhost <<EOF
AUTH "%s"
SENTINEL MASTERS
EOF
)

s_down_time=$(echo "$s_down_time" | grep -A1 's-down-time' | tail -n1 | tr -d ' ')

if [ -z "$s_down_time" ]; then
    exit 0
elif [ "$s_down_time" -gt 5000 ]; then
    exit 1
else
    exit 0
fi
`

const DownTimeScriptNoAuth = `
s_down_time=$(redis-cli -p "%s" -h localhost <<EOF
SENTINEL MASTERS
EOF
)

s_down_time=$(echo "$s_down_time" | grep -A1 's-down-time' | tail -n1 | tr -d ' ')

if [ -z "$s_down_time" ]; then
    exit 0
elif [ "$s_down_time" -gt 5000 ]; then
    exit 1
else
    exit 0
fi
`

func GetPingScriptAuth(port, cert, key, cacert, password string) []string {
	return []string{"/bin/sh", "-c", fmt.Sprintf(PingScriptAuth, port, cert, key, cacert, password)}
}

func GetPingScript(port, password string) []string {
	if password == "" {
		return []string{"/bin/sh", "-c", fmt.Sprintf(PingScriptNoAuth, port)}
	}
	return []string{"/bin/sh", "-c", fmt.Sprintf(PingScriptNoAuthPassword, port, password)}
}

func GetReplicaSyncScriptAuth(port, cert, key, cacert, password string) []string {
	return []string{"/bin/sh", "-c", fmt.Sprintf(SyncScriptAuth, port, cert, key, cacert, password)}
}

func GetReplicaSyncScript(port, password string) []string {
	if password == "" {
		return []string{"/bin/sh", "-c", fmt.Sprintf(SyncScriptNoAuth, port)}
	}
	return []string{"/bin/sh", "-c", fmt.Sprintf(SyncScriptNoAuthPassword, port, password)}
}

func GetDownTimeScriptAuth(port, cert, key, cacert, password string) []string {
	return []string{"/bin/sh", "-c", fmt.Sprintf(DownTimeScriptAuth, port, cert, key, cacert, password)}
}

func GetDownTimeScript(port, password string) []string {
	if password == "" {
		return []string{"/bin/sh", "-c", fmt.Sprintf(DownTimeScriptNoAuth, port)}
	}
	return []string{"/bin/sh", "-c", fmt.Sprintf(DownTimeScriptNoAuthPassword, port, password)}
}

func GetReplicaReadinessScriptAuth(port, cert, key, cacert, password string) []string {
	return []string{"/bin/sh", "-c", fmt.Sprintf(ReadinessScriptAuth, port, cert, key, cacert, password)}
}

func GetReplicaReadinessScript(port, password string) []string {
	if password == "" {
		return []string{"/bin/sh", "-c", fmt.Sprintf(ReadinessScriptNoAuth, port)}
	}
	return []string{"/bin/sh", "-c", fmt.Sprintf(ReadinessScriptNoAuthPassword, port, password)}
}
