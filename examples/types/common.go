package types

// 登录相关结构体
type LoginArgs struct {
	Username string `validate:""`
	Password string `validate:"required"`
	Token    string `mod:"from=query"`
}

type LoginReply struct {
	Uid   string
	Token string
}

// 用户相关结构体
type UserArgs struct {
	UserID string `validate:"required"`
	Name   string
}

type UserReply struct {
	ID   string
	Name string
	Role string
}
