package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/all4dich/golang/buildanalysis/oebuildjobs"
)

var validId = regexp.MustCompile(`^(BB_VERSION|BUILD_SYS|DATETIME|DISTRO|DISTRO_VERSION|MACHINE|NATIVELSBSTRING|TARGET_FPU|TARGET_SYS|TUNE_FEATURES|WEBOS_DISTRO_BUILD_ID|WEBOS_DISTRO_MANUFACTURING_VERSION|WEBOS_DISTRO_RELEASE_CODENAME|WEBOS_DISTRO_TOPDIR_DESCRIBE|WEBOS_DISTRO_TOPDIR_REVISION|WEBOS_ENCRYPTION_KEY_TYPE|meta|meta-qt5|meta-starfish-product)\ .*`)
var validTimeBuildsh = regexp.MustCompile(`.*build\.sh\ +--machines*`)
var validRmBuild = regexp.MustCompile(`.*rm\ -rf.*BUILD$`)
var validRmBuildArtifacts = regexp.MustCompile(`.*rm\ -rf.*BUILD-ARTIFACTS$`)
var validRmDownloads = regexp.MustCompile(`.*rm\ -rf.*downloads$`)
var validRmSstatecache = regexp.MustCompile(`.*rm\ -rf.*sstate-cache$`)
var validRsyncArtifacts = regexp.MustCompile(`.*rsync\ -arz.*\ BUILD-ARTIFACTS.*`)
var validNoOfScratch = regexp.MustCompile(`NOTE:\s+do_populate_lic.*sstate.*`)

func CheckBuildExistInDB(buildDir string, coll *mgo.Collection) bool {
	buildEle := strings.Split(buildDir, "/")
	buildJobName := buildEle[len(buildEle)-3]
	buildNumber, _ := strconv.Atoi(buildEle[len(buildEle)-1])
	//n, err := coll.Find(bson.M{"jobname": buildJobName}).Select(bson.M{"buildnumber": buildNumber}).Count()
	n, err := coll.Find(bson.M{"jobname": buildJobName, "$and": []interface{}{
		bson.M{"buildnumber": buildNumber},
	}}).Count()
	if err != nil {
		log.Println(err)
		return true
	}
	if n == 1 {
		return true
	} else {
		return false
	}
}

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

func AnalyzeBuild(buildDir string) string {
	start := time.Now()
	var _ = start
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
		keyMap := map[string]string{"BB_VERSION": "", "BUILD_SYS": "", "DATETIME": "", "DISTRO": "", "DISTRO_VERSION": "", "MACHINE": "", "NATIVELSBSTRING": "", "TARGET_FPU": "", "TARGET_SYS": "", "TUNE_FEATURES": "", "WEBOS_DISTRO_BUILD_ID": "", "WEBOS_DISTRO_MANUFACTURING_VERSION": "", "WEBOS_DISTRO_RELEA    SE_CODENAME": "", "WEBOS_DISTRO_TOPDIR_DESCRIBE": "", "WEBOS_DISTRO_TOPDIR_REVISION": "", "WEBOS_ENCRYPTION_KEY_TYPE": "", "meta": "", "meta-qt5": "", "meta-starfish-product": ""}
		for err == nil {
			line, isPrefix, err = buildLogReader.ReadLine()
			eachLine := string(line)
			eachLineSplit := strings.Split(eachLine, " ")
			r := ParseMeta(eachLine)
			r_length := len(r)
			var _ = eachLineSplit
			if _, ok := keyMap[eachLineSplit[0]]; ok {
				if r_length == 3 {
					buildInfo[r[0]] = r[2]
				}
			}
			if r_length > 5 && r[0] == "NOTE:" && r[1] == "do_populate_lic:" && r[3] == "sstate" && r[4] == "reuse" {
				buildInfo["num_of_from_scratch"] = eachLineSplit[7]
				continue
			}
			if r_length > 18 && r[0] == "TIME:" && r[12] == "rsync" && r[13] == "-arz" && r[17] == "BUILD-ARTIFACTS/build_changes.log" {
				buildInfo["time_rsync_artifacts"] = eachLineSplit[2]
				continue
			}
			//if r_length > 18 && r[0] == "TIME:" && r[12] == "sh" && r[13] == "-c" && r[17] == "--targets='" {
			if r_length > 18 && r[0] == "TIME:" && r[12] == "sh" && r[13] == "-c" {
				buildInfo["time_build_sh"] = eachLineSplit[2]
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
	return fmt.Sprintf("%s,%s,%s,%s", buildJobName, buildNumber, buildInfo["time_build_sh"], xmlEntity)
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
	done := make(chan int, *nThread)
	//fmt.Println("Job Name, Build Number, Result, Host, Duration, Start, Gerrit Received, Time Diff")

	// Create goroutines that handle each build's log and build.xml files
	for j := 0; j < *nThread; j++ {
		go func(j int) {
			session, err := mgo.Dial("156.147.69.55:27017")
			if err != nil {
				panic(err)
			}
			defer session.Close()
			db := session.DB("git_api_server")
			db.Login("log_manager", "Sanfrancisco")
			coll := db.C("verifyjob")
			index := mgo.Index{
				Key:        []string{"buildjob", "data"},
				Unique:     true,
				DropDups:   true,
				Background: true,
				Sparse:     true,
			}
			err = coll.EnsureIndex(index)
			if err != nil {
				panic(err)
			}
			for buildJob := range buildjobs {
				isExist := CheckBuildExistInDB(buildJob, coll)
				if isExist {
					var _ = err
				} else {
					s := AnalyzeBuild(buildJob)
					s_ele := strings.Split(s, ",")
					i_jobname := s_ele[0]
					arr := strings.Split(i_jobname, "-")
					i_machine := arr[3]
					i_buildnumber, _ := strconv.Atoi(s_ele[1])
					i_buildsh, _ := strconv.ParseFloat(s_ele[2], 64)
					i_duration, _ := strconv.ParseFloat(s_ele[5], 64)
					i_start, _ := strconv.ParseFloat(s_ele[6], 64)
					i_received, _ := strconv.ParseFloat(s_ele[7], 64)
					i_timediff, _ := strconv.ParseFloat(s_ele[8], 64)
					i_project := s_ele[9]
					i_branch := s_ele[10]
					i_number, _ := strconv.Atoi(s_ele[11])
					i_url := s_ele[12]
					coll.Remove(bson.M{"jobname": s_ele[0], "$and": []interface{}{
						bson.M{"buildnumber": i_buildnumber},
					}})
					coll.Insert(&struct {
						ID             bson.ObjectId `bson:"_id,omitempty"`
						Jobname        string
						Buildnumber    int
						Result         string
						Host           string
						Duration       float64
						Start          float64
						Gerritreceived float64
						Timediff       float64
						Build_sh       float64
						Project        string
						Branch         string
						Number         int
						Url            string
						Machine        string
					}{
						Jobname:        i_jobname,
						Buildnumber:    i_buildnumber,
						Result:         s_ele[3],
						Host:           s_ele[4],
						Duration:       i_duration,
						Start:          i_start,
						Gerritreceived: i_received,
						Timediff:       i_timediff,
						Build_sh:       i_buildsh,
						Project:        i_project,
						Branch:         i_branch,
						Number:         i_number,
						Url:            i_url,
						Machine:        i_machine,
					})
				}
			}
			done <- 1
		}(j)
	}

	// Create a goroutine that find builds that has build.xml and log file,
	// If they exist, that build's location is sent to 'listChannel'
	// and other routines will handle it
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
				} else {
					log.Println("INFO: " + build.Name() + " can't be added ")
				}
			}
		}
		log.Println("END: Getting build directories")
		close(buildjobs)
	}()

	// Check If all log analyzer routines are completed
	for m := 0; m < *nThread; m++ {
		<-done
	}
	log.Println("END:")
}
