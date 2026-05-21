package firmware

// latest is the hardcoded "latest available firmware" per device type.
// There is no real OTA — this map is the only source of truth. The keys
// must stay in sync with the device_type enum on devices.device_type
// (validator tag in devices.CreateRequest).
var latest = map[string]string{
	"vespin_classic": "1.0.2",
	"vespin_mini":    "1.0.2",
	"vespin_pro":     "1.1.0",
}
