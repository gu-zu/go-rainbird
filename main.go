package rainbird

type Device struct {
	ip    string
	pass  string
	msgid int
}

// Initialize device instance
func Get(ip string, pass string) *Device {
	dev := new(Device)
	dev.ip = ip
	dev.pass = pass
	dev.msgid = 1
	return dev
}
