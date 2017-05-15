package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/google/go-github/github"
)

var queue = make(chan *github.PullRequestEvent)

func CheckoutMergedPullRequest(pre *github.PullRequestEvent) {
	log.Println("processing ", pre.Number)
	dir, err := ioutil.TempDir("/tmp", "ci")
	if err != nil {
		panic(err)
	}

	log.Println("temporary directory: ", dir)
	os.Chdir(dir)
	output, err := exec.Command("git", "clone", *pre.Repo.CloneURL).CombinedOutput()
	log.Printf("%s\n", output)
	if err != nil {
		panic(err)
	}
	os.Chdir(*pre.Repo.Name)

	output, err = exec.Command("git-pr", strconv.Itoa(*pre.Number)).CombinedOutput()
	log.Printf("%s\n", output)

	output, err = exec.Command("git", "merge", "master").CombinedOutput()
	log.Printf("%s\n", output)

	// now test it

	//cmd := fmt.Sprintf("curl -L %s | git am", *pre.PullRequest.PatchURL)
	//log.Println(cmd)
	//output, err = exec.Command("bash", "-c", cmd).CombinedOutput()
	//log.Printf("%s\n", output)
}

func event_handler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	var event = r.Header.Get("X-Github-Event")
	if event == "pull_request" {
		decoder := json.NewDecoder(r.Body)
		var pre github.PullRequestEvent
		err := decoder.Decode(&pre)
		log.Println("queueing pull request: ", *pre.Number)
		queue <- &pre
		//CheckoutMergedPullRequest(&pre)
		if err != nil {
			io.WriteString(w, err.Error())
		}
		io.WriteString(w, "queued")
	} else {
		io.WriteString(w, "ignored")
	}
}

func main() {
	go func() {
		for {
			log.Println("starting goroutine")
			pre := <-queue
			log.Println("dequeing pull request: ", *pre.Number)
			CheckoutMergedPullRequest(pre)
		}
	}()
	http.HandleFunc("/event_handler", event_handler)
	http.ListenAndServe(":8000", nil)
}
