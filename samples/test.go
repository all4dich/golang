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

	var i int = 0

	for _, build := range builds {
		if build.IsDir() {
			buildXmlFile := job_dir + "/" + build.Name() + "/build.xml"
			logFile := job_dir + "/" + build.Name() + "/log"
			_, err1 := os.Stat(buildXmlFile)
			_, err2 := os.Stat(logFile)
			if err1 == nil && err2 == nil {
				i = i + 1
			}
			if err1 != nil {
				fmt.Println("Build: " + build.Name())
				fmt.Println("	build xml not found")
			}
			if err2 != nil {
				fmt.Println("Build: " + build.Name())
				fmt.Println("	log file not found")
			}
		}
	}
	fmt.Printf("number of builds: %d\n", i)
	fmt.Println(os.Args)

}
