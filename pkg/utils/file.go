package utils

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func CreateDir(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return os.MkdirAll(dir, os.ModePerm)
	}
	return nil
}

// ReadPartFile read part of file: top 50 lines and bottom 50 lines
func ReadPartFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	const (
		TopLineNum    = 1000
		BottomLineNum = 1000
	)
	var (
		topLines    []string
		bottomLines []string
		count       int
	)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		count++
		line := scanner.Text()
		if count <= TopLineNum {
			topLines = append(topLines, line)
		} else {
			bottomLines = append(bottomLines, line)
			if len(bottomLines) > BottomLineNum {
				bottomLines = bottomLines[1:]
			}
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, fmt.Errorf("read file error: %s", err)
	}

	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown"
	}
	absPath, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	absPath = filepath.Join(absPath, path)

	if len(bottomLines) > 0 {
		if count > TopLineNum+BottomLineNum {
			ellipsis := fmt.Sprintf("\n...\n(%d lines more, original log file path: %s, hostname: %s)\n...\n", count-TopLineNum-BottomLineNum, absPath, hostname)
			topLines = append(topLines, ellipsis)
		}
		topLines = append(topLines, bottomLines...)
	}

	return topLines, nil
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// Zip ZIP 压缩
func Zip(srcFile string, destZip string) error {
	file, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer func(zipfile *os.File) {
		_ = zipfile.Close()
	}(file)

	archive := zip.NewWriter(file)
	defer func(archive *zip.Writer) {
		_ = archive.Close()
	}(archive)
	err = filepath.Walk(srcFile, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		// 移除根路径之前的路径，这样映射后是正常的目录结构
		header.Name = strings.TrimPrefix(path, srcFile+"/")
		if info.IsDir() {
			if path != srcFile {
				header.Name += "/"
			}
		} else {
			header.Method = zip.Deflate
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer func(file *os.File) {
				_ = file.Close()
			}(file)
			writer, err := archive.CreateHeader(header)
			if err != nil {
				return err
			}
			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}
		return err
	})

	return err
}
