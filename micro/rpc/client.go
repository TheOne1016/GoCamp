package rpc

import (
	"encoding/json"
	"errors"
	"net"
	"reflect"
	"time"

	"golang.org/x/net/context"
)

// InitClientProxy, 要为 GetById 之类的函数类型的字段赋值
func InitClientProxy(addr string, service Service) error {

	client := NewClient(addr)
	//在这里初始化一个Proxy

	return setFuncField(service, client)

}

func setFuncField(service Service, p Proxy) error {
	if service == nil {
		return errors.New("rpc: 不支持 nil")
	}
	val := reflect.ValueOf(service)
	typ := val.Type()
	//只支持指向结构体的一级指针
	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		return errors.New("rpc: 只支持指向结构体的一级指针")
	}

	val = val.Elem()
	typ = typ.Elem()

	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		fieldTyp := typ.Field(i)
		fieldVal := val.Field(i)

		if fieldVal.CanSet() {
			fn := func(args []reflect.Value) (results []reflect.Value) {

				retVal := reflect.New(fieldTyp.Type.Out(0).Elem())

				// args[0] 是 context
				ctx := args[0].Interface().(context.Context)
				// args[1] 是 req
				reqData, err := json.Marshal(args[1].Interface())
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}
				req := &Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Arg:         reqData,
				}

				//要真的发起调用了
				resp, err := p.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				err = json.Unmarshal(resp.Data, retVal.Interface())
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				return []reflect.Value{retVal,
					reflect.Zero(reflect.TypeOf(new(error)).Elem())}
			}

			// 我要设置值给 GetById
			fnVal := reflect.MakeFunc(fieldTyp.Type, fn)
			fieldVal.Set(fnVal)
		}
	}

	return nil
}

const numOfLengthBytes = 8

type Client struct {
	addr string
}

func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
	}
}

func (c *Client) Invoke(ctx context.Context, req *Request) (*Response, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	//正儿八经地把请求发过去服务器
	resp, err := c.Send(data)
	if err != nil {
		return nil, err
	}

	return &Response{
		Data: resp,
	}, nil
}

func (c *Client) Send(data []byte) ([]byte, error) {
	conn, err := net.DialTimeout("tcp", c.addr, time.Second*3)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = conn.Close()
	}()

	req := EncodeMsg(data)

	_, err = conn.Write(req)
	if err != nil {
		return nil, err
	}

	return ReadMsg(conn)

}
