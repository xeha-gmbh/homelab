package shared

type Payload struct {
	// flavor of the OS
	Flavor 		string
	// path of the output files
	OutputPath 	string
	// if should be made bootable via USB
	UsbBoot		bool

	// Attributes of the new user
	Timezone 	string
	Username 	string
	Password 	string
	Hostname 	string
	Domain 		string
}
