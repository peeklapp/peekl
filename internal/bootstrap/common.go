package bootstrap

type BootstrapState int

const (
	BootstrapNone BootstrapState = iota
	BootstrapPendingCert
	BootstrapComplete
)
