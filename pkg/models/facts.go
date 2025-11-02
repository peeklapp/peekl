package models

type OsFacts struct {
	Arch   string
	Distro string
}

type PackageFact struct {
	Name    string
	Version string
}

type Facts struct {
	Os       OsFacts
	Packages []PackageFact
}
