package oebuildjobs

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
)

func ExampleTest() {
	v := Build{}

	file, err := os.Open("/Users/sunjoo/temp/jenkins_home/jobs/starfish-drd4tv-verify-h15/builds/10001/build.xml")
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
	fmt.Println(v.StartTime)
	//fmt.Println(v.Duration)
	fmt.Println(v.Result)
	//fmt.Println(v.Host)
	/*
		for _, each_parameter := range v.Parameters {
			fmt.Println(each_parameter.Name, each_parameter.Value)
		}
		fmt.Println(v.Description)
	*/
	// Output:
	// 1463613210268
	// SUCCESS
}
