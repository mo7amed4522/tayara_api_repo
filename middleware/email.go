package middleware

import (
	"bytes"
	"io"
	"strconv"
	"time"

	"github.com/mo7amed4522/tayara_api/plugin/email/utils"
	utils2 "github.com/mo7amed4522/tayara_api/utils"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/system"
	"github.com/mo7amed4522/tayara_api/service"
	"go.uber.org/zap"
)

var userService = service.ServiceGroupApp.SystemServiceGroup.UserService

func ErrorToEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		var username string
		claims, _ := utils2.GetClaims(c)
		if claims.Username != "" {
			username = claims.Username
		} else {
			id, _ := strconv.Atoi(c.Request.Header.Get("x-user-id"))
			user, err := userService.FindUserById(id)
			if err != nil {
				username = "Unknown"
			}
			username = user.Username
		}
		body, _ := io.ReadAll(c.Request.Body)
		// Rewrite the request body body back to the request body, ioutil.ReadAll will clear the data in c.Request.Body
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		record := system.SysOperationRecord{
			Ip:     c.ClientIP(),
			Method: c.Request.Method,
			Path:   c.Request.URL.Path,
			Agent:  c.Request.UserAgent(),
			Body:   string(body),
		}
		now := time.Now()

		c.Next()

		latency := time.Since(now)
		status := c.Writer.Status()
		record.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
		str := "The received request is" + record.Body + "\n" + "Request method is" + record.Method + "\n" + "Error information is as follows" + record.ErrorMessage + "\n" + "Time" + latency.String() + "\n"
		if status != 200 {
			subject := username + "" + record.Ip + "called" + record.Path + "error"
			if err := utils.ErrorToEmail(subject, str); err != nil {
				global.GVA_LOG.Error("ErrorToEmail Failed, err:", zap.Error(err))
			}
		}
	}
}
