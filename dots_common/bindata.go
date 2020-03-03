// Code generated by go-bindata.
// sources:
// schemas/create_identifiers.json
// schemas/install_filtering_rule.json
// schemas/mitigation_request.json
// schemas/session_configuration.json
// schemas/test.json
// DO NOT EDIT!

package dots_common

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _schemasCreate_identifiersJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xc4\x91\xc1\x4a\xfc\x40\x0c\xc6\xef\x7d\x8a\x90\xff\xff\xb8\x63\x3d\x78\xea\x59\x04\x2f\x82\x82\x0f\x30\x76\xd3\x6e\x96\x76\x66\x48\x23\xba\xc8\xbc\xbb\x8c\xc5\x75\x2a\x53\x45\x90\xf5\xd2\xc2\x97\xef\xcb\x2f\x99\xbc\x54\x00\xf8\x7f\x6a\x77\x34\x5a\x6c\x00\x77\xaa\xa1\xa9\xeb\xfd\xe4\x9d\x99\xd5\x33\x2f\x7d\xbd\x15\xdb\xa9\x39\xbf\xa8\x67\xed\x1f\x6e\x52\x4e\x0f\x81\x52\xc8\x3f\xec\xa9\xd5\x59\x0b\xe2\x03\x89\x32\x4d\xd8\x40\xea\x0e\x80\xbc\x25\xa7\xdc\x31\xc9\x51\x03\x40\x3b\xb0\x4d\x26\xb4\x22\xf6\xf0\x96\x9e\xdd\x4a\xe3\x94\x19\xdf\xad\xc6\xd9\x91\x16\x7a\x36\xc2\xa4\xc2\xae\xc7\x63\x29\x6e\x3e\xd2\x1c\xd6\x52\x4b\xf2\x0a\xfd\x4b\x0e\x40\x2c\x32\x83\x50\xc7\xcf\x7f\xc0\xf5\xa2\x46\xac\xeb\x57\x5f\xea\xe7\xec\xec\xbc\x59\xb5\x70\xe8\xac\x3a\xf8\x27\x12\x93\xa6\x49\x1d\xd8\x29\xf5\x24\x9f\x5a\x00\xe0\x63\x08\x05\xdb\xc2\x15\xbf\xdd\x59\xc5\x76\x1d\xb7\x26\x88\x57\xdf\xfa\xe1\xf7\x36\x2f\x0c\x54\x1e\xe1\xea\xf6\xf2\xe6\xf4\xc7\xbe\xbf\xbb\x3e\x11\xb4\xca\xff\xe9\x1b\xab\xf8\x1a\x00\x00\xff\xff\xcb\xcb\xee\x3a\x3b\x04\x00\x00")

func schemasCreate_identifiersJsonBytes() ([]byte, error) {
	return bindataRead(
		_schemasCreate_identifiersJson,
		"schemas/create_identifiers.json",
	)
}

func schemasCreate_identifiersJson() (*asset, error) {
	bytes, err := schemasCreate_identifiersJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "schemas/create_identifiers.json", size: 1083, mode: os.FileMode(436), modTime: time.Unix(1519989957, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _schemasInstall_filtering_ruleJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x52\x41\x4e\xc3\x30\x10\xbc\xf7\x15\x96\xe1\x58\x13\x0e\x3d\xe5\x37\x8b\xb3\x34\x2e\x89\x1d\xad\xb7\xa0\x08\xe5\xef\xc8\x31\x04\x27\x76\x73\x40\x42\xea\x25\x59\xcd\xce\x8c\x66\xb4\xfe\x3c\x08\x21\x1f\xbd\x6e\xb1\x07\x59\x0b\xd9\x32\x0f\x75\x55\x5d\xbc\xb3\x2a\xa2\x4f\x8e\xce\x55\x43\xf0\xca\xea\xf9\x54\x45\xec\x41\x1e\x83\x8e\xc7\x01\x83\xc8\xbd\x5c\x50\x73\xc4\x06\x72\x03\x12\x1b\xf4\xb2\x16\xc1\x5d\x08\x09\x5a\xa3\xf7\xaa\x33\x9e\x7f\xd1\x19\xef\x82\x1c\x88\x60\x9c\xd5\x33\x6a\x18\xfb\x94\x16\x89\xca\x42\x8f\x2b\x34\x09\xe0\x99\x8c\x3d\xcb\x65\x35\x1d\xd7\xda\x6f\xde\x1f\xb4\x4b\x70\x85\x96\x29\x6d\xb5\x50\x30\xef\x70\xa3\xc7\x0c\xd3\xb5\xc3\x52\x99\xdd\x50\x9b\x60\x33\xb7\x07\xd6\x6d\x16\x68\x65\x93\x1c\x66\xb5\x2f\x1c\x69\xb5\xf7\xee\x4a\x1a\x95\x19\xde\x4f\xca\x22\x7f\x38\x7a\x4b\x52\x6d\xed\x84\x90\x0d\x7a\x36\x16\xd8\x38\x7b\x4b\xb5\x11\x4d\xbb\xdd\x40\x07\xab\x7f\xe9\xd6\xa0\x1d\x8b\x9b\xc4\x3b\xbf\xe6\x42\x29\x5f\x35\x73\x28\x97\xce\x8b\x17\xca\xc7\x16\x48\xbd\xe1\xfb\xcf\x49\xc0\xa8\x3a\x73\x7f\x59\xf7\x9f\xdb\xa1\x34\xff\x4c\xf1\x1f\xbe\xd3\x61\xfa\x0a\x00\x00\xff\xff\x40\x61\xd1\x74\x21\x05\x00\x00")

func schemasInstall_filtering_ruleJsonBytes() ([]byte, error) {
	return bindataRead(
		_schemasInstall_filtering_ruleJson,
		"schemas/install_filtering_rule.json",
	)
}

func schemasInstall_filtering_ruleJson() (*asset, error) {
	bytes, err := schemasInstall_filtering_ruleJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "schemas/install_filtering_rule.json", size: 1313, mode: os.FileMode(436), modTime: time.Unix(1519989957, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _schemasMitigation_requestJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x94\xcf\x4a\x03\x31\x10\xc6\xef\x7d\x8a\x10\x3d\x76\x5d\x0f\x9e\x7a\x16\xc1\x8b\xa0\xe0\x03\xc4\xdd\xd9\x74\xca\x6e\x12\x66\x47\xb4\x48\xde\x5d\x76\xb7\x2d\x49\xba\x7f\x10\x54\x7a\x29\x25\xdf\xe4\xcb\x6f\x26\x5f\xf6\x6b\x25\x84\xbc\x6e\x8b\x2d\x34\x4a\x6e\x84\xdc\x32\xbb\x4d\x9e\xef\x5a\x6b\xb2\x61\xf5\xc6\x92\xce\x4b\x52\x15\x67\xb7\x77\xf9\xb0\x76\x25\xd7\xdd\x3e\xde\x3b\xe8\x36\xd9\xb7\x1d\x14\x3c\xac\x39\xb2\x0e\x88\x11\x5a\xb9\x11\x9d\xbb\x10\xb2\x41\x46\xad\x18\x7b\x4f\xdb\xef\x19\x14\x21\x64\x51\x23\x18\xce\xb0\x04\xc3\x58\x21\x50\x20\x06\x47\x28\x22\xb5\xef\x4f\x38\x08\xc8\xd0\xb4\x51\x6d\x50\xdd\x32\xa1\xd1\xf2\x24\xf9\xc3\x3f\x7f\x34\x90\x29\xc7\x8f\x8f\x0a\x7a\xc2\x32\x11\x03\x33\x34\x0c\x1a\x48\x06\xaa\x5f\x47\xc8\x8a\x34\x70\x86\x6e\xda\x23\x05\x9a\x84\x9a\x9d\x41\x38\x87\x49\x0c\x47\x50\xe1\xe7\x65\xa0\x58\xe2\x8c\x94\xd1\xf0\xdb\x38\x41\x5c\x23\x7d\x24\xba\x91\x5e\xdb\x0f\xa0\x9e\x2b\xbc\xd9\xd4\x46\x08\xf9\xee\xdc\x48\x61\x52\xe7\xe3\x71\x24\x4d\x9c\xee\xc3\xb2\x2d\x6c\x3d\xd7\xce\xd8\x10\x66\xc6\x30\x1f\xce\x65\xb2\x87\xe7\xfb\xa7\xbf\xc1\x19\x8b\xc9\x12\xcd\xeb\xcb\xe3\xe5\xc0\xa8\x1a\xd5\x6c\xf2\xfe\x17\xa7\xc6\x0a\x18\x9b\xf3\xf7\xb3\x14\x81\xe8\x6d\x9e\x7f\x44\x57\xc7\x5f\xbf\xf2\xdf\x01\x00\x00\xff\xff\xea\x22\x97\x60\x40\x06\x00\x00")

func schemasMitigation_requestJsonBytes() ([]byte, error) {
	return bindataRead(
		_schemasMitigation_requestJson,
		"schemas/mitigation_request.json",
	)
}

func schemasMitigation_requestJson() (*asset, error) {
	bytes, err := schemasMitigation_requestJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "schemas/mitigation_request.json", size: 1600, mode: os.FileMode(436), modTime: time.Unix(1520589908, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _schemasSession_configurationJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xb4\x94\xc1\x92\x9b\x30\x0c\x86\xef\x3c\x05\xc3\xee\x71\x5d\x7a\xe8\x89\xeb\xf6\xd2\x73\x9f\x40\x80\x30\xda\xda\x32\x23\x8b\x6d\x3b\x1d\xde\xbd\xe3\x50\x32\x24\x4d\x4a\xd2\x21\x3e\x81\xec\x5f\x9f\x7e\x81\xfc\x2b\xcb\xf3\xe2\x39\x36\x3d\x7a\x28\xaa\xbc\xe8\x55\x87\xaa\x2c\xdf\x62\x60\x33\x47\x3f\x04\xb1\x65\x2b\xd0\xa9\xf9\xf8\xa9\x9c\x63\x4f\xc5\x4b\xd2\xe9\xcf\x01\x93\x28\xd4\x6f\xd8\xe8\x1c\x1b\x24\x0c\x28\x4a\x18\x8b\x2a\x4f\xd9\xd3\x2a\x08\xb5\x33\x6d\xd0\x68\x22\x59\x06\x67\x9a\x1e\x98\xd1\x55\xcb\x6b\xe0\x8e\xec\x4a\x71\x50\x3d\x0b\x76\x29\xff\x53\xd9\x62\x47\x4c\x4a\x81\x63\xf9\xf5\x20\x79\x9d\x15\x7f\x8e\x4f\x59\x9e\x4f\x07\xfe\xea\xe4\xba\x80\x13\xd1\x19\xe6\x82\x8d\xe3\xde\x45\x3b\xc7\x5d\x4f\x4a\x16\x94\xd8\x5e\x76\xb0\xe1\xe4\xd4\xc3\xb2\xa6\x97\x73\x0c\xb5\x0e\xf7\x04\x64\x7f\x3f\x1f\xa1\xc5\xde\x3d\x8a\x18\x23\x05\x36\xd4\x5e\xab\x7d\xc9\x4d\xac\x68\x51\xb6\xfb\xd1\x23\x88\xd6\x08\x6a\x92\x44\xde\xc1\xdd\xd7\x96\x2f\xac\xaf\xa3\x08\xb2\x6e\xb3\x3c\xc5\x98\xbe\x6f\x5f\x1b\x70\x2e\x7c\xc7\xab\x36\x76\x60\xc1\x0f\x23\xa8\x02\x1c\x3d\xe9\xe3\x38\xd0\x7c\x33\x4a\x1e\xc3\xf8\x60\x88\x00\xb7\xc1\x9b\x0e\x1a\x0d\x72\x1f\xea\x33\x36\xe4\xc1\xdd\x8c\x53\x21\x6b\x51\xcc\x32\x92\x81\xb7\x7e\xb7\x3a\x04\x87\xc0\xf7\x4d\xc7\xaa\x03\xbb\x4d\x48\x33\x27\x34\xef\xe0\x46\xfc\xdf\x21\xf9\x67\xd5\x67\xcd\x7c\x4c\xe5\xa6\x9d\x29\x5b\x0e\x78\xf4\xf5\x8d\x06\xd2\xad\x9e\x4d\xbf\x03\x00\x00\xff\xff\x8f\x53\xba\xfb\xa4\x06\x00\x00")

func schemasSession_configurationJsonBytes() ([]byte, error) {
	return bindataRead(
		_schemasSession_configurationJson,
		"schemas/session_configuration.json",
	)
}

func schemasSession_configurationJson() (*asset, error) {
	bytes, err := schemasSession_configurationJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "schemas/session_configuration.json", size: 1700, mode: os.FileMode(436), modTime: time.Unix(1521106846, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _schemasTelemetrySetupConfiguration_requestJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x94\xcf\x4a\x03\x31\x10\xc6\xef\x7d\x8a\x10\x3d\x76\x5d\x0f\x9e\x7a\x16\xc1\x8b\xa0\xe0\x03\xc4\xdd\xd9\x74\xca\x6e\x12\x66\x47\xb4\x48\xde\x5d\x76\xb7\x2d\x49\xba\x7f\x10\x54\x7a\x29\x25\xdf\xe4\xcb\x6f\x26\x5f\xf6\x6b\x25\x84\xbc\x6e\x8b\x2d\x34\x4a\x6e\x84\xdc\x32\xbb\x4d\x9e\xef\x5a\x6b\xb2\x61\xf5\xc6\x92\xce\x4b\x52\x15\x67\xb7\x77\xf9\xb0\x76\x25\xd7\xdd\x3e\xde\x3b\xe8\x36\xd9\xb7\x1d\x14\x3c\xac\x39\xb2\x0e\x88\x11\x5a\xb9\x11\x9d\xbb\x10\xb2\x41\x46\xad\x18\x7b\x4f\xdb\xef\x19\x14\x21\x64\x51\x23\x18\xce\xb0\x04\xc3\x58\x21\x50\x20\x06\x47\x28\x22\xb5\xef\x4f\x38\x08\xc8\xd0\xb4\x51\x6d\x50\xdd\x32\xa1\xd1\xf2\x24\xf9\xc3\x3f\x7f\x34\x90\x29\xc7\x8f\x8f\x0a\x7a\xc2\x32\x11\x03\x33\x34\x0c\x1a\x48\x06\xaa\x5f\x47\xc8\x8a\x34\x70\x86\x6e\xda\x23\x05\x9a\x84\x9a\x9d\x41\x38\x87\x49\x0c\x47\x50\xe1\xe7\x65\xa0\x58\xe2\x8c\x94\xd1\xf0\xdb\x38\x41\x5c\x23\x7d\x24\xba\x91\x5e\xdb\x0f\xa0\x9e\x2b\xbc\xd9\xd4\x46\x08\xf9\xee\xdc\x48\x61\x52\xe7\xe3\x71\x24\x4d\x9c\xee\xc3\xb2\x2d\x6c\x3d\xd7\xce\xd8\x10\x66\xc6\x30\x1f\xce\x65\xb2\x87\xe7\xfb\xa7\xbf\xc1\x19\x8b\xc9\x12\xcd\xeb\xcb\xe3\xe5\xc0\xa8\x1a\xd5\x6c\xf2\xfe\x17\xa7\xc6\x0a\x18\x9b\xf3\xf7\xb3\x14\x81\xe8\x6d\x9e\x7f\x44\x57\xc7\x5f\xbf\xf2\xdf\x01\x00\x00\xff\xff\xea\x22\x97\x60\x40\x06\x00\x00")

func schemasTelemetrySetupConfiguration_requestJsonBytes() ([]byte, error) {
	return bindataRead(
		_schemasTelemetrySetupConfiguration_requestJson,
		"schemas/telemetry_setup_request.json",
	)
}

func schemasTelemetrySetupConfiguration_requestJson() (*asset, error) {
	bytes, err := schemasTelemetrySetupConfiguration_requestJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "schemas/telemetry_setup_request.json", size: 1600, mode: os.FileMode(436), modTime: time.Unix(1520589908, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}


var _schemasTestJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x4c\xcc\x31\x0e\xc2\x30\x0c\x85\xe1\x3d\xa7\xb0\x0c\x23\x25\x0c\x4c\xb9\x4d\x28\x26\x6d\xa5\xd6\x96\xed\x05\xa1\xde\x1d\x99\x48\xa8\xc3\x5b\x3e\xe9\x7f\x9f\x04\x80\x67\x1b\x27\x5a\x2b\x16\xc0\xc9\x5d\x4a\xce\x8b\xf1\x36\x74\xbd\xb2\xb6\xfc\xd4\xfa\xf2\xe1\x76\xcf\xdd\x4e\x78\x89\xce\xdf\x42\x11\xf1\x63\xa1\xd1\xbb\x89\xb2\x90\xfa\x4c\x86\x05\xe2\x1d\x00\x57\x32\xab\x8d\xfe\x70\x48\xcd\x75\xde\x1a\xfe\x78\x4f\xb1\x3d\x7d\x03\x00\x00\xff\xff\x42\x56\xee\x03\x94\x00\x00\x00")

func schemasTestJsonBytes() ([]byte, error) {
	return bindataRead(
		_schemasTestJson,
		"schemas/test.json",
	)
}

func schemasTestJson() (*asset, error) {
	bytes, err := schemasTestJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "schemas/test.json", size: 148, mode: os.FileMode(436), modTime: time.Unix(1519989957, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"schemas/create_identifiers.json": schemasCreate_identifiersJson,
	"schemas/install_filtering_rule.json": schemasInstall_filtering_ruleJson,
	"schemas/mitigation_request.json": schemasMitigation_requestJson,
	"schemas/session_configuration.json": schemasSession_configurationJson,
	"schemas/telemetry_setup_request.json": schemasTelemetrySetupConfiguration_requestJson,
	"schemas/test.json": schemasTestJson,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = &bintree{nil, map[string]*bintree{
	"schemas": &bintree{nil, map[string]*bintree{
		"create_identifiers.json": &bintree{schemasCreate_identifiersJson, map[string]*bintree{}},
		"install_filtering_rule.json": &bintree{schemasInstall_filtering_ruleJson, map[string]*bintree{}},
		"mitigation_request.json": &bintree{schemasMitigation_requestJson, map[string]*bintree{}},
		"session_configuration.json": &bintree{schemasSession_configurationJson, map[string]*bintree{}},
		"telemetry_setup_request.json": &bintree{schemasTelemetrySetupConfiguration_requestJson, map[string]*bintree{}},
		"test.json": &bintree{schemasTestJson, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

