package example

import (
	"github.com/dlshle/aghs/server"
)

func main() {
	httpServer := server.New()
	err := httpServer.RegisterService(NewStudentService())
	if err != nil {
		panic(err)
	}
	httpServer.Start("0.0.0.0:1234")
}
