package go_rpc

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	"time"
)

func TestInitClientProxy(t *testing.T) {
	server := NewServer()
	server.RegisterService(&UserServiceServer{})
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)
	usClient := &UserService{}
	err := initClientProxy(":8081", usClient)
	require.NoError(t, err)
	resp, err := usClient.GetById(context.Background(), &GetByIdReq{Id: 123})
	require.NoError(t, err)
	assert.Equal(t, &GetByIdResp{
		Msg: "hello, world",
	}, resp)
}

type UserService struct {
	// 用反射来赋值
	// 类型是函数的字段，它不是方法（它不是定义在 UserService 上的方法）
	// 本质上是一个字段
	GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)
}

func (u UserService) Name() string {
	return "user-service"
}

type GetByIdReq struct {
	Id int
}

type GetByIdResp struct {
	Msg string
}

type UserServiceServer struct {
}

func (u *UserServiceServer) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	log.Println(req)
	return &GetByIdResp{
		Msg: "hello, world",
	}, nil
}

func (u *UserServiceServer) Name() string {
	return "user-service"
}
