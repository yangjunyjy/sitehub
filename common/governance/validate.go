package governance

import (
	"fmt"
	"github.com/zeromicro/go-zero/zrpc"
	"reflect"
	"strings"
	"time"
)

// Validate prevents go-zero's Endpoints > Target > Etcd precedence from silently
// bypassing Nacos, and ensures the process has time to withdraw before draining.
func Validate(value any, c Config) error {
	if !c.Registry.Enabled {
		return nil
	}
	var walk func(reflect.Value) error
	walk = func(v reflect.Value) error {
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return nil
			}
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			return nil
		}
		if v.CanInterface() {
			if rpc, ok := v.Interface().(zrpc.RpcClientConf); ok {
				if len(rpc.Endpoints) > 0 || len(rpc.Etcd.Hosts) > 0 || !strings.HasPrefix(rpc.Target, "nacos:///") {
					return fmt.Errorf("registry mode requires every RPC client to use only a nacos:/// target")
				}
				return nil
			}
			if rpc, ok := v.Interface().(zrpc.RpcServerConf); ok && len(rpc.Etcd.Hosts) > 0 {
				return fmt.Errorf("Nacos and native etcd registration cannot both be enabled")
			}
		}
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).IsExported() {
				if err := walk(v.Field(i)); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := walk(reflect.ValueOf(value)); err != nil {
		return err
	}
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if mode := v.FieldByName("Mode"); mode.IsValid() && mode.String() == "pro" && c.Registry.AllowLoopback {
		return fmt.Errorf("production registry cannot allow loopback addresses")
	}
	if shutdown := v.FieldByName("Shutdown"); shutdown.IsValid() {
		wrap := shutdown.FieldByName("WrapUpTime")
		wait := shutdown.FieldByName("WaitTime")
		if c.Nacos.TimeoutMs > 60000 {
			return fmt.Errorf("Nacos.TimeoutMs must not exceed 60000 in registry mode")
		}
		minimumWrap := max(10*time.Second, time.Duration(c.Nacos.TimeoutMs)*time.Millisecond+5*time.Second)
		if wrap.IsValid() && time.Duration(wrap.Int()) < minimumWrap {
			return fmt.Errorf("Shutdown.WrapUpTime must allow the Nacos timeout plus 5s propagation (minimum 10s)")
		}
		if wait.IsValid() && wrap.IsValid() && time.Duration(wait.Int()) < time.Duration(wrap.Int())+5*time.Second {
			return fmt.Errorf("Shutdown.WaitTime must exceed WrapUpTime by at least 5s")
		}
	}
	return nil
}
