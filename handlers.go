package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
)

var customContentType = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".html": "text/html",
	".txt":  "text/plain",
	".sh":   "text/plain",
	".m3u":  "text/plain",
	".pls":  "text/plain",
	".org":  "text/plain",
	".pdf":  "application/pdf",
}

// Extension convertion map
var extConvertions = map[string]string{
	".sh":  ".txt",
	".bat": ".txt",
	".m3u": ".txt",
	".pls": ".txt",
}

// Handles and processes the home page
func home(w http.ResponseWriter, r *http.Request) {
	tmpl.Execute(w, template.HTML(fmt.Sprintf(`https://%s/`, r.Host)))
}

// Upload a file, save and attribute a hash
func upload(w http.ResponseWriter, r *http.Request) {
	glog.Info("Upload request recieved")

	// Prepare to get the file
	file, header, err := r.FormFile("file")
	defer func() {
		file.Close()
		glog.Infof(`File "%s" closed.`, header.Filename)
	}()
	if err != nil {
		glog.Errorf("Error retrieving file.")
		glog.Errorf("Error: %s", err.Error())

		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad request. Error retrieving file.")
		return
	}

	const storagePath = "./storage"
	if err := os.MkdirAll(storagePath, 0777); err != nil {
		glog.Error("Error creating storage on server...")
		glog.Errorf("Error: %s", err.Error())

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "No storage available.")
		return
	}

	// Creates filename with UUID
	var uuid string = GenerateUUID()
	// Adds inherited extension
	var ext = filepath.Ext(header.Filename)
	if len(ext) > 0 {
		newExt, ok := extConvertions[ext]
		if ok {
			ext = newExt
		}
	}
	var filename string = fmt.Sprintf("%s%s", uuid, ext)
	// Ensures file is unique
	_, err = os.Stat(filename)
	for !os.IsNotExist(err) {
		uuid = GenerateUUID()
		filename := fmt.Sprintf("%s%s", uuid, ext)
		_, err = os.Stat(filename)
	}

	f, err := os.OpenFile(path.Join(storagePath, filename), os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		glog.Errorf("Error creating file.")
		glog.Errorf("Error: %s", err.Error())

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error creating file.")
		return
	}
	defer f.Close()

	if _, err := io.Copy(f, file); err != nil {
		glog.Errorf("Error writing file.")
		glog.Errorf("Error: %s", err.Error())

		w.WriteHeader(http.StatusInsufficientStorage)
		fmt.Fprintf(w, "Insufficient Storage. Error storing file.")
		return
	}

	// All good
	fmt.Fprintf(w, "OK, Successfully Uploaded\n https://%s/%s\n", r.Host, filename)
}

// Gets the file using the provided UUID on the URL
func getFile(w http.ResponseWriter, r *http.Request) {
	glog.Info(fmt.Sprintf("Retrieve request received: %s", r.URL.Path))
	var filename string = strings.Replace(r.URL.Path[1:], "/", "", -1)
	var path string = fmt.Sprintf("./storage/%s", filename)

	glog.Infof(`Retrieving Path "%s"`, path)

	_, err := os.Stat(path)
	if err != nil {
		glog.Errorf(`Error checking filepath "%s"`, path)
		glog.Errorf("Error: %s", err.Error())
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "File Not Found.")
		return
	}

	glog.Infof(`Retrieving Filename "%s"`, fmt.Sprintf("./%s", filename))

	ext := filepath.Ext(filename)
	if ct, ok := customContentType[ext]; ok {
		glog.Infof(`Setting custom Content-Type "%s"`, ct)
		w.Header().Set("Content-Type", ct)
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%s", filename))
	}
	http.ServeFile(w, r, path)
}
