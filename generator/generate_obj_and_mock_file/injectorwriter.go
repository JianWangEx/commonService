package generate_obj_and_mock_file

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
)

type InjectorWriter struct {
	config       Config
	outputBytes  map[string][]byte // relative_file_path->content, such as internal/business/payment_method/impl/operator/service/payment_method_service_injector.go->content
	IsMockAppend bool
}

func NewInjectorWriter(config Config, outputBytes map[string][]byte) *InjectorWriter {
	return &InjectorWriter{config: config, outputBytes: outputBytes}
}

// Write is very simple, just creates and writes according the outputBytes(relative_file_path->content).
//
//		For normal case, it will override exist content, and create files like "internal/business/payment/xxx_injector.go" besides the original interface file.
//	 But if IsMockAppend is true, it will append the content to the end of the file if only the content hasn't existed in the file.
func (w *InjectorWriter) Write() error {
	for fileName, outputBytes := range w.outputBytes {
		// Create file for each file that contains injectors
		flag := os.O_RDWR | os.O_CREATE | os.O_TRUNC
		if w.IsMockAppend {
			flag = os.O_APPEND | os.O_WRONLY
			_, err := os.Stat(fileName)
			if os.IsNotExist(err) {
				// may be in mock free mode, then no mock file is generated, can ignore.
				continue
			}
		}
		f, err := os.OpenFile(fileName, flag, 0755)
		if err != nil {
			return err
		}
		if w.IsMockAppend {
			content, _ := ioutil.ReadFile(fileName)
			if bytes.Contains(content, outputBytes) {
				// if already contains the content, no need to write again
				// it may happen in incremental mode
				continue
			}
		}
		_, err = f.Write(outputBytes)
		if err != nil {
			return err
		}

		err = f.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	return nil
}
