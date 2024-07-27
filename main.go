package rainbird

import (
	"log"
	"os"
	"runtime/debug"
)

type Device struct {
	ip    string
	pass  string
	msgid int
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	BI, _ := debug.ReadBuildInfo()
	log.Println("rainbird module imported by", "\""+BI.Path+"\"", "go version", BI.GoVersion)
}

// For testing
func logToFile() {
	logFile, err := os.OpenFile("rainbird.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Println("Couldn't open log file:", err)
	} else {
		log.SetOutput(logFile)
	}
}

// Initialize device instance
func Get(ip string, pass string) *Device {
	dev := new(Device)
	dev.ip = ip
	dev.pass = pass
	dev.msgid = 1
	return dev
}
