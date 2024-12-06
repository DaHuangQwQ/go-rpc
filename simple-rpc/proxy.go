package simple_rpc

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
)

// initClientProxy go 的代理模式
func initClientProxy(addr string, service Service) error {
	client, err := NewClient(addr)
	if err != nil {
		return err
	}
	return setFuncField(service, client)
}

func setFuncField(service Service, p Proxy) error {
	if service == nil {
		return errors.New("service is nil")
	}

	typ := reflect.TypeOf(service)
	val := reflect.ValueOf(service)

	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
		val = val.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return errors.New("service must be struct")
	}

	numField := val.NumField()

	for i := 0; i < numField; i++ {
		fieldTyp := typ.Field(i)
		fieldVal := val.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		fn := reflect.MakeFunc(fieldTyp.Type, func(args []reflect.Value) (results []reflect.Value) {
			retVal := reflect.New(fieldTyp.Type.Out(0).Elem())
			ctx := args[0].Interface().(context.Context)
			reqData, err := json.Marshal(args[1].Interface())
			if err != nil {
				return []reflect.Value{retVal, reflect.ValueOf(err)}
			}
			req := &Request{
				ServiceName: service.Name(),
				MethodName:  fieldTyp.Name,
				Arg:         reqData,
			}

			resp, err := p.Invoke(ctx, req)
			if err != nil {
				return []reflect.Value{retVal, reflect.ValueOf(err)}
			}

			err = json.Unmarshal(resp.Data, retVal.Interface())
			if err != nil {
				return []reflect.Value{retVal, reflect.ValueOf(err)}
			}

			return []reflect.Value{retVal, reflect.Zero(reflect.TypeOf(new(error)).Elem())}
		})

		fieldVal.Set(fn)
	}
	return nil
}
