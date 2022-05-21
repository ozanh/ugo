package importers

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/ozanh/ugo"
)

// FileImporter is an implemention of ugo.ExtImporter to import files from file
// system. It uses absolute paths of module as import names.
type FileImporter struct {
	WorkDir    string
	FileReader func(string) ([]byte, error)
	name       string
}

var _ ugo.ExtImporter = (*FileImporter)(nil)

// Get impelements ugo.ExtImporter and returns itself if name is not empty.
func (m *FileImporter) Get(name string) ugo.ExtImporter {
	if name == "" {
		return nil
	}
	m.name = name
	return m
}

// Name returns the absoule path of the module. A previous Get call is required
// to get the name of the imported module.
func (m *FileImporter) Name() string {
	if m.name == "" {
		return ""
	}
	path := m.name
	if !filepath.IsAbs(path) {
		path = filepath.Join(m.WorkDir, path)
		if p, err := filepath.Abs(path); err == nil {
			path = p
		}
	}
	return path
}

// Import returns the content of the path determined by Name call. Empty name
// will return an error.
func (m *FileImporter) Import(moduleName string) (interface{}, error) {
	// Note that; moduleName == Name()
	if m.name == "" || moduleName == "" {
		return nil, errors.New("invalid import call")
	}
	if m.FileReader == nil {
		return ioutil.ReadFile(moduleName)
	}
	return m.FileReader(moduleName)
}

// Fork returns a new instance of FileImporter as ugo.ExtImporter by capturing
// the working directory of the module. moduleName should be the same value
// provided by Name call.
func (m *FileImporter) Fork(moduleName string) ugo.ExtImporter {
	// Note that; moduleName == Name()
	return &FileImporter{
		WorkDir:    filepath.Dir(moduleName),
		FileReader: m.FileReader,
	}
}

// ShebangReadFile reads given path and returns the content of the file. If file
// starts with Shebang #! , it is replaced with //.
// This function can be used as ReadFile callback in FileImporter.
func ShebangReadFile(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err == nil {
		Shebang2Slashes(data)
	}
	return data, err
}

// Shebang2Slashes replaces first two bytes of given p with two slashes if they
// are Shebang chars.
func Shebang2Slashes(p []byte) {
	if len(p) > 1 && string(p[:2]) == "#!" {
		copy(p, "//")
	}
}
