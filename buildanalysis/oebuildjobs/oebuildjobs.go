package oebuildjobs

import "encoding/xml"
import "fmt"

type Account struct {
	Name     string `xml:"name"`
	Email    string `xml:"email"`
	Username string `xml:"username"`
}

type Provider struct {
	Name    string `xml:"name"`
	Host    string `xml:"host"`
	Port    int    `xml:"port"`
	Scheme  string `xml:"scheme"`
	Url     string `xml:"url"`
	Version string `xml:"version"`
}

type GerritChange struct {
	Project       string  `xml:"project"`
	Branch        string  `xml:"branch"`
	Id            string  `xml:"id"`
	Number        int     `xml:"number"`
	Subject       string  `xml:"subject"`
	CommitMessage string  `xml:"commitMessage"`
	Owner         Account `xml:"owner"`
	Url           string  `xml:"url"`
}

type PatchSet struct {
	Number    int     `xml:"number"`
	Revision  string  `xml:"revision"`
	Ref       string  `xml:"ref"`
	Draft     bool    `xml:"draft"`
	Uploader  Account `xml:"uploader"`
	Author    Account `xml:"author"`
	Parents   string  `xml:"parents>string"`
	CreatedOn string  `xml:"createdOn"`
}

type Files []string

type Approval struct {
	Type  string `xml:"type"`
	Value string `xml:"value"`
}

type TEvent struct {
	Provider     Provider     `xml:"provider"`
	Account      Account      `xml:"account"`
	GerritChange GerritChange `xml:"change"`
	PatchSet     PatchSet     `xml:"patchSet"`
	Files        Files        `xml:"files>string"`
	Comment      string       `xml:"comment"`
	ReceivedOn   int64        `xml:"receivedOn"`
	Approvals    []Approval   `xml:"approvals>com.sonyericsson.hudson.plugins.gerrit.gerritevents.dto.attr.Approval"`
}

type TriggeredItem struct {
	BuildNumber int    `xml:"buildNumber"`
	ProjectId   string `xml:"projectId"`
}

type RetriggerContext struct {
	ThisBuild TriggeredItem   `xml:"thisBuild"`
	Others    []TriggeredItem `xml:"others>triggeredItemEntity"`
}

type RetriggerEvent struct {
	Context RetriggerContext `xml:"context"`
}

type EachParameter struct {
	Name  string `xml:"name"`
	Value string `xml:"value"`
}

type commonBuildAttr struct {
	Description string `xml:"description"`
	Duration    int64  `xml:"duration"`
	Host        string `xml:"builtOn"`
	Result      string `xml:"result"`
	Start       int64  `xml:"startTime"`
	Workspace   string `xml:"workspace"`
}

type GerritChangeInfo struct {
	Project      string   `xml:"change>project"`
	Branch       string   `xml:"change>branch"`
	Changenumber int      `xml:"change>number"`
	Url          string   `xml:"change>url"`
	Changeid     string   `xml:"change>id"`
	Patchset     PatchSet `xml:"patchSet"`
	ReceivedOn   int64    `xml:"receivedOn"`
}

type VerifyBuild struct {
	XMLName xml.Name `xml:"build"`
	xmlroot string   "actions"
	commonBuildAttr
	GerritChangeInfo GerritChangeInfo `xml:"actions>com.sonyericsson.hudson.plugins.gerrit.trigger.hudsontrigger.BadgeAction>tEvent"`
	GitChangeInfo    GitChangeInfo    `xml:"actions>hudson.plugins.git.util.BuildData"`
	RetriggerInfo    RetriggerEvent   `xml:"actions>com.sonyericsson.hudson.plugins.gerrit.trigger.hudsontrigger.actions.RetriggerAction"`
	Parameters       []EachParameter  `xml:"actions>hudson.model.ParametersAction>parameters>hudson.model.StringParameterValue"`
}

type CauseAction struct {
	XMLName xml.Name
	Content []byte `xml:",innerxml"`
}

type Causes struct {
	Causes []CauseAction `xml:",any"`
}

type GitChangeInfo struct {
	Branch        string `xml:"buildsByBranchName>entry>hudson.plugins.git.util.Build>marked>branches>hudson.plugins.git.Branch>name"`
	Commithash    string `xml:"buildsByBranchName>entry>hudson.plugins.git.util.Build>marked>sha1"`
	Buildnumber   int    `xml:"buildsByBranchName>entry>hudson.plugins.git.util.Build>hudsonBuildNumber"`
	Repositoryurl string `xml:"remoteUrls>string"`
}

type OfficialBuild struct {
	XMLName xml.Name `xml:"build"`
	commonBuildAttr
	Causes        Causes        `xml:"actions>hudson.model.CauseAction>causes"`
	GitChangeInfo GitChangeInfo `xml:"actions>hudson.plugins.git.util.BuildData"`
}

func (v VerifyBuild) String() string {
	duration := float64(v.Duration) / 1000
	startTime := float64(v.Start / 1000)
	gerritReceived := float64(v.GerritChangeInfo.ReceivedOn / 1000)
	timeDiff := startTime - gerritReceived
	// Result, buildOn, Duration, Start time, Gerrit received
	return fmt.Sprintf("%s,%s,%.2f,%.2f,%.2f,%.2f,%s,%s,%d,%s", v.Result, v.Host, duration, startTime,
		gerritReceived, timeDiff, v.GerritChangeInfo.Project, v.GerritChangeInfo.Branch,
		v.GerritChangeInfo.Changenumber,
		v.GerritChangeInfo.Url,
	)
}
