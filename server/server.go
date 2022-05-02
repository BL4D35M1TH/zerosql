package server

import (
	"database/sql"
	"embed"
	"html/template"
	"log"
	"net/http"
	"sanndy/database"
	"strconv"
	"strings"
)

var DefaultTags = []string{"sfw", "nsfw", "boobs", "belly", "thighs", "armpit", "legs", "cleavage"}

//go:embed static/*
var static embed.FS

type Server struct {
	Router     http.Handler
	ImageStore database.IData
	DataStore  database.Store
	Templates  *template.Template
}

type ImgVer struct {
	Original string
	Filename string
}

func CreateServer(root, db string) (http.Handler, error) {
	router := http.NewServeMux()
	sqlDB, err := sql.Open("sqlite3", db)
	if err != nil {
		return nil, err
	}
	dataStore, err := database.CreateStore(sqlDB)
	if err != nil {
		return nil, err
	}
	imageStore, err := database.CreateStorage(root, dataStore)
	if err != nil {
		return nil, err
	}
	templates, err := template.ParseFS(static, "static/*.html")
	if err != nil {
		return nil, err
	}
	server := Server{router, imageStore, dataStore, templates}
	router.HandleFunc("/upload", server.UploadImage)
	router.HandleFunc("/view", server.RenderImages)
	router.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir(root))))
	return server.Router, nil
}

func (srv *Server) RenderImages(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 0
	}
	tags := r.URL.Query()["tags"]
	images, err := srv.DataStore.ByTags(tags, int64(page*10), 10, database.DESC)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	var imgNames []ImgVer
	for _, img := range images {
		filename := strings.Split(img.Path, ".")[0]
		imgNames = append(imgNames, ImgVer{img.Path, filename})
	}
	pageData := struct {
		Prev, Next int
		List       []ImgVer
	}{
		Prev: page - 1,
		Next: page + 1,
		List: imgNames,
	}
	err = srv.Templates.ExecuteTemplate(w, "view.html", pageData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("IP: %s, Page: %d, Images: %d", r.RemoteAddr, page, len(images))
}

func (srv *Server) UploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		file, headers, err := r.FormFile("image")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		tags := r.Form["tags"]
		err = srv.ImageStore.Save(file, tags)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Printf("%s %s %s %d", r.RemoteAddr, headers.Filename, headers.Header.Get("Content-Type"), headers.Size/1000)
	}
	err := srv.Templates.ExecuteTemplate(w, "upload.html", DefaultTags)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("IP: %s", r.RemoteAddr)

}
