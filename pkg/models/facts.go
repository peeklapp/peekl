package models

type OsFacts struct {
	Arch   string `json:"arch"`
	Distro string `json:"distro"`
}

type PackageFact struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Facts struct {
	Os       OsFacts   `json:"os"`
	Packages []Package `json:"packages"`
}
