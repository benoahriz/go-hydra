package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/ahmetalpbalkan/go-dexec"
	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/mux"
)

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

	cmd.Stdout = os.Stdout
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
	file, err := os.OpenFile("./test/"+filename, os.O_RDONLY, 0666)
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
	cmd := d.Command(m, "tr", "[:lower:]", "[:upper:]")
	log.Println(cmd)

	log.Printf("Read file %s", filename)
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("ReadFile error: %v\n", err)
	}

	log.Println("Starting StdinPipe")
	wc, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("StdinPipe error: %s\n", err)
	}
	log.Println("Starting StdoutPipe")
	rc, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("StdoutPipe error: %s\n", err)
	}
	var out bytes.Buffer
	// cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Stdout = &out
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	//stream the file to the WriteCloser
	log.Printf("Attempting to stream %s to writecloser %v bytes to send.", filename, len(file))
	n, err := fmt.Fprint(wc, string(file))
	fmt.Fprintln(wc, "EOF")
	if err != nil {
		log.Printf("file Fprint error: %s\n", err)
	}
	log.Printf("%d bytes streamed to the writeCloser", n)

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	log.Println("close writecloser")
	err = wc.Close()
	if err != nil {
		log.Printf("WriteCloser error: %s\n", err)
	}

	// c1 := exec.Command("ls")
	// c1.Stdout = &out
	// err = c1.Start()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	n64, err := out.ReadFrom(rc)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("out.readfrom number of bytes: %v", n64)

	err = rc.Close()
	if err != nil {
		log.Printf("ReadCloser error: %s\n", err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(out.Bytes())

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

// Index mux router
func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome!\n")
}

func TodoIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(todos); err != nil {
		panic(err)
	}
}

func TodoShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var todoId int
	var err error
	if todoId, err = strconv.Atoi(vars["todoId"]); err != nil {
		panic(err)
	}
	todo := RepoFindTodo(todoId)
	if todo.Id > 0 {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(todo); err != nil {
			panic(err)
		}
		return
	}

	// If we didn't find it, 404
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusNotFound)
	if err := json.NewEncoder(w).Encode(jsonErr{Code: http.StatusNotFound, Text: "Not Found"}); err != nil {
		panic(err)
	}

}

/*
Test with this curl command:
curl -H "Content-Type: application/json" -d '{"name":"New Todo"}' http://localhost:8080/todos
*/
func TodoCreate(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &todo); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	t := RepoCreateTodo(todo)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(t); err != nil {
		panic(err)
	}
}
