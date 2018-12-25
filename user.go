package redis_manager

type UserInfo struct {
	Username string   `json:"username" yaml:"username"`
	Password string   `json:"password" yaml:"password"`
	Roles    []string `json:"roles" yaml:"roles"`
}
