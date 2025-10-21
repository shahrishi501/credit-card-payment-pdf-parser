package utils

import "github.com/pdfcpu/pdfcpu/pkg/api"

func DecryptPDF(inputPath string, outputPath string, password string) error {
	config := api.LoadConfiguration()
	config.UserPW = password
	if err := api.DecryptFile(inputPath, outputPath, config); err != nil {
		return err
	}
	return nil
}