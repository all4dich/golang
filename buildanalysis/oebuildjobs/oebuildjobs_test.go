package oebuildjobs

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
)

func getTagInfo(content []byte) (tag, value string) {
	type TagDef struct {
		XMLName xml.Name
		Content string `xml:",innerxml"`
	}
	v := TagDef{}
	r := xml.Unmarshal(content, &v)
	var _ = r
	return v.XMLName.Local, v.Content
}

func ExampleVerifyBuildXml() {
	v := VerifyBuild{}
	//file, err := os.Open("/Users/sunjoo/temp/jenkins_home/jobs/starfish-drd4tv-verify-h15/builds/10001/build.xml")
	file, err := os.Open("build.xml")
	if err != nil {
		fmt.Println("Error: Cannot open a file")
		os.Exit(1)
	}
	r := bufio.NewReader(file)
	dat, err := ioutil.ReadAll(r)
	if err != nil {
		fmt.Println("Error: Cannot read a string")
		os.Exit(1)
	}
	err = xml.Unmarshal(dat, &v)
	var _ = err
	fmt.Println(string(v.Result))
	fmt.Println(string(v.BuildEvent.Provider.Host))
	fmt.Println(v.BuildEvent.GerritChange.Number)
	fmt.Println(v.BuildEvent.GerritChange.Id)
	// Output:
	// SUCCESS
	// wall.lge.com
	// 97589
	// I693e80759e98f7ac1a57c78a41cc7e4ae2fb78c7
}

func ExampleOfficialBuild() {
	v := OfficialBuild{}
	file, err := os.Open("/Users/sunjoo/temp/jenkins_home/jobs/starfish-drd4tv-official-h15/builds/lastSuccessfulBuild/build.xml")
	if err != nil {
		fmt.Println("Error: Cannot open a file")
		os.Exit(1)
	}
	r := bufio.NewReader(file)
	dat, err := ioutil.ReadAll(r)
	if err != nil {
		fmt.Println("Error: Cannot read a string")
		os.Exit(1)
	}
	err = xml.Unmarshal(dat, &v)
	content := v.Causes.Causes[0].Content
	tagname := v.Causes.Causes[0].XMLName
	_, c := getTagInfo(content)
	fmt.Println(c)
	fmt.Println(tagname.Local)
	// Output:
	// 127.0.0.1

}
