package rainbird

import (
	"log"
	"os"
	"runtime/debug"
)

type device struct {
	ip    string
	pass  string
	msgid int
}

func init() {
	//https://www.jajaldoang.com/post/go-write-log-to-file/
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	BI, _ := debug.ReadBuildInfo()
	log.Println("rainbird module imported by", "\""+BI.Path+"\"", "go version", BI.GoVersion)
}
func LogToFile() {
	logFile, err := os.OpenFile("rainbird.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Println("Couldn't open log file:", err)
	} else {
		log.SetOutput(logFile)
	}
}

func Get(ip string, pass string) *device {
	dev := new(device)
	dev.ip = ip
	dev.pass = pass
	dev.msgid = 1
	return dev
}
