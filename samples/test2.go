package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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

	var buildjobs = []string{}
	for _, build := range builds {
		if build.IsDir() {
			buildXmlFile := job_dir + "/" + build.Name() + "/build.xml"
			logFile := job_dir + "/" + build.Name() + "/log"
			_, err1 := os.Stat(buildXmlFile)
			_, err2 := os.Stat(logFile)
			if err1 == nil && err2 == nil {
				buildjobs = append(buildjobs, buildXmlFile)
			}
		}
	}

	numCPU := 4

	c := make(chan int, numCPU)

	for i := 0; i < numCPU; i++ {
		go func(i int, n int, u []string, c chan int) {
			for ; i < n; i++ {
				fmt.Printf("%d:%s\n", i, u[i])
			}
			c <- 1
		}(i*len(buildjobs)/numCPU, (i+1)*len(buildjobs)/numCPU, buildjobs, c)
	}

	for j := 0; j < numCPU; j++ {
		<-c
	}
}
