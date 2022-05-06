package middleware

import (
	"github.com/Octane0411/qoou/common/qoou_jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"time"
)

func JWT() func(c *gin.Context) {
	return func(c *gin.Context) {
		tokenString := c.Request.Header.Get("Authorization")
		if tokenString == "" {
			c.JSON(401, gin.H{
				"message": "请求头中未包含token",
			})
			c.Abort()
			return
		}
		token, err := jwt.ParseWithClaims(tokenString, &qoou_jwt.QoouCliams{}, func(token *jwt.Token) (interface{}, error) {
			// since we only use the one private key to sign the tokens,
			// we also only use its public counter part to verify
			return qoou_jwt.VerifyKey, nil
		})
		if err != nil {
			c.JSON(401, gin.H{
				"message": "token验证失败",
			})
			c.Abort()
			return
		}
		claims := token.Claims.(*qoou_jwt.QoouCliams)
		if time.Now().After(claims.ExpirationTime) {
			c.JSON(401, gin.H{
				"message": "token已过期",
			})
			c.Abort()
			return
		}
		c.Set("username", claims.Username)
	}
}
