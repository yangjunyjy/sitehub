package governance

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/zrpc"
)

// Validate keeps provider selection behind the registry proxy rather than a
// competing framework resolver, and reserves time to withdraw before draining.
func Validate(value any, c Config) error {
	if !c.Registry.Enabled {
		return nil
	}
	if err := c.ControlPlane.Validate(); err != nil {
		return err
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
				if len(rpc.Endpoints) > 0 || len(rpc.Etcd.Hosts) > 0 || !strings.HasPrefix(rpc.Target, "discovery:///") {
					return fmt.Errorf("registry proxy requires every RPC client to use only a discovery:/// target")
				}
				return nil
			}
			if rpc, ok := v.Interface().(zrpc.RpcServerConf); ok && len(rpc.Etcd.Hosts) > 0 {
				return fmt.Errorf("registry proxy and native framework registration cannot both be enabled")
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
		minimumWrap := max(10*time.Second, c.ControlPlane.RequestTimeout+5*time.Second, time.Duration(c.ControlPlane.LeaseTTL)*time.Second+5*time.Second)
		if wrap.IsValid() && time.Duration(wrap.Int()) < minimumWrap {
			return fmt.Errorf("Shutdown.WrapUpTime must be at least %s (configured %s; includes request timeout, lease expiry and 5s propagation)", minimumWrap, time.Duration(wrap.Int()))
		}
		if wait.IsValid() && wrap.IsValid() && time.Duration(wait.Int()) < time.Duration(wrap.Int())+5*time.Second {
			return fmt.Errorf("Shutdown.WaitTime must exceed WrapUpTime by at least 5s")
		}
	}
	return nil
}
