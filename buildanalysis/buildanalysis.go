package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/all4dich/golang/buildanalysis/builddata"
	"github.com/all4dich/golang/buildanalysis/oebuildjobs"
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
	n, err := coll.Find(bson.M{"jobname": buildJobName, "$and": []interface{}{
		bson.M{"buildnumber": buildNumber},
	}}).Count()
	if err != nil {
		log.Println(err)
		return true
	}
	if n != 0 {
		log.Println("Log: Exist = ", buildDir)
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

func GetFloat(i string, buildnumber int) float64 {
	r, err := strconv.ParseFloat(i, 64)
	if err != nil {
		return 0.0
	}
	return r
}

func AnalyzeBuild(buildDir string) (v oebuildjobs.BuildInfo, b map[string]string) {
	start := time.Now()
	var _ = start
	buildLogFile := buildDir + "/log"
	buildXmlFile := buildDir + "/build.xml"
	buildEle := strings.Split(buildDir, "/")
	buildJobName := buildEle[len(buildEle)-3]
	buildNumber := buildEle[len(buildEle)-1]
	buildInfo := make(map[string]string)
	buildInfo["jobname"] = buildJobName
	buildInfo["buildnumber"] = buildNumber
	buildLog, err := os.Open(buildLogFile)
	if err == nil {
		buildLogReader := bufio.NewReader(buildLog)
		var (
			isPrefix bool  = true
			err      error = nil
			line     []byte
		)
		var _ = isPrefix
		/**
				 * keyMap: List of keys that are written on a build logs with values like belows
				 * Example:
				 *     BB_VERSION           = "1.40.0"
		     *     BUILD_SYS            = "x86_64-linux"
		     *     NATIVELSBSTRING      = "ubuntu-18.04"
		     *     TARGET_SYS           = "arm-starfish-linux-gnueabi"
		     *     MACHINE              = "k7hp"
		     *     DISTRO               = "starfish"
		     *     DISTRO_VERSION       = "6.0.0"
		     *     TUNE_FEATURES        = "arm armv7a vfp thumb neon cortexa9 webos-cortexa9"
		     *     TARGET_FPU           = "softfp"
		     *     WEBOS_DISTRO_MANUFACTURING_VERSION = "00.00.00"
		     *     WEBOS_ENCRYPTION_KEY_TYPE = "develkey"
		     *     WEBOS_DISTRO_RELEASE_CODENAME = "kisscurl-kalaupapa"
		     *     WEBOS_DISTRO_BUILD_ID = "verf.hq-1445
		*/
		keyMap := map[string]string{"BB_VERSION": "", "BUILD_SYS": "", "DATETIME": "", "DISTRO": "", "DISTRO_VERSION": "", "MACHINE": "", "NATIVELSBSTRING": "", "TARGET_FPU": "", "TARGET_SYS": "", "TUNE_FEATURES": "", "WEBOS_DISTRO_BUILD_ID": "", "WEBOS_DISTRO_MANUFACTURING_VERSION": "", "WEBOS_DISTRO_RELEA    SE_CODENAME": "", "WEBOS_DISTRO_TOPDIR_DESCRIBE": "", "WEBOS_DISTRO_TOPDIR_REVISION": "", "WEBOS_ENCRYPTION_KEY_TYPE": "", "meta": "", "meta-qt5": "", "meta-starfish-product": ""}
		for err == nil {
			line, isPrefix, err = buildLogReader.ReadLine()
			eachLine := string(line)
			eachLineSplit := strings.Split(eachLine, " ")
			r := ParseMeta(eachLine)
			r_length := len(r)
			var _ = eachLineSplit
			last_element := ""
			last_element_split := strings.Split(last_element, "/")
			last_element_size := 1

			if r_length > 2 {
				last_element = r[r_length-1]
				last_element_split = strings.Split(last_element, "/")
				last_element_size = len(last_element_split)
			}

			// Extract values for keys in 'keyMap' map object
			if _, ok := keyMap[eachLineSplit[0]]; ok {
				if r_length == 3 {
					buildInfo[r[0]] = r[2]
				} else if r[0] == "TUNE_FEATURES" { //For 'TUNE_FEATURES' fields
					temp_str := ""
					for _, v := range eachLineSplit {
						if v != "" && v != "=" && v != "TUNE_FEATURES" {
							temp_str = temp_str + " " + v
						}
					}
					buildInfo[r[0]] = temp_str
				}
			}
			if r_length > 5 && r[0] == "NOTE:" && r[1] == "do_populate_lic:" && r[3] == "sstate" && r[4] == "reuse" {
				buildInfo["num_of_from_scratch"] = eachLineSplit[7]
				continue
			}
			// Check every line's last element if it has a caprica report url
			if last_element_size > 3 && eachLineSplit[0] == "NOTE:" && last_element_split[last_element_size-3] == "Builds" && last_element_split[last_element_size-2] == "Details" {
				buildInfo["caprica"] = last_element
				continue
			}
			if r_length > 18 && r[0] == "TIME:" && r[12] == "rsync" && r[13] == "-arz" && strings.Split(r[17], "/")[0] == "BUILD-ARTIFACTS" {
				buildInfo["time_rsync_artifacts"] = eachLineSplit[2]
				continue
			}
			if r_length > 18 && r[0] == "TIME:" && r[12] == "bash" && r[13] == "-c" && r[17] == "BUILD-ARTIFACTS" {
				buildInfo["time_rsync_artifacts"] = eachLineSplit[2]
				continue
			}
			if r_length > 14 && r[0] == "TIME:" && r[12] == "rm" && r[13] == "-rf" {
				lastEle := strings.Split(r[14], "/")
				n := len(lastEle)
				if lastEle[n-1] == "BUILD" {
					buildInfo["time_rm_BUILD"] = r[2]
				} else if lastEle[n-1] == "BUILD-ARTIFACTS" {
					buildInfo["time_rm_BUILD_ARTIFACTS"] = r[2]
				} else if lastEle[n-1] == "downloads" {
					buildInfo["time_rm_downloads"] = r[2]
				} else if lastEle[n-1] == "sstate-cache" {
					buildInfo["time_rm_sstate"] = r[2]
				}
				continue
			}
			if r_length > 18 && r[0] == "TIME:" && r[12] == "sh" && r[13] == "-c" {
				buildInfo["time_build_sh"] = eachLineSplit[2]
				continue
			}
			if r_length > 18 && r[0] == "TIME:" && r[12] == "ssh" && r[13] == "-o" {
				buildInfo["time_build_sh"] = eachLineSplit[2]
				continue
			}
			if r_length > 12 && r[0] == "TIME:" && r[11] == "bitbake" {
				buildInfo["time_bitbake"] = eachLineSplit[1]
				continue
			}
		}
	}
	var _ = xml.Header
	var _ = oebuildjobs.BuildInfo{}
	var _ = buildXmlFile
	buildXml, err := os.Open(buildXmlFile)
	if err != nil {
		log.Fatal("Can't open a file " + buildXmlFile)
	}
	r := bufio.NewReader(buildXml)
	buildXmlbytes, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal("ERROR: Cannot read data from a xml file ")
	}
	buildXmlDatStr := string(buildXmlbytes)
	fileDataStrNew := strings.Replace(buildXmlDatStr, "<?xml version='1.1'", "<?xml version='1.0'", 1)
	buildXmlDat := []byte(fileDataStrNew)
	xmlEntity := oebuildjobs.BuildInfo{}
	err = xml.Unmarshal(buildXmlDat, &xmlEntity)
	v = xmlEntity
	b = buildInfo
	return v, b
}

func contains(intArray []int, number int) bool {
	for _, v := range intArray {
		if v == number {
			return true
		}
	}
	return false
}

func main() {
	jenkinsHome := flag.String("jenkinsHome", "/binary/build_results/jenkins_home_backup", "Jenkins configuration and data directory")
	jobName := flag.String("jobName", "starfish-drd4tv-official-h15", "Set a job name to parse")
	nThread := flag.Int("n", 4, "Number of threads")
	dbHost := flag.String("dbHost", "", "DB Host")
	dbPort := flag.String("dbPort", "", "DB Port")
	dbName := flag.String("dbName", "", "DB Name")
	dbColl := flag.String("dbColl", "", "DB Collection name")
	dbUser := flag.String("dbUser", "", "DB Username")
	dbPass := flag.String("dbPass", "", "DB Password")
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

	dbUrl := fmt.Sprintf("%s:%s", *dbHost, *dbPort)
	buildjobs := make(chan string, *nThread)
	done := make(chan int, *nThread)

	session_to_get_number_of_builds, err_session := mgo.Dial(dbUrl)
	if err_session != nil {
		panic(err_session)
	}
	defer session_to_get_number_of_builds.Close()
	db_main := session_to_get_number_of_builds.DB(*dbName)
	db_main.Login(*dbUser, *dbPass)
	coll_main := db_main.C(*dbColl)
	// curr_builds
	// - A list of builds that exist in a database for a job
	// - By default, it's a empty list for builddata.BuildData
	var curr_builds []builddata.BuildData
	err_coll := coll_main.Find(bson.M{"jobname": jobName}).All(&curr_builds)
	if err_coll != nil {
		panic(err_coll)
	}
	// buildNumbers
	// - A list of build numbers
	// - Filled from 'curr_builds'
	// - Used to check if a target build on build job directory has already been inserted into a database
	buildNumbers := make([]int, len(curr_builds))
	for i, v := range curr_builds {
		buildNumbers[i] = v.Buildnumber
	}

	// Create goroutines that handle each build's log and build.xml files
	for j := 0; j < *nThread; j++ {
		go func(j int) {
			session, err := mgo.Dial(dbUrl)
			if err != nil {
				panic(err)
			}
			defer session.Close()
			db := session.DB(*dbName)
			db.Login(*dbUser, *dbPass)
			coll := db.C(*dbColl)
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
			latest_build := builddata.BuildData{}
			latest_build_number := 0
			err_latest_build := coll.Find(bson.M{"jobname": jobName}).Sort("-buildnumber").One(&latest_build)
			if err_latest_build != nil {
				latest_build_number = 0
			} else {
				latest_build_number = latest_build.Buildnumber
			}
			log.Println("INFO: Lastest Build Number = ", latest_build_number)
			for buildJob := range buildjobs {
				//isExist := CheckBuildExistInDB(buildJob, coll)
				buildEle := strings.Split(buildJob, "/")
				buildNumber, _ := strconv.Atoi(buildEle[len(buildEle)-1])
				//if buildNumber <= latest_build_number {
				if contains(buildNumbers, buildNumber) {
					var _ = err
					log.Println("INFO: Already exist on database: ", buildJob)
				} else {
					log.Println("INFO: Get information from ", buildJob)
					v, b := AnalyzeBuild(buildJob)
					i_jobname := b["jobname"]
					arr := strings.Split(i_jobname, "-")
					i_machine := arr[3]
					i_buildnumber, _ := strconv.Atoi(b["buildnumber"])
					i_duration := v.Duration / 1000
					i_start := v.Start / 1000
					i_waiting_in_queue := v.WaitingTime / 1000
					i_duration_in_queue := float64(v.WaitingTime) / 1000
					i_timediff := 0
					if v.GerritChangeInfo.ReceivedOn != 0 {
						i_timediff = i_start - (v.GerritChangeInfo.ReceivedOn / 1000)
					}
					coll.Remove(bson.M{"jobname": i_jobname, "$and": []interface{}{
						bson.M{"buildnumber": i_buildnumber},
					}})
					i_parameters := bson.M{}
					for _, eachParameter := range v.Parameters {
						i_parameters[eachParameter.Name] = eachParameter.Value
					}
					i_parameters["BB_VERSION"] = strings.Replace(b["BB_VERSION"], "\"", "", -1)
					i_parameters["BUILD_SYS"] = strings.Replace(b["BUILD_SYS"], "\"", "", -1)
					i_parameters["NATIVELSBSTRING"] = strings.Replace(b["NATIVELSBSTRING"], "\"", "", -1)
					i_parameters["TARGET_SYS"] = strings.Replace(b["TARGET_SYS"], "\"", "", -1)
					i_parameters["DISTRO"] = strings.Replace(b["DISTRO"], "\"", "", -1)
					i_parameters["DISTRO_VERSION"] = strings.Replace(b["DISTRO_VERSION"], "\"", "", -1)
					i_parameters["TUNE_FEATURES"] = strings.Replace(b["TUNE_FEATURES"], "\"", "", -1)
					i_parameters["TARGET_FPU"] = strings.Replace(b["TARGET_FPU"], "\"", "", -1)
					i_parameters["WEBOS_DISTRO_MANUFACTURING_VERSION"] = strings.Replace(b["WEBOS_DISTRO_MANUFACTURING_VERSION"], "\"", "", -1)
					i_parameters["WEBOS_ENCRYPTION_KEY_TYPE"] = strings.Replace(b["WEBOS_ENCRYPTION_KEY_TYPE"], "\"", "", -1)
					i_parameters["WEBOS_DISTRO_RELEASE_CODENAME"] = strings.Replace(b["WEBOS_DISTRO_RELEASE_CODENAME"], "\"", "", -1)
					i_parameters["WEBOS_DISTRO_BUILD_ID"] = strings.Replace(b["WEBOS_DISTRO_BUILD_ID"], "\"", "", -1)
					i_parameters["WEBOS_DISTRO_TOPDIR_REVISION"] = strings.Replace(b["WEBOS_DISTRO_TOPDIR_REVISION"], "\"", "", -1)
					i_parameters["WEBOS_DISTRO_TOPDIR_DESCRIBE"] = strings.Replace(b["WEBOS_DISTRO_TOPDIR_DESCRIBE"], "\"", "", -1)
					i_parameters["caprica"] = b["caprica"]

					coll.Insert(&builddata.BuildData{
						Jobname:           i_jobname,
						Buildnumber:       i_buildnumber,
						Result:            v.Result,
						Host:              v.Host,
						Duration:          i_duration,
						Duration_in_queue: i_duration_in_queue,
						Start:             i_start,
						Waiting:           i_waiting_in_queue,
						Workspace:         v.Workspace,
						Description:       v.Description,
						Timediff:          i_timediff,
						Machine:           i_machine,
						Parameters:        i_parameters,
						Cause: bson.M{
							"parent_project":     v.Causes.Parent_project,
							"parent_user":        v.Causes.Parent_user,
							"parent_buildnumber": v.Causes.Parent_buildnumber,
							"parent_url":         v.Causes.Parent_url,
							"userid":             v.Causes.Userid,
							"retriggeredby":      v.Causes.Retriggeredby,
						},
						GerritChangeInfo: bson.M{
							"project":      v.GerritChangeInfo.Project,
							"branch":       v.GerritChangeInfo.Branch,
							"changenumber": v.GerritChangeInfo.Changenumber,
							"changeid":     v.GerritChangeInfo.Changeid,
							"url":          v.GerritChangeInfo.Url,
							"receivedon":   int(v.GerritChangeInfo.ReceivedOn / 1000),
							"patchset": bson.M{
								"number":    v.GerritChangeInfo.Patchset.Number,
								"ref":       v.GerritChangeInfo.Patchset.Ref,
								"parents":   v.GerritChangeInfo.Patchset.Parents,
								"createdon": v.GerritChangeInfo.Patchset.CreatedOn,
								"author": bson.M{
									"name":  v.GerritChangeInfo.Patchset.Author.Name,
									"email": v.GerritChangeInfo.Patchset.Author.Email,
								},
								"uploader": bson.M{
									"name":  v.GerritChangeInfo.Patchset.Uploader.Name,
									"email": v.GerritChangeInfo.Patchset.Uploader.Email,
								},
							},
						},
						GitChangeInfo: bson.M{
							"branch":        v.GitChangeInfo.Branch,
							"commithash":    v.GitChangeInfo.Commithash,
							"buildnumber":   v.GitChangeInfo.Buildnumber,
							"repositoryurl": v.GitChangeInfo.Repositoryurl,
						},
						Time_build_sh:           GetFloat(b["time_build_sh"], i_buildnumber),
						Time_bitbake:            GetFloat(b["time_bitbake"], i_buildnumber),
						Time_rm_BUILD:           GetFloat(b["time_rm_BUILD"], i_buildnumber),
						Time_rm_BUILD_ARTIFACTS: GetFloat(b["time_rm_BUILD_ARTIFACTS"], i_buildnumber),
						Time_rm_downloads:       GetFloat(b["time_rm_downloads"], i_buildnumber),
						Time_rm_sstate:          GetFloat(b["time_rm_sstate"], i_buildnumber),
						Time_rsync_artifacts:    GetFloat(b["time_rsync_artifacts"], i_buildnumber),
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
