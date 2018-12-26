package server

type Config struct {
	Debug       bool          `json:"debug" yaml:"debug"`
	Listen      string        `json:"listen" yaml:"listen"`
	CacheDir    string        `json:"cache_dir" yaml:"cache_dir"`
	CacheSize   int           `json:"cache_size" yaml:"cache_size"`
	User        []UserInfo    `json:"user" yaml:"user"`
	RedisServer []RedisConfig `json:"redis_server" yaml:"redis_server"`
}
