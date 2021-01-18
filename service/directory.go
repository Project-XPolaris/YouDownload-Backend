package service

import (
	"io/ioutil"
	"path/filepath"
)

type FileItem struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
}
func ReadDirectory(dirPath string) ([]FileItem,error) {
	path,err := filepath.Abs(filepath.Clean(dirPath))
	if err != nil {
		return nil, err
	}
	items,err := ioutil.ReadDir(path)
	if err != nil {
		return nil,err
	}
	fileItems := make([]FileItem,0)
	for _, item := range items {
		i := FileItem{
			Name: item.Name(),
			Path: filepath.Join(path,item.Name()),
			Type: "File",
		}
		if item.IsDir() {
			i.Type = "Directory"
		}
		fileItems = append(fileItems, i)
	}
	return fileItems,nil
}