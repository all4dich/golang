package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"runtime"
	"time"
)

func AnalyzeBuild(buildDir string) {
	//buildXmlFile := buildDir + "/build.xml"
	buildLogFile := buildDir + "/log"

	buildLog, err := os.Open(buildLogFile)
	if err == nil {
		/*
			buildLogReader := bufio.NewReader(buildLog)
			fmt.Println(buildLogReader.ReadString('\n'))
		*/
		buildLogScanner := bufio.NewScanner(buildLog)
		//WEBOS_DISTRO_TOPDIR_REVISION
		start := time.Now()
		for buildLogScanner.Scan() {
			eachLine := buildLogScanner.Text()
			validId := regexp.MustCompile(`^WEBOS_DISTRO_TOPDIR_REVISION.*`)
			if validId.MatchString(eachLine) {
				var _ = eachLine
				//fmt.Println(eachLine)
			}
			//fmt.Println(buildLogScanner.Text())
		}
		fmt.Println(time.Since(start))
	}
}

func main() {
	jenkinsHome := flag.String("jenkinsHome", "/binary/build_results/jenkins_home_backup", "Jenkins configuration and data directory")
	jobName := flag.String("jobName", "starfish-drd4tv-official-h15", "Set a job name to parse")
	nThread := flag.Int("n", 4, "Number of threads")
	flag.Parse()

	log.Printf("Jenkins Home: %s", *jenkinsHome)
	log.Printf("Job: %s", *jobName)
	log.Printf("Number of threads: %d", *nThread)
	log.Printf("runtime: %d", runtime.NumCPU())
	job_dir := *jenkinsHome + "/jobs/" + *jobName + "/builds"
	builds, err := ioutil.ReadDir(job_dir)
	if err != nil {
		log.Fatal(err)
	}

	buildjobs := make(chan string, *nThread)
	done := make(chan int)
	for j := 0; j < *nThread; j++ {
		go func(j int) {
			for buildJob := range buildjobs {
				//time.Sleep(time.Millisecond * 100)
				AnalyzeBuild(buildJob)
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
					buildjobs <- job_dir + "/" + build.Name()
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
