package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func buildHandler(w http.ResponseWriter, r *http.Request) {
	gituser := r.URL.Query().Get("user")
	reponame := r.URL.Query().Get("repo")
	gobuilder := r.URL.Query().Get("gobuild")

	log.Println("checking if repo exists")
	repoResponse := repoCheck(reponame)
	if repoResponse != true {
		log.Println("repo does not exist, running git clone")
		cloneResponse := gitclone(gituser, reponame)
		if cloneResponse > 0 {
			log.Println("an error occured cloning repo")
			log.Println("user:" + gituser + ", repo:" + reponame)
			io.WriteString(w, "error cloning repo\n")
			return
		}
		log.Println("clone sucessful")
		io.WriteString(w, "clone sucessful\n")
	}
	log.Println("running git pull")
	gitresponse := gitpull(gituser, reponame)
	if gitresponse > 0 {
		log.Println("an error occured running git pull")
		log.Println("user:" + gituser + ", repo:" + reponame)
		io.WriteString(w, "error running git pull\n")
		return
	}
	log.Println("git pull sucessful")
	io.WriteString(w, "pull sucessful")
	if gobuilder == "true" {
		log.Println("running go build")
		gobuildresponse := gobuild(reponame)
		if gobuildresponse > 0 {
			log.Println("an error occured running go build")
			log.Println("reponame:" + reponame)
			io.WriteString(w, "error running go build\n")
			return
		}
		io.WriteString(w, "go build sucessful\n")
		log.Println("go build sucessful")
		log.Println("copying cupserviosr confg")
		err := copySupervisorConf(reponame)
		if err != nil {
			log.Println("error occured copying supervisor conf\n")
			log.Println(err)
			io.WriteString(w, "error copying supervisor conf\n")
			return
		}
		io.WriteString(w, "supervisor conf copied")
		log.Println("supervisor conf copied sucessfully\n")
		log.Println("restarting app in supervisor\n")
		supervisorReturn := restartSupervisor(reponame)
		if supervisorReturn > 0 {
			log.Println("an error occured restarting supervisor")
			io.WriteString(w, "error restarting supervisor\n")
			return
		}
		io.WriteString(w, "supervisor restarted")
		log.Println("supervisor restarted sucessfully")
		io.WriteString(w, "build sucessful")
		log.Println("build sucessful")
	}
}

func gitpull(gituser string, reponame string) int {
	os.Chdir(myrepos + "/" + reponame)
	cmd := "git"
	args := []string{"pull"}

	if err := exec.Command(cmd, args...).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func gitclone(gituser string, reponame string) int {
	os.Chdir(myrepos)

	cmd := exec.Command("git", "clone", "git@github.com:"+gituser+"/"+reponame)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return 1
	}
	fmt.Println("Result:  " + out.String())
	return 0
}

func gobuild(reponame string) int {
	os.Chdir(myrepos + "/" + reponame)

	cmd := exec.Command("go", "build")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return 1
	}
	fmt.Println("Result:  " + out.String())
	return 0
}

func repoCheck(repo string) bool {
	if _, err := os.Stat(myrepos + "/" + repo); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	return true
}

func copySupervisorConf(reponame string) error {
	in, err := os.Open(myrepos + "/" + reponame + "/" + reponame + ".conf")
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create("/etc/supervisor/conf.d/" + reponame + ".conf")
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		return err
	}
	return cerr
}

func restartSupervisor(reponame string) int {
	cmd := exec.Command("supervisorctl", "restart", reponame)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return 1
	}
	fmt.Println("Result:  " + out.String())
	return 0
}
