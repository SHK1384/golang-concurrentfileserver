package core

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	mathRand "math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const MySecret string = "abc&1*~#^2^#s0^=)^^7%b34"

func encrypt(key, text []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	stringed := string(ciphertext)
	answer := ""
	for _, c := range stringed {
		answer += string((c % 26) + 97)
	}
	return answer, nil
}

type FileHandler interface {
	UploadFile(ctx context.Context, file []byte, mimeType string) (string, error)
	DownloadFile(ctx context.Context, fileID string) ([]byte, string, error)
	ExistFile(ctx context.Context, fileID string) (bool, error)
}

type FileHandlerImpl struct {
}

func NewFileHandlerImpl() FileHandler {
	return &FileHandlerImpl{}
}

func (f *FileHandlerImpl) UploadFile(ctx context.Context, file []byte, mimeType string) (string, error) {
	encText, err := encrypt([]byte(MySecret), file[:min(10, len(file))])
	if err != nil {
		return "", err
	}
	fileID := fmt.Sprintf("%s-%s", strconv.FormatUint(mathRand.Uint64(), 10), encText)
	fl, err := os.Create(fmt.Sprintf("./files/%s.txt", fileID))
	if err != nil {
		return "ops", err
	}
	defer fl.Close()
	_, err = fl.Write(file)
	if err != nil {
		return "pos", err
	}
	return fileID, nil
}

func (f *FileHandlerImpl) DownloadFile(ctx context.Context, fileID string) ([]byte, string, error) {
	start := time.Now()
	files, err := os.ReadDir("./files")
	var fileName = ""
	for _, file := range files {
		if strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())) == fileID {
			fileName = file.Name()
		}
	}
	fileFullName := fmt.Sprintf("./files/%s", fileName)

	fl, err := os.Open(fileFullName)
	if err != nil {
		panic(err)
	}
	defer fl.Close()
	fi, err := fl.Stat()
	buf := make([]byte, 10)
	var wg sync.WaitGroup
	var myFile []byte
	var fileSize = fi.Size()
	var sz = int(fileSize / 10)
	for i := 0; i < sz; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b, err := fl.Read(buf)
			fmt.Println(string(buf))
			if err != io.EOF {
				myFile = append(myFile, buf[:b]...)
			}
		}()
	}
	wg.Wait()
	buf = make([]byte, fileSize%10)
	b, err := fl.Read(buf)
	if err != io.EOF {
		myFile = append(myFile, buf[:b]...)
	}
	delta := time.Since(start)
	fmt.Println(delta)
	return myFile, strings.Split(http.DetectContentType(myFile), ";")[0], nil
}

func (f *FileHandlerImpl) ExistFile(ctx context.Context, fileID string) (bool, error) {
	files, err := os.ReadDir("./files")
	if err != nil {
		return false, err
	}
	var fileName = ""
	for _, file := range files {
		if strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())) == fileID {
			fileName = file.Name()
		}
	}
	if fileName == "" {
		return false, nil
	} else {
		return true, nil
	}
}
