package cdglobal

// Shortkeyer defines the interface to keyboard shortkeys.
//
type Shortkeyer interface {
	// ConfFilePath returns the shortkey conf file path.
	//
	ConfFilePath() string

	// Demander returns the shortkey Demander.
	//
	Demander() string

	// Description returns the shortkey description.
	//
	Description() string

	// GroupName returns the shortkey group name.
	//
	GroupName() string

	// IconFilePath returns the shortkey icon file path.
	//
	IconFilePath() string

	// KeyName returns the config key name.
	//
	KeyName() string

	// KeyString returns the shortkey key reference as string.
	//
	KeyString() string

	// Rebind rebinds the shortkey.
	//
	Rebind(keystring, description string) bool

	// Success returns the shortkey success.
	//
	Success() bool
}

// Crypto defines a way to decrypt and encrypt strings.
//
type Crypto interface {
	DecryptString(str string) string
	EncryptString(str string) string
}
