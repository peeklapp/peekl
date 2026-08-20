package responses

type FileResponseEntry struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

type GetCatalog struct {
	NodeTarball FileResponseEntry `json:"node_tarball"`
	CodeTarball FileResponseEntry `json:"code_tarball"`
}
