package main

import (
	"fmt"
	"github.com/Octane0411/qoou/common/qoou_jwt"
	"github.com/Octane0411/qoou/server/router"
)

func main() {
	token, _ := qoou_jwt.CreateToken("octane0411")
	fmt.Println(token)
	r := router.NewRouter()
	r.Run(":8080")
}
