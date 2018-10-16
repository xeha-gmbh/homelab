package shared

type Payload struct {
	// flavor of the OS
	Flavor 		string
	// path to the input ISO
	InputIso 	string
	// path of the output files
	OutputPath 	string
	// name of the output ISO file
	OutputName 	string
	// if should be made bootable via USB
	UsbBoot		bool

	// Attributes of the new user
	Timezone 	string
	Username 	string
	Password 	string
	Hostname 	string
	Domain 		string
}
