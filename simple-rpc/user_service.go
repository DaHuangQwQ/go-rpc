package simple_rpc

import (
	"context"
	"log"
)

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
