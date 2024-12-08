package go_rpc

import (
	"context"
	"errors"
	"go-rpc/message"
	"go-rpc/serialize"
	"reflect"
)

// initClientProxy go 的代理模式 for client
func initClientProxy(addr string, service Service) error {
	client, err := NewClient(addr)
	if err != nil {
		return err
	}
	return setFuncField(service, client, &serialize.JsonSerializer{})
}

func setFuncField(service Service, p Proxy, s serialize.Serializer) error {
	if service == nil {
		return errors.New("go-rpc: service is nil")
	}

	typ := reflect.TypeOf(service)
	val := reflect.ValueOf(service)

	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
		val = val.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return errors.New("go-rpc: service must be struct")
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
			reqData, err := s.Encode(args[1].Interface())
			if err != nil {
				return []reflect.Value{retVal, reflect.ValueOf(err)}
			}
			req := &message.Request{
				ServiceName: service.Name(),
				MethodName:  fieldTyp.Name,
				Data:        reqData,
				Serializer:  s.Code(),
			}

			resp, err := p.Invoke(ctx, req)
			if err != nil {
				return []reflect.Value{retVal, reflect.ValueOf(err)}
			}

			var retErr error
			if len(resp.Error) > 0 {
				retErr = errors.New(string(resp.Error))
			}

			if len(resp.Data) > 0 {
				err = s.Decode(resp.Data, retVal.Interface())
				if err != nil {
					// 反序列化的 error
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}
			}

			var retErrVal reflect.Value
			if retErr == nil {
				retErrVal = reflect.Zero(reflect.TypeOf(new(error)).Elem())
			} else {
				retErrVal = reflect.ValueOf(retErr)
			}

			return []reflect.Value{retVal, retErrVal}
		})

		fieldVal.Set(fn)
	}
	return nil
}
