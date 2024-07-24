package main

import (
	"ConcurrentFileServer/core"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type ExistBool struct {
	Exist bool `json:"exist"`
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("D:\\Golang\\ConcurrentFileServer\\web\\templates\\forms.html"))
	if r.Method != http.MethodPost {
		tmpl.Execute(w, nil)
		return
	}
	file, _, _ := r.FormFile("file")
	defer file.Close()

	content, _ := ioutil.ReadAll(file)
	ctx := context.Background()
	handler := core.NewFileHandlerImpl()
	fileID, err := handler.UploadFile(ctx, content, strings.Split(http.DetectContentType(content), ";")[0])
	if err != nil {
		panic(err)
		return
	}
	tmpl.Execute(w, struct {
		Success bool
		FileId  string
	}{true, fileID})

}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("D:\\Golang\\ConcurrentFileServer\\web\\templates\\downloadforms.html"))
	if r.Method != http.MethodPost {
		tmpl.Execute(w, nil)
		return
	}
	fildId := r.FormValue("fileId")
	handler := core.NewFileHandlerImpl()
	ctx := context.Background()
	file, header, err := handler.DownloadFile(ctx, fildId)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", header)
	w.Write(file)
}

func existFile(w http.ResponseWriter, r *http.Request) {
	fileID := r.FormValue("fileID")
	handler := core.NewFileHandlerImpl()
	ctx := context.Background()
	exist, err := handler.ExistFile(ctx, fileID)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Println(fileID)
	if exist {
		response := ExistBool{true}
		json.NewEncoder(w).Encode(response)
	} else {
		response := ExistBool{false}
		json.NewEncoder(w).Encode(response)
	}
}

func main() {
	http.HandleFunc("/uploadfile", uploadFile)
	http.HandleFunc("/downloadfile", downloadFile)
	http.HandleFunc("/existfile", existFile)

	err := http.ListenAndServe(":8000", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
