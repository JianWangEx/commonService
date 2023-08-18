package generate_obj_and_mock_file

const getAllPackagesCmd = "go list ./..."

type Config struct {
	IdentifierOfProvider string
	IdentifierOfInjector string
	OutputDirPathName    string // please do not include "./" start directly with first subdirectory name
	OutputPackageName    string // must match the last directory name in OutputDirPathName
	CheckSumFileName     string
	ScanPkgs             []string
	Incremental          bool   // incremental update
	FilterKeyword        string // keyword of the directory, can be any words like "/services/" or "internal/services" or "client", or "api" .etc, better with "/" so it can be more precise
	Debug                bool   // Debug model
}

func (c Config) PartialMode() bool {
	if c.Incremental || c.FilterKeyword != "" {
		return true
	}
	return false
}

var DefaultConfig = Config{
	IdentifierOfProvider: "@Provider",
	IdentifierOfInjector: "@Injector",
	OutputDirPathName:    "internal/register",
	OutputPackageName:    "register",
	CheckSumFileName:     "checksum.txt",
	ScanPkgs: []string{
		"/internal/services/",
		"/internal/operators",
		"/internal/service_common/",
		"/internal/data/dao/",
		"/pkg/",
	},
}
