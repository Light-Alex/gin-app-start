package controller

import (
	"encoding/json"
	"path"
	"strconv"

	"gin-app-start/internal/common"
	"gin-app-start/internal/config"
	"gin-app-start/internal/dto"
	"gin-app-start/internal/service"
	"gin-app-start/pkg/errors"
	"gin-app-start/pkg/logger"
	"gin-app-start/pkg/response"
	"gin-app-start/pkg/utils"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type userSession struct {
	UserId   uint   `json:"userId"`
	UserName string `json:"username"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
}

func getUserSession(sessionData interface{}) (userSession, error) {
	// 从session中获取用户信息
	var user userSession
	if err := json.Unmarshal(sessionData.([]byte), &user); err != nil {
		logger.Error("Unmarshal user session data error:", zap.Error(err))
		return user, errors.WrapBusinessError(10037, "Unmarshal user session data error", err)
	}

	return user, nil
}

type UserController struct {
	userService service.UserService
}

func NewUserController(userService service.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// Login godoc
//
//	@Summary		Login user
//	@Description	Login user with username and password
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LoginRequest	true	"User login information"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/api/v1/users/login [post]
func (ctrl *UserController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Parameter binding failed", zap.Error(err))
		response.Error(c, 10001, "Parameter binding failed: "+err.Error())
		return
	}

	u, err := ctrl.userService.Login(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Login failed: ", zap.Error(err))
		response.Error(c, 10035, "Login failed: "+err.Error())
		return
	}

	data := gin.H{
		"userId":   u.ID,
		"username": u.Username,
		"phone":    u.Phone,
		"email":    u.Email,
		"avatar":   config.GlobalConfig.File.UrlPrefix + u.Avatar,
	}

	value, err := json.Marshal(data)
	if err != nil {
		logger.Error("Marshal data failed: ", zap.Error(err))
		response.Error(c, 10035, "Login failed: "+err.Error())
		return
	}

	session := sessions.Default(c)
	session.Set(common.SESSION_KEY, value)
	session.Save()

	response.Success(c, data)
}

// CreateUser godoc
//
//	@Summary		Create a new user
//	@Description	Create a new user with username, email, phone and password
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.CreateUserRequest	true	"User information"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/api/v1/users [post]
func (ctrl *UserController) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Parameter binding failed", zap.Error(err))
		response.Error(c, 10001, "Parameter binding failed: "+err.Error())
		return
	}

	user, err := ctrl.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Create user failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	response.Success(c, user)
}

// CreateUser godoc
//
//	@Summary		Change a user's password
//	@Description	Change a user's password with old password and new password
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.UpdatePasswordRequest	true	"User update password information"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/api/v1/users/change_pwd [post]
func (ctrl *UserController) ChangePassword(c *gin.Context) {
	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Parameter binding failed", zap.Error(err))
		response.Error(c, 10001, "Parameter binding failed: "+err.Error())
		return
	}

	// 校验请求中的用户名是否与session中的用户名一致
	session := sessions.Default(c)
	sessionData := session.Get(common.SESSION_KEY)
	user, err := getUserSession(sessionData)
	if err != nil {
		logger.Error("Get user session failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	// 校验用户名是否一致
	if user.UserName != common.ADMIN_NAME && req.Username != user.UserName {
		logger.Error("User %s can only change own password", zap.String("username", user.UserName))
		response.Error(c, 10036, "overstepping authority")
		return
	}

	if err = ctrl.userService.UpdatePassword(c.Request.Context(), &req); err != nil {
		logger.Error("Change password failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	session.Clear()
	session.Save()

	response.Success(c, nil)
}

// CreateUser godoc
//
//	@Summary		Upload Avatar Image
//	@Description	Upload avatar image for user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	    formData	file 	true  	"User avatar image"
//	@Param			username	formData	string	true	"username"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/api/v1/users/upload_avatar [post]
func (ctrl *UserController) UploadImage(c *gin.Context) {
	username := c.PostForm("username")
	sessionData, _ := c.Get(common.SESSION_KEY)
	user, err := getUserSession(sessionData)
	if err != nil {
		logger.Error("Get user session failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	if user.UserName != common.ADMIN_NAME && username != user.UserName {
		logger.Error("User can only upload image for own", zap.String("username", user.UserName))
		response.Error(c, 10036, "overstepping authority")
		return
	}

	file, err := c.FormFile("name")
	if err != nil {
		logger.Error("Get form file failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	_, err = ctrl.userService.GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		logger.Error("Get user by username failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	dst := path.Join(config.GlobalConfig.File.DirName, username)

	// 暂时保存文件到服务器，TODO:上传到oss、七牛云
	filename, err := utils.SaveToFile(file, dst)
	if err != nil {
		logger.Error("save to file error:", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	err = ctrl.userService.UploadImage(c.Request.Context(), username, filename)
	if err != nil {
		logger.Error("Upload image failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	// 返回头像url
	avatarUrl := config.GlobalConfig.File.UrlPrefix + filename
	response.Success(c, avatarUrl)
}

// GetImage godoc
//
//	@Summary		Get user image by username and image name
//	@Description	Get user image by username and image name
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			username	query		string	true	"username"
//	@Param			imageName	query		string	true	"image name"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/api/v1/users/file?username={username}&imageName={imageName} [get]
func (ctrl *UserController) GetImage(c *gin.Context) {
	username := c.Query("username")
	imageName := c.Query("imageName")
	if username == "" || imageName == "" {
		logger.Error("Parameter binding failed")
		response.Error(c, 10001, "Parameter binding failed")
		return
	}

	sessionData, _ := c.Get(common.SESSION_KEY)
	user, err := getUserSession(sessionData)
	if err != nil {
		logger.Error("Get user session failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	if user.UserName != common.ADMIN_NAME && username != user.UserName {
		logger.Error("User can only get image for own", zap.String("username", user.UserName))
		response.Error(c, 10036, "overstepping authority")
		return
	}

	fileName := path.Join(config.GlobalConfig.File.DirName, username, imageName)
	c.File(fileName)
}

// GetUser godoc
//
//	@Summary		Get user by ID
//	@Description	Get user information by user ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/api/v1/users/{id} [get]
func (ctrl *UserController) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.Error("Invalid user ID", zap.Error(err))
		response.Error(c, 10001, "Invalid user ID")
		return
	}

	sessionData, _ := c.Get(common.SESSION_KEY)
	user, err := getUserSession(sessionData)
	if err != nil {
		logger.Error("Get user session failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}
	if user.UserName != common.ADMIN_NAME && user.UserId != uint(id) {
		logger.Error("User can only get user info for own", zap.String("username", user.UserName), zap.Uint("session id", user.UserId), zap.Uint("request id", uint(id)))
		response.Error(c, 10036, "overstepping authority")
		return
	}

	userData, err := ctrl.userService.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Get user failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	response.Success(c, userData)
}

// UpdateUser godoc
//
//	@Summary		Update user information
//	@Description	Update user information by user ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"User ID"
//	@Param			request	body		dto.UpdateUserRequest	true	"User information to update"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		404		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/api/v1/users/{id} [put]
func (ctrl *UserController) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.Error("Invalid user ID", zap.Error(err))
		response.Error(c, 10001, "Invalid user ID")
		return
	}

	sessionData, _ := c.Get(common.SESSION_KEY)
	user, err := getUserSession(sessionData)
	if err != nil {
		logger.Error("Get user session failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}
	if user.UserName != common.ADMIN_NAME && user.UserId != uint(id) {
		logger.Error("User can only get user info for own", zap.String("username", user.UserName))
		response.Error(c, 10036, "overstepping authority")
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Parameter binding failed", zap.Error(err))
		response.Error(c, 10001, "Parameter binding failed: "+err.Error())
		return
	}

	userData, err := ctrl.userService.UpdateUser(c.Request.Context(), uint(id), &req)
	if err != nil {
		logger.Error("Update user failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	response.Success(c, userData)
}

// DeleteUser godoc
//
//	@Summary		Delete user
//	@Description	Delete user by user ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/api/v1/users/{id} [delete]
func (ctrl *UserController) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.Error("Invalid user ID", zap.Error(err))
		response.Error(c, 10001, "Invalid user ID")
		return
	}

	sessionData, _ := c.Get(common.SESSION_KEY)
	user, err := getUserSession(sessionData)
	if err != nil {
		logger.Error("Get user session failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}
	if user.UserName != common.ADMIN_NAME && user.UserId != uint(id) {
		logger.Error("User can only get user info for own", zap.String("username", user.UserName))
		response.Error(c, 10036, "overstepping authority")
		return
	}

	if err := ctrl.userService.DeleteUser(c.Request.Context(), uint(id)); err != nil {
		logger.Error("Delete user failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Deleted successfully", nil)
}

// ListUsers godoc
//
//	@Summary		List users
//	@Description	Get paginated list of users
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int	false	"Page number"		default(1)
//	@Param			page_size	query		int	false	"Page size"			default(10)
//	@Success		200			{object}	response.Response
//	@Failure		500			{object}	response.Response
//	@Router			/api/v1/users [get]
func (ctrl *UserController) ListUsers(c *gin.Context) {
	sessionData, _ := c.Get(common.SESSION_KEY)
	user, err := getUserSession(sessionData)
	if err != nil {
		logger.Error("Get user session failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}
	if user.UserName != common.ADMIN_NAME {
		logger.Error("Only admin can get users list")
		response.Error(c, 10036, "overstepping authority")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	users, total, err := ctrl.userService.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		logger.Error("List users failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	response.SuccessWithPage(c, users, total, page, pageSize)
}

func (ctrl *UserController) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Parameter binding failed", zap.Error(err))
		response.Error(c, 10001, "Parameter binding failed: "+err.Error())
		return
	}

	session := sessions.Default(c)
	sessionData := session.Get(common.SESSION_KEY)
	user, err := getUserSession(sessionData)
	if err != nil {
		logger.Error("Get user session failed", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	if user.UserName != req.Username {
		logger.Error("User not authorized to logout", zap.String("username", req.Username))
		response.Error(c, 10036, "overstepping authority")
		return
	}

	session.Clear()
	session.Save()

	response.SuccessWithMessage(c, "Logout successfully", nil)
}

func handleServiceError(c *gin.Context, err error) {
	var bizErr *errors.BusinessError
	if e, ok := err.(*errors.BusinessError); ok {
		bizErr = e
		response.Error(c, bizErr.Code, bizErr.Message)
	} else {
		logger.Error("Unknown error", zap.Error(err))
		response.Error(c, 50000, "Internal server error")
	}
}
