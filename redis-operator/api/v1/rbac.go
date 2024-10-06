package v1

// +kubebuilder:rbac:groups=redis.redis.operator,resources=redisreplications;redissentinels,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redis.redis.operator,resources=redisreplications/status;redissentinels/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redis.redis.operator,resources=redisreplications/finalizers;redissentinels/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets;endpoints;pods;events;secrets;configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods;configmaps;secrets;services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
