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
	d      dexec.Docker
	sums   = make(map[int]string)
	client docker.Client
	bufLen = 1024
)

type flushWriter struct {
	f http.Flusher
	w io.Writer
}

func (fw *flushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)
	if fw.f != nil {
		fw.f.Flush()
	}
	return
}

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
		Config: &docker.Config{Image: "gonitro/unoconv2"}})
	d := dexec.Docker{client}
	if err != nil {
		log.Fatal(err)
	}
	cmd := d.Command(m, "unoconv", "--stdin", "--stdout", "--format=txt")
	fmt.Println(cmd)
	fmt.Println(filename)

	//////

	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	///////

	fmt.Println("Starting StdinPipe")
	wc, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("StdinPipe error: %s\n", err)
	}
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	// cmd.Wait()
	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println(err)
	}

	wb, err := io.Copy(wc, file)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Copy written bytes: %v\n", wb)
	////////

	byteArray := []byte("myString")
	w.Write(byteArray)
	// w.Write(&cmd.Stdout)
	///////

	err = wc.Close()
	if err != nil {
		fmt.Printf("WriteCloser error: %s\n", err)
	}
	fmt.Println("Closed Pipe")

}

func writeCmdOutput(res http.ResponseWriter, pipeReader io.ReadCloser) {
	buffer := make([]byte, bufLen)
	for {
		fmt.Println("before reading")
		n, err := pipeReader.Read(buffer)
		fmt.Printf("Copy written bytes: %v err: %v \n", n, err)
		if err != nil || n <= 0 {
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

func toUpperDocker(w http.ResponseWriter, filename string) {
	client := clientConn()
	m, err := dexec.ByCreatingContainer(docker.CreateContainerOptions{
		Config: &docker.Config{Image: "busybox"}})
	d := dexec.Docker{client}
	if err != nil {
		log.Fatal(err)
	}
	cmd := d.Command(m, "tr", "[:lower:]", "[:upper:]")
	fmt.Println(cmd)

	cmd.Stdout = os.Stdout

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("ReadFile error: %v\n", err)
	}
	fmt.Println("Starting StdinPipe")
	wc, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("StdinPipe error: %s\n", err)
	}
	///

	///
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	// cmd.Stdout = w
	// io.Copy(w, cmd.Stdout)

	//stream the file to the WriteCloser
	_, err = fmt.Fprint(wc, string(file))
	if err != nil {
		fmt.Printf("file Fprint error: %s\n", err)
	}

	err = wc.Close()
	if err != nil {
		fmt.Printf("WriteCloser error: %s\n", err)
	}

}
func convertPdfHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GET params were:", r.URL.Query())
	filename := r.URL.Query().Get("filename")
	if filename != "" {
		unoconConvert(w, filename)
	}
}
func toUpperHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GET params were:", r.URL.Query())
	filename := r.URL.Query().Get("filename")
	if filename != "" {
		toUpperDocker(w, filename)
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
	http.HandleFunc("/convert/pdf", convertPdfHandler)
	http.HandleFunc("/toupper/txt", toUpperHandler)
	http.ListenAndServe(":8080", nil)
}
