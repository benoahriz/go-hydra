package main

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ahmetalpbalkan/go-dexec"
	"github.com/fsouza/go-dockerclient"
)

var (
	d       dexec.Docker
	sums    = make(map[int]string)
	client  docker.Client
	BUF_LEN = 1024
)

func clientConn() *docker.Client {
	endpoint := "unix:///var/run/docker.sock"
	client, _ := docker.NewClient(endpoint)
	return client
}

func echoTest(echotext string) string {
	input := `Hello world from container`
	return fmt.Sprintf("%s", input)
}

func toUpper(input string) string {
	client := clientConn()
	m, err := dexec.ByCreatingContainer(docker.CreateContainerOptions{
		Config: &docker.Config{Image: "busybox"}})
	if err != nil {
		panic(err)
	}
	d := dexec.Docker{client}
	cmd := d.Command(m, "sh", "-c", fmt.Sprintf("echo '%s' | tr '[:lower:]' '[:upper:]';echo \"ContainerId: ${HOSTNAME}\"", input))
	b, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%s", b)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>%s</h1>",
		toUpper("toUpper test "))
}

func unoconConvert(w http.ResponseWriter, filename string) {
	client := clientConn()
	m, err := dexec.ByCreatingContainer(docker.CreateContainerOptions{
		Config: &docker.Config{Image: "unoconv2"}})
	d := dexec.Docker{client}
	if err != nil {
		log.Fatal(err)
	}
	// cmd := d.Command(m, "tr", "[:lower:]", "[:upper:]")
	cmd := d.Command(m, "unoconv", "--stdin", "--stdout", "--format=txt")
	wc, err := cmd.StdinPipe() // <--
	if err != nil {
		panic(err)
	}
	cmd.Stdout = w
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
	}
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	fmt.Fprint(wc, file) // <--
	wc.Close()

	if err := cmd.Wait(); err != nil {
		panic(err)
	}

}

func writeCmdOutput(res http.ResponseWriter, pipeReader *io.PipeReader) {
	buffer := make([]byte, BUF_LEN)
	for {
		n, err := pipeReader.Read(buffer)
		if err != nil {
			pipeReader.Close()
			break
		}

		data := buffer[0:n]
		res.Write(data)
		if f, ok := res.(http.Flusher); ok {
			f.Flush()
		}
		//reset buffer
		for i := 0; i < n; i++ {
			buffer[i] = 0
		}
	}
}

func filenameHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GET params were:", r.URL.Query())
	filename := r.URL.Query().Get("filename")
	if filename != "" {
		unoconConvert(w, filename)
	}
}

// upload is an upload handler
func upload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("upload.gtpl")
		t.Execute(w, token)
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		fmt.Fprintf(w, "%v", handler.Header)
		f, err := os.OpenFile("./test/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
	}
}
func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/upload", upload)
	fs := http.FileServer(http.Dir("test"))
	http.Handle("/test/", http.StripPrefix("/test/", fs))
	http.HandleFunc("/convert/pdf", filenameHandler)
	http.ListenAndServe(":8080", nil)
}
