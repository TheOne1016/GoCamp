package web

import (
	"GoCamp/webook/internal/domain"
	"GoCamp/webook/internal/service"
	"net/http"
	"time"
	"unicode/utf8"

	ijwt "GoCamp/webook/internal/web/jwt"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const biz = "login"

// ensure UserHandler implement handler interface
// var _ handler = &UserHandler{}
// or
var _ handler = (*UserHandler)(nil)

// 在UserHandler上定义跟用户有关的路由
type UserHandler struct {
	svc                                service.UserService
	emailExp, passwordExp, birthdayExp *regexp.Regexp
	codeSvc                            service.CodeService
	ijwt.Handler
	cmd redis.Cmdable
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService,
	jwtHdl ijwt.Handler) *UserHandler {
	const (
		emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		// 和上面比起来，用 ` 看起来就比较清爽
		passwordRegexPattern = `^(?=.*\d)(?=.*[^\da-zA-Z\s])(\d|[^\da-zA-Z\s]){8,}$`

		birthdayRegexPattern = `^\d{4}-\d{2}-\d{2}$`
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	birthdayExp := regexp.MustCompile(birthdayRegexPattern, regexp.None)
	return &UserHandler{
		svc:         svc,
		emailExp:    emailExp,
		passwordExp: passwordExp,
		birthdayExp: birthdayExp,
		codeSvc:     codeSvc,
		Handler:     jwtHdl,
	}

}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	//采用gin的分组路由
	ug := server.Group("/users")
	ug.POST("/signup", u.SignUp)

	ug.POST("/login", u.LoginJWT)
	ug.POST("/logout", u.LogoutJWT)

	ug.POST("/edit", u.Edit)

	ug.GET("/profile", u.ProfileJWT)

	ug.POST("/login_sms/code/send", u.SendLoginSMSCode)

	ug.POST("/login_sms", u.LoginSMS)
	ug.POST("/refresh_token", u.RefreshToken)

}

func (u *UserHandler) LogoutJWT(ctx *gin.Context) {
	err := u.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "退出登录失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "退出登录OK",
	})
}

// RefreshToken 可以同时刷新长短 token，用redis来记录是否有效，即 refresh_token是一次性的
func (u *UserHandler) RefreshToken(ctx *gin.Context) {

	//只有这个接口，拿出来的才是 refresh_token，其它的地方都是 access_token
	refreshToken := u.ExtractToken(ctx)
	var rc ijwt.RefreshClaims
	token, err := jwt.ParseWithClaims(refreshToken, &rc, func(token *jwt.Token) (interface{}, error) {
		return ijwt.RtKey, nil
	})
	if err != nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = u.CheckSession(ctx, rc.Ssid)
	if err != nil {
		//要么 redis 有问题， 要么已经退出登录
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	//搞个新的 access_token
	err = u.SetJWTToken(ctx, rc.Uid, rc.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		zap.L().Error("设置JWT Token 出现异常", zap.Error(err),
			zap.String("method", "UserHandler:RefreshToken"))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "刷新成功",
	})
}

func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	ok, err := u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		zap.L().Error("校验验证码出错", zap.Error(err))
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "验证码有误",
		})
		return
	}

	//我这个手机号，会不会是一个新用户呢
	user, err := u.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	if err = u.SetLoginToken(ctx, user.Id); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 4,
		Msg:  "验证码校验通过",
	})

}

func (u *UserHandler) SendLoginSMSCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	err := u.codeSvc.Send(ctx, biz, req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送成功",
		})
	case service.ErrCodeSendToMany:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送太频繁，请稍后再试",
		})
	default:
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}

}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		ConfirmPassword string `json:"confirmPassword"`
		Password        string `json:"password"`
	}

	var req SignUpReq
	//Bind方法会根据Content-Type来解析你的数据到 req 里面
	//解析错了，就会直接写回一个 400 的错误
	if err := ctx.Bind(&req); err != nil {
		return
	}

	//校验邮箱
	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "你的邮箱格式不对")
		return
	}

	//校验密码
	if req.ConfirmPassword != req.Password {
		ctx.String(http.StatusOK, "两次输入的密码不一致")
		return
	}

	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		//记录日志
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "密码必须大于8位，包含数字、特殊字符")
		return
	}

	//调用一下svc的方法
	err = u.svc.SignUp(ctx, domain.User{Email: req.Email, Password: req.Password})

	if err == service.ErrUserDuplicate {
		ctx.String(http.StatusOK, "邮箱或手机号码冲突")
		return
	}

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "注册成功")

}

func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "用户名或密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	//在这里用JWT设置登录态
	//生成一个JWTtoken

	err = u.SetLoginToken(ctx, user.Id)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "登录成功")
	return
}

func (u *UserHandler) setJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := ijwt.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
		Uid:  uid,
		Ssid: ssid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("Pe9um9NrZbtVbGzmjIaoMXa4WbY00iuy"))
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "用户名或密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	//在这里登陆成功了，设置session
	sess := sessions.Default(ctx)
	//可以随便设置放在session里面的值了
	sess.Set("userId", user.Id)

	sess.Options(sessions.Options{
		//Secure:   true,
		//HttpOnly: true,
		MaxAge: 30 * 60,
	})
	sess.Save()
	ctx.String(http.StatusOK, "登录成功")
	return
}

func (u *UserHandler) Edit(ctx *gin.Context) {
	type ExtraReq struct {
		NickName  string `json:"nickname"`
		Birthday  string `json:"birthday"`
		BirefInfo string `json:"birefInfo"`
	}
	var req ExtraReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	//校验出生日期
	ok, err := u.birthdayExp.MatchString(req.Birthday)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "你的出生日期格式不对，应该为 xxxx-xx-xx")
		return
	}

	//校验昵称长度
	if utf8.RuneCountInString(req.NickName) > 12 {
		ctx.String(http.StatusOK, "昵称不能超过12个字符")
		return
	}

	//校验简介长度
	if utf8.RuneCountInString(req.BirefInfo) > 100 {
		ctx.String(http.StatusOK, "简介应不能超过100个字符")
		return
	}

	sess := sessions.Default(ctx)
	id := sess.Get("userId")
	if id == nil {
		//没有登陆
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	Id := id.(int64)
	err = u.svc.Edit(ctx, Id, req.NickName, req.Birthday, req.BirefInfo)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "编辑成功")
	return

}

func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	c, ok := ctx.Get("claims")
	//你可以断定，必然有 claims
	// if !ok {
	// 	ctx.String(http.StatusOK, "系统错误")
	// 	return
	// }
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		//你可以考虑监控住这里
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	Id := claims.Uid
	user, err := u.svc.GetProfile(ctx, Id)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"邮箱":   user.Email,
		"昵称":   user.NickName,
		"出生年月": user.Birthday,
		"简介":   user.BirefInfo})
	//ctx.String(http.StatusOK, "这是你的Profile")

}

func (u *UserHandler) Profile(ctx *gin.Context) {
	//sess := sessions.Default(ctx)
	//id := sess.Get("userId")
	//if id == nil {
	//	//没有登陆
	//	ctx.AbortWithStatus(http.StatusUnauthorized)
	//	return
	//}
	//Id := id.(int64)
	//user, err := u.svc.GetProfile(ctx, Id)
	//if err != nil {
	//	ctx.String(http.StatusOK, "系统错误")
	//	return
	//}

	//ctx.JSON(http.StatusOK, gin.H{
	//	"邮箱":   user.Email,
	//	"昵称":   user.NickName,
	//	"出生年月": user.Birthday,
	//	"简介":   user.BirefInfo})

}
