package controller

import (
	"encoding/json"
	"net/http"
	"path"
	"strconv"

	"gin-app-start/internal/code"
	"gin-app-start/internal/common"
	"gin-app-start/internal/config"
	"gin-app-start/internal/dto"
	"gin-app-start/internal/service"
	"gin-app-start/internal/validation"
	"gin-app-start/pkg/errors"
	"gin-app-start/pkg/utils"

	"github.com/gin-gonic/gin"
)

type userSession struct {
	UserId   uint   `json:"userId"`
	UserName string `json:"username"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
}

func getUserSession(sessionData interface{}) (userSession, error) {
	if sessionData == nil {
		return userSession{}, errors.New("Session data is nil")
	}

	// 从session中获取用户信息
	var user userSession
	if err := json.Unmarshal(sessionData.([]byte), &user); err != nil {
		return user, err
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
func (ctrl *UserController) Login() common.HandlerFunc {
	return func(c common.Context) {
		var req dto.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				validation.Error(err)).WithError(err),
			)
			return
		}

		u, err := ctrl.userService.Login(c, &req)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AdminLoginError,
				code.Text(code.AdminLoginError)).WithError(err),
			)
			return
		}

		data := gin.H{
			"userId":   u.ID,
			"username": u.Username,
			"phone":    u.Phone,
			"email":    u.Email,
			"avatar":   config.GlobalConfig.File.UrlPrefix + u.Avatar,
		}

		value, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusInternalServerError,
				code.MarshalError,
				code.Text(code.MarshalError)).WithError(err),
			)
			return
		}

		session := c.GetSession()
		session.Set(common.SESSION_KEY, value)
		session.Save()

		c.Payload(data)
	}
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
func (ctrl *UserController) CreateUser() common.HandlerFunc {
	return func(c common.Context) {
		var req dto.CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				validation.Error(err)).WithError(err),
			)
			return
		}

		user, err := ctrl.userService.CreateUser(c, &req)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AdminCreateError,
				code.Text(code.AdminCreateError)).WithError(err),
			)
			return
		}
		c.Payload(user)
	}
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
func (ctrl *UserController) ChangePassword() common.HandlerFunc {
	return func(c common.Context) {
		var req dto.UpdatePasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				validation.Error(err)).WithError(err),
			)
			return
		}

		// 校验请求中的用户名是否与session中的用户名一致
		sessionData := c.SessionUserInfo()
		user, err := getUserSession(sessionData)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(err),
			)
			return
		}

		// 校验用户名是否一致
		if user.UserName != common.ADMIN_NAME && req.Username != user.UserName {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(errors.New(user.UserName + " overstepping authority")),
			)
			return
		}

		if err = ctrl.userService.UpdatePassword(c, &req); err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AdminModifyPasswordError,
				code.Text(code.AdminModifyPasswordError)).WithError(err),
			)
			return
		}

		// 清除session，重新登录
		session := c.GetSession()
		session.Clear()
		session.Save()

		c.Payload("Change password success")
	}
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
func (ctrl *UserController) UploadImage() common.HandlerFunc {
	return func(c common.Context) {
		username := c.PostForm("username")

		sessionData := c.SessionUserInfo()
		user, err := getUserSession(sessionData)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(err),
			)
			return
		}

		if user.UserName != common.ADMIN_NAME && username != user.UserName {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(errors.New(user.UserName + " overstepping authority")),
			)
			return
		}

		file, err := c.FormFile("file")
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		_, err = ctrl.userService.GetUserByUsername(c, username)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AdminDetailError,
				code.Text(code.AdminDetailError)).WithError(err),
			)
			return
		}

		dst := path.Join(config.GlobalConfig.File.DirName, username)

		// 暂时保存文件到服务器，TODO:上传到oss、七牛云
		filename, err := utils.SaveToFile(file, dst)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.FileUploadError,
				code.Text(code.FileUploadError)).WithError(err),
			)
			return
		}

		err = ctrl.userService.UploadImage(c, username, filename)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AdminUpdateError,
				code.Text(code.AdminUpdateError)).WithError(err),
			)
			return
		}

		// 返回头像url
		avatarUrl := config.GlobalConfig.File.UrlPrefix + filename
		c.Payload(avatarUrl)
	}
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
func (ctrl *UserController) GetImage() common.HandlerFunc {
	return func(c common.Context) {
		username := c.Query("username")
		imageName := c.Query("imageName")
		if username == "" || imageName == "" {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.ParamQueryError,
				code.Text(code.ParamQueryError)).WithError(errors.New("username or imageName is empty")),
			)
			return
		}

		sessionData := c.SessionUserInfo()
		user, err := getUserSession(sessionData)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(err),
			)
			return
		}

		if user.UserName != common.ADMIN_NAME && username != user.UserName {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(errors.New(user.UserName + " overstepping authority")),
			)
			return
		}

		fileName := path.Join(config.GlobalConfig.File.DirName, username, imageName)

		c.File(fileName)
	}
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
func (ctrl *UserController) GetUser() common.HandlerFunc {
	return func(c common.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.ParseError,
				code.Text(code.ParseError)).WithError(err),
			)
			return
		}

		sessionData := c.SessionUserInfo()
		user, err := getUserSession(sessionData)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(err),
			)
			return
		}

		if user.UserName != common.ADMIN_NAME && user.UserId != uint(id) {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(errors.New(user.UserName + " overstepping authority")),
			)
			return
		}

		userData, err := ctrl.userService.GetUser(c, uint(id))
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AdminDetailError,
				code.Text(code.AdminDetailError)).WithError(err),
			)
			return
		}

		c.Payload(userData)
	}
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
func (ctrl *UserController) UpdateUser() common.HandlerFunc {
	return func(c common.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.ParseError,
				code.Text(code.ParseError)).WithError(err),
			)
			return
		}

		sessionData := c.SessionUserInfo()
		user, err := getUserSession(sessionData)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(err),
			)
			return
		}

		if user.UserName != common.ADMIN_NAME && user.UserId != uint(id) {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(errors.New(user.UserName + " overstepping authority")),
			)
			return
		}

		var req dto.UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				validation.Error(err)).WithError(err),
			)
			return
		}

		userData, err := ctrl.userService.UpdateUser(c, uint(id), &req)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AdminUpdateError,
				code.Text(code.AdminUpdateError)).WithError(err),
			)
			return
		}

		c.Payload(userData)
	}
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
func (ctrl *UserController) DeleteUser() common.HandlerFunc {
	return func(c common.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.ParseError,
				code.Text(code.ParseError)).WithError(err),
			)
			return
		}

		sessionData := c.SessionUserInfo()
		user, err := getUserSession(sessionData)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(err),
			)
			return
		}

		if user.UserName != common.ADMIN_NAME && user.UserId != uint(id) {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(errors.New(user.UserName + " overstepping authority")),
			)
			return
		}

		if err := ctrl.userService.DeleteUser(c, uint(id)); err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AdminDeleteError,
				code.Text(code.AdminDeleteError)).WithError(err),
			)
			return
		}

		c.Payload("Deleted successfully")
	}
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
func (ctrl *UserController) ListUsers() common.HandlerFunc {
	return func(c common.Context) {
		var res dto.ListUsersResponse
		sessionData := c.SessionUserInfo()
		user, err := getUserSession(sessionData)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(err),
			)
			return
		}
		if user.UserName != common.ADMIN_NAME {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(errors.New(user.UserName + " overstepping authority")),
			)
			return
		}

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

		users, total, err := ctrl.userService.ListUsers(c, page, pageSize)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AdminListError,
				code.Text(code.AdminListError)).WithError(err),
			)
			return
		}

		res.Users = users
		res.Total = total
		res.Page = page
		res.PageSize = pageSize

		c.Payload(res)
	}
}

func (ctrl *UserController) Logout() common.HandlerFunc {
	return func(c common.Context) {
		var req dto.LogoutRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				validation.Error(err)).WithError(err),
			)
			return
		}

		sessionData := c.SessionUserInfo()
		user, err := getUserSession(sessionData)
		if err != nil {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(err),
			)
			return
		}

		if user.UserName != req.Username {
			c.AbortWithError(common.Error(
				http.StatusBadRequest,
				code.AuthorizationError,
				code.Text(code.AuthorizationError)).WithError(errors.New(user.UserName + " overstepping authority")),
			)
			return
		}

		// 清除session，重新登录
		session := c.GetSession()
		session.Clear()
		session.Save()

		c.Payload("Logout successfully")
	}
}
