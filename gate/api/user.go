package api

import (
	"common"
	"common/biz"
	"common/config"
	"common/jwts"
	"common/logs"
	"common/rpc"
	"context"
	"framework/myError"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"time"
	"user/pb"
)

type UserHandler struct {
}

// NewUserHandler 返回值类型：*UserHandler;创建 UserHandler 的零值实例，并取其地址：&UserHandler{}
func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// Register 注册的逻辑业务
func (u *UserHandler) Register(ctx *gin.Context) {
	// 步骤1：接收并绑定参数
	var req pb.RegisterParams
	err2 := ctx.ShouldBindJSON(&req)
	if err2 != nil {
		common.Fail(ctx, biz.RequestDataError)
		return
	}
	// 步骤2：调用RPC注册服务（传递实际参数）
	response, err := rpc.UserClient.Register(context.TODO(), &req)
	if response == nil {
		common.Fail(ctx, biz.Fail)
	}
	if err != nil {
		common.Fail(ctx, myError.ToError(err))
		return
	}
	uid := response.Uid
	if len(uid) == 0 {
		common.Fail(ctx, biz.SqlError)
	}
	logs.Info("uid: %s", uid)

	//  步骤3：生成JWT令牌
	claims := jwts.CustomClaims{
		Uid: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), //7天有效
		},
	}
	token, err := jwts.GenToken(&claims, config.Conf.Jwt.Secret)
	if err != nil {
		logs.Error("❌ Register jwt gen token err:%v", err)
		common.Fail(ctx, biz.Fail)
		return
	}
	// 步骤4：返回响应
	result := map[string]any{
		//	返回一个token。token使用jwt来生成
		//JWT由三部分组成，1、头，定义加密算法2、存储数据3、签名（base64）
		"token": token,
		"serverInfo": map[string]any{
			"host": config.Conf.Services["connector"].ClientHost,
			"port": config.Conf.Services["connector"].ClientPort,
		},
	}

	common.Success(ctx, result)
}
