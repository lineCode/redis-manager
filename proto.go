package redis_manager

import (
	"encoding/json"

	"github.com/go-redis/redis"
)

type Proto struct {
	ServerID int `json:"server_id"` // redis server id
	KeyProto
}

type KeyProto struct {
	Key  string          `json:"key"`
	TTL  int64           `json:"ttl"`
	Type string          `json:"type"`
	Size int64           `json:"size"`
	Data json.RawMessage `json:"data"`
}

type DelKeysProto struct {
	Name string   `json:"name"`
	Keys []string `json:"keys"`
}

type KVProto struct {
	Val string `json:"val"`
}

type ListProto struct {
	Val []interface{} `json:"val"`
}

type HashProto struct {
	Val map[string]interface{} `json:"val"`
}

type SetProto struct {
	Val []interface{} `json:"val"`
}

type ZsetProto struct {
	Val []redis.Z `json:"val"`
}

type RenameProto struct {
	OldKey string `json:"old_key"`
	NewKey string `json:"new_key"`
}

type SaveProto struct {
	KeyProto
	Field string `json:"field"`
	Index int64  `json:"index"` // list
}

type ScanProto struct {
	Keyword string `json:"keyword"`
	Regex   bool   `json:"regex"`
	Start   int64  `json:"start"`
	Num     int64  `json:"num"`
	Iter    int64  `json:"iter"`
	Asc     bool   `json:"asc"`
}

type listValProto struct {
	Index int64  `json:"index"`
	Val   string `json:"val"`
}

type ValueScanProto struct {
	KeyProto
	ScanProto
}

type ServerListProto struct {
	RedisBaseCfg
	Initialize        bool       `json:"initialize"`
	MemoryUsedPercent int64      `json:"memory_used_percent"`
	LastRefreshTime   int64      `json:"last_refresh_time"`
	RefreshStatus     int32      `json:"refresh_status"`
	DelTask           []*DelTask `json:"del_task"`
	Preload           bool       `json:"preload"`
}

type RemoveDelTaskProto struct {
	TaskID int64 `json:"task_id"`
}

type LoginProto struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResponseProto struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func newLoginProto(data []byte, err error) (LoginProto, error) {
	var p LoginProto
	if err != nil {
		return p, err
	}
	err = json.Unmarshal(data, &p)
	return p, err
}

func newResponseProto(code int, message string, data interface{}) *ResponseProto {
	return &ResponseProto{
		code, message, data,
	}
}

func newRemoveDelTaskProto(data []byte, err error) (RemoveDelTaskProto, error) {
	var p RemoveDelTaskProto
	if err != nil {
		return p, err
	}
	err = json.Unmarshal(data, &p)
	return p, err
}

func newKeyProto(data []byte, err error) (KeyProto, error) {
	var p KeyProto
	if err != nil {
		return p, err
	}
	err = json.Unmarshal(data, &p)
	return p, err
}

func newDelKeysProto(data []byte, err error) (DelKeysProto, error) {
	var p DelKeysProto
	if err != nil {
		return p, err
	}
	err = json.Unmarshal(data, &p)
	return p, err
}

func newSaveProto(data []byte, err error) (SaveProto, error) {
	var p SaveProto
	if err != nil {
		return p, err
	}
	err = json.Unmarshal(data, &p)
	return p, err
}

func newScanProto(data []byte, err error) (ScanProto, error) {
	var p ScanProto
	if err != nil {
		return p, err
	}
	err = json.Unmarshal(data, &p)
	return p, err
}
func newValueScanProto(data []byte, err error) (ValueScanProto, error) {
	var p ValueScanProto
	if err != nil {
		return p, err
	}
	err = json.Unmarshal(data, &p)
	return p, err
}

func newRenameProto(data []byte, err error) (RenameProto, error) {
	var p RenameProto
	if err != nil {
		return p, err
	}
	err = json.Unmarshal(data, &p)
	return p, err
}
