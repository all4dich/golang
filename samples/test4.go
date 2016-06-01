package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func main() {
	jenkinsHome := flag.String("jenkinsHome", "/binary/build_results/jenkins_home_backup", "Jenkins configuration and data directory")
	jobName := flag.String("jobName", "starfish-drd4tv-official-h15", "Set a job name to parse")
	flag.Parse()
	job_dir := *jenkinsHome + "/jobs/" + *jobName + "/builds"
	builds, err := ioutil.ReadDir(job_dir)
	if err != nil {
		log.Fatal(err)
	}

	buildjobs := make(chan string, 10)
	done := make(chan int)
	for j := 0; j < 4; j++ {
		go func(j int) {
			for buildJob := range buildjobs {
				time.Sleep(time.Millisecond * 100)
				var _ = buildJob
			}
			done <- 1
		}(j)
	}
	go func() {
		log.Println("Start: Building")
		for _, build := range builds {
			if build.IsDir() {
				buildXmlFile := job_dir + "/" + build.Name() + "/build.xml"
				logFile := job_dir + "/" + build.Name() + "/log"
				_, err1 := os.Stat(buildXmlFile)
				_, err2 := os.Stat(logFile)
				if err1 == nil && err2 == nil {
					buildjobs <- buildXmlFile
				}
			}
		}
		log.Println("END: Building")
		close(buildjobs)
		fmt.Println(buildjobs)
	}()
	<-done
	fmt.Println("END:")
}
