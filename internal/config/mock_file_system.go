package config

// MockFileSystem is a mock implementation of the FileSystemReader interface for testing
type MockFileSystem struct {
	HomeDir           string
	StatError         error
	CreateErr         error
	ConfigFileContent string
	WriteFileErr      error
	WriteFileContent  string
}

func (m *MockFileSystem) ReadFile(name string) (string, error) {
	if m.StatError != nil {
		return "", m.StatError
	}
	return m.ConfigFileContent, nil
}

func (m *MockFileSystem) WriteFile(name string, content string) error {
	m.WriteFileContent = content
	return m.WriteFileErr
}
