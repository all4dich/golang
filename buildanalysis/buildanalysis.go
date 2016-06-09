package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/all4dich/golang/buildanalysis/oebuildjobs"
)

func Filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func ParseMeta(params ...string) (paramData []string) {
	data := ""
	sep := ""
	switch len(params) {
	case 1:
		data = params[0]
		sep = " "
	case 2:
		data = params[0]
		sep = params[1]
	default:
		return []string{}
	}
	paramData = Filter(strings.Split(data, sep), func(i string) bool {
		if i == "" {
			return false
		} else if i == " " {
			return false
		} else {
			return true
		}
	})
	return paramData
}

func AnalyzeBuild(buildDir string) {
	start := time.Now()
	buildLogFile := buildDir + "/log"
	buildXmlFile := buildDir + "/build.xml"
	buildEle := strings.Split(buildDir, "/")
	buildJobName := buildEle[len(buildEle)-3]
	buildNumber := buildEle[len(buildEle)-1]
	buildInfo := make(map[string]string)
	buildLog, err := os.Open(buildLogFile)
	if err == nil {
		buildLogReader := bufio.NewReader(buildLog)
		var (
			isPrefix bool  = true
			err      error = nil
			line     []byte
		)
		var _ = isPrefix

		validId := regexp.MustCompile(`^(BB_VERSION|BUILD_SYS|DATETIME|DISTRO|DISTRO_VERSION|MACHINE|NATIVELSBSTRING|TARGET_FPU|TARGET_SYS|TUNE_FEATURES|WEBOS_DISTRO_BUILD_ID|WEBOS_DISTRO_MANUFACTURING_VERSION|WEBOS_DISTRO_RELEASE_CODENAME|WEBOS_DISTRO_TOPDIR_DESCRIBE|WEBOS_DISTRO_TOPDIR_REVISION|WEBOS_ENCRYPTION_KEY_TYPE|meta|meta-qt5|meta-starfish-product)\ .*`)
		validTimeBuildsh := regexp.MustCompile(`.*build\.sh\ +--machines*`)
		validRmBuild := regexp.MustCompile(`.*rm\ -rf.*BUILD$`)
		validRmBuildArtifacts := regexp.MustCompile(`.*rm\ -rf.*BUILD-ARTIFACTS$`)
		validRmDownloads := regexp.MustCompile(`.*rm\ -rf.*downloads$`)
		validRmSstatecache := regexp.MustCompile(`.*rm\ -rf.*sstate-cache$`)
		validRsyncArtifacts := regexp.MustCompile(`.*rsync\ -arz.*\ BUILD-ARTIFACTS.*`)
		validNoOfScratch := regexp.MustCompile(`NOTE:\s+do_populate_lic.*sstate.*`)
		/*
		   time_build_sh = 0.0
		   time_rm_BUILD = 0.0
		   time_rm_BUILD_ARTIFACTS = 0.0
		   time_rm_downloads = 0.0
		   time_rm_sstatecache = 0.0
		   time_rsync_artifacts = 0.0
		   time_rsync_ipk = 0.0
		   num_of_from_scratch = 0
		*/
		for err == nil {
			line, isPrefix, err = buildLogReader.ReadLine()
			eachLine := string(line)
			eachLineSplit := strings.Split(eachLine, " ")
			if validId.MatchString(eachLine) {
				r := ParseMeta(eachLine)
				if _, ok := buildInfo[r[0]]; !ok {
					if len(r) == 3 {
						buildInfo[r[0]] = r[2]
					}
				}
			}
			if validTimeBuildsh.MatchString(eachLine) {
				buildInfo["time_build_sh"] = eachLineSplit[2]
				continue
			}
			if validRmBuild.MatchString(eachLine) {
				buildInfo["time_rm_BUILD"] = eachLineSplit[2]
				continue
			}
			if validRmBuildArtifacts.MatchString(eachLine) {
				buildInfo["time_rm_BUILD_ARTIFACTS"] = eachLineSplit[2]
				continue
			}
			if validRmDownloads.MatchString(eachLine) {
				buildInfo["time_rm_downloads"] = eachLineSplit[2]
				continue
			}
			if validRmDownloads.MatchString(eachLine) {
				buildInfo["time_rm_downloads"] = eachLineSplit[2]
				continue
			}
			if validRmSstatecache.MatchString(eachLine) {
				buildInfo["time_rm_sstatecache"] = eachLineSplit[2]
				continue
			}
			if validNoOfScratch.MatchString(eachLine) {
				buildInfo["num_of_from_scratch"] = eachLineSplit[2]
				continue
			}
			if validRsyncArtifacts.MatchString(eachLine) {
				buildInfo["time_rsync_artifacts"] = eachLineSplit[2]
				continue
			}
		}
	}
	var _ = xml.Header
	var _ = oebuildjobs.VerifyBuild{}
	var _ = buildXmlFile
	buildXml, err := os.Open(buildXmlFile)
	if err != nil {
		log.Fatal("Can't open a file " + buildXmlFile)
	}
	r := bufio.NewReader(buildXml)
	buildXmlDat, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal("ERROR: Cannot read data from a xml file ")
	}
	xmlEntity := oebuildjobs.VerifyBuild{}
	err = xml.Unmarshal(buildXmlDat, &xmlEntity)
	fmt.Printf("%s,%s,%s,%s\n", buildJobName, buildNumber, xmlEntity, buildInfo["time_build_sh"])
	log.Println("FINISHED: ", time.Since(start))
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
	runtime.GOMAXPROCS(16)
	job_dir := *jenkinsHome + "/jobs/" + *jobName + "/builds"
	builds, err := ioutil.ReadDir(job_dir)
	if err != nil {
		log.Fatal(err)
	}

	buildjobs := make(chan string, *nThread)
	done := make(chan int, *nThread)
	fmt.Println("Job Name, Build Number, Result, Host, Duration, Start, Gerrit Received, Time Diff")
	for j := 0; j < *nThread; j++ {
		go func(j int) {
			for buildJob := range buildjobs {
				AnalyzeBuild(buildJob)
			}
			done <- 1
		}(j)
	}
	go func() {
		log.Println("Start: Getting build directories")
		validInt := regexp.MustCompile(`^\d+$`)
		for _, build := range builds {
			if build.IsDir() && validInt.MatchString(build.Name()) {
				buildXmlFile := job_dir + "/" + build.Name() + "/build.xml"
				logFile := job_dir + "/" + build.Name() + "/log"
				_, err1 := os.Stat(buildXmlFile)
				_, err2 := os.Stat(logFile)
				if err1 == nil && err2 == nil {
					buildjobs <- job_dir + "/" + build.Name()
				}
			}
		}
		log.Println("END: Getting build directories")
		close(buildjobs)
	}()
	for m := 0; m < *nThread; m++ {
		<-done
	}
	log.Println("END:")
}
