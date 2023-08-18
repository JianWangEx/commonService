package generate_obj_and_mock_file

import (
	"fmt"
	"os"
	"strings"
)

type ProviderWriter struct {
	config Config
	// category->file_name->content
	// Example: business -> paymentimploperator_payment_url_operator_impl.go -> content(ioc.AddProvider...)
	outputBytes map[string]map[string][]byte
	IsMock      bool
}

func NewProviderWriter(config Config, outputBytes map[string]map[string][]byte) *ProviderWriter {
	return &ProviderWriter{config: config, outputBytes: outputBytes}
}

// Write creates and writes the ProviderWriter.outputBytes to files. will remove all the category directory first if not incremental mode. If IsMock, the files will be created under "mock" subdirectory, and others remains same
func (w *ProviderWriter) Write() error {
	if w.IsMock {
		w.config.OutputDirPathName += "/mock"
	}
	_ = w.MakeDir()
	for category, nameToBytes := range w.outputBytes {
		dirPath := w.config.OutputDirPathName + "/" + category
		if !w.config.PartialMode() { // in PartialMode, we cannot remove existing directory
			// Remove existing directory
			err := os.RemoveAll("./" + dirPath)
			if err != nil {
				return err
			}
		}
		// Make subdirectory for each register type / layer
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			return err
		}

		for name, bytes := range nameToBytes {
			// Create file for each file that contains providers
			filePath := dirPath + "/" + name
			f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}
			_, err = f.Write(bytes)
			if err != nil {
				return err
			}
			err = f.Close()
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}
	return nil
}

// nolint:gosec
func (w *ProviderWriter) MakeDir() error {
	dirArr := strings.Split(w.config.OutputDirPathName, "/")
	var currDirStr string
	for _, dir := range dirArr {
		if dir == "" || dir == "." {
			continue
		}
		if currDirStr == "" {
			currDirStr = dir
		} else {
			currDirStr += "/" + dir
		}
		if _, err := os.Stat(currDirStr); os.IsNotExist(err) {
			fmt.Printf("\ncreating directory %v\n", currDirStr)
			err = os.Mkdir(currDirStr, 0755)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
