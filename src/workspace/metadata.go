package workspace

type Metadata struct {
	path string
}

func (m *Metadata) GetRootPath() string {
	return m.path
}
