package rainbird

type Device struct {
	ip       string
	pass     string
	msgid    int
	useCache bool
	cache    map[string][]byte
}

// Initialize device instance
func Get(ip string, pass string) *Device {
	dev := new(Device)
	dev.ip = ip
	dev.pass = pass
	dev.msgid = 1
	dev.useCache = false
	return dev
}

/*
Enable in memory caching of user-set commands.

Cache the current schedules for a zone, raindelay and modelandversion,
cache will automatically update if new values are set through this api.

To manually clear (and therefore refresh) the cache, simply call again with use=true
*/
func (rb *Device) UseCaching(use bool) {
	rb.useCache = use
	rb.cache = map[string][]byte{}
}
